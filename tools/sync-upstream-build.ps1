[CmdletBinding()]
param(
    [string]$OutputDir = "E:\sub2api-build",
    [string]$CustomBranch = "custom/group-redemption",
    [switch]$ForceBuild,
    [switch]$NoPush
)

$ErrorActionPreference = "Stop"
Set-StrictMode -Version Latest

function Invoke-Native {
    param(
        [Parameter(Mandatory = $true)][string]$FilePath,
        [Parameter(ValueFromRemainingArguments = $true)][string[]]$Arguments
    )

    & $FilePath @Arguments
    if ($LASTEXITCODE -ne 0) {
        throw "Command failed ($LASTEXITCODE): $FilePath $($Arguments -join ' ')"
    }
}

function Resolve-Tool {
    param([string]$Name, [string[]]$Fallbacks = @())

    $command = Get-Command $Name -ErrorAction SilentlyContinue
    if ($null -ne $command) {
        return $command.Source
    }
    foreach ($candidate in $Fallbacks) {
        if (Test-Path -LiteralPath $candidate) {
            return $candidate
        }
    }
    throw "Required tool was not found: $Name"
}

$repoRoot = (Resolve-Path (Join-Path $PSScriptRoot "..")).Path
$git = Resolve-Tool "git" @("E:\runtime\git\cmd\git.exe")
$go = Resolve-Tool "go" @("E:\runtime\go\bin\go.exe")
$corepack = Resolve-Tool "corepack" @("E:\runtime\corepack.cmd")
$originalLocation = Get-Location

try {
    Set-Location $repoRoot

    if (@(& $git status --porcelain).Count -ne 0) {
        throw "The worktree is not clean. Commit or stash changes before syncing."
    }

    $originUrl = (& $git remote get-url origin).Trim()
    $upstreamUrl = (& $git remote get-url upstream).Trim()
    if (-not $originUrl -or -not $upstreamUrl) {
        throw "Both origin (your fork) and upstream (official repository) are required."
    }

    Write-Host "Fetching official and fork branches..."
    Invoke-Native $git fetch upstream main
    Invoke-Native $git fetch origin main

    $localMain = (& $git rev-parse main).Trim()
    $upstreamMain = (& $git rev-parse upstream/main).Trim()
    $hasUpdate = $localMain -ne $upstreamMain

    if (-not $hasUpdate -and -not $ForceBuild) {
        Write-Host "No official main update was found. Use -ForceBuild to rebuild anyway."
        exit 0
    }

    if ($hasUpdate) {
        Write-Host "Updating fork main from official main..."
        Invoke-Native $git switch main
        Invoke-Native $git merge --ff-only upstream/main
        if (-not $NoPush) {
            Invoke-Native $git push origin main
        }

        Write-Host "Merging main into $CustomBranch..."
        Invoke-Native $git switch $CustomBranch
        try {
            Invoke-Native $git merge --no-edit main
        }
        catch {
            Write-Host "Merge conflicts need manual resolution. The custom branch was not built or pushed." -ForegroundColor Yellow
            & $git status --short
            throw
        }
    }
    else {
        Invoke-Native $git switch $CustomBranch
    }

    Write-Host "Installing and checking frontend dependencies..."
    Push-Location (Join-Path $repoRoot "frontend")
    try {
        $oldCi = $env:CI
        try {
            $env:CI = "true"
            Invoke-Native $corepack pnpm@9 install --frozen-lockfile
            Invoke-Native $corepack pnpm@9 run typecheck
            Invoke-Native $corepack pnpm@9 run build
        }
        finally {
            $env:CI = $oldCi
        }
    }
    finally {
        Pop-Location
    }

    Write-Host "Running backend tests..."
    Push-Location (Join-Path $repoRoot "backend")
    try {
        Invoke-Native $go test ./...

        New-Item -ItemType Directory -Force -Path $OutputDir | Out-Null
        $baseVersion = (Get-Content -LiteralPath "cmd\server\VERSION" -Raw).Trim()
        $version = "$baseVersion-group-redemption.$(Get-Date -Format 'yyyyMMdd')"
        $commit = (& $git rev-parse --short=12 HEAD).Trim()
        $buildDate = (Get-Date).ToUniversalTime().ToString("yyyy-MM-ddTHH:mm:ssZ")
        $exePath = Join-Path $OutputDir "sub2api.exe"
        $ldflags = "-s -w -X main.Version=$version -X main.Commit=$commit -X main.Date=$buildDate"

        Write-Host "Building $exePath..."
        $oldCgo = $env:CGO_ENABLED
        $oldGoos = $env:GOOS
        $oldGoarch = $env:GOARCH
        try {
            $env:CGO_ENABLED = "0"
            $env:GOOS = "windows"
            $env:GOARCH = "amd64"
            Invoke-Native $go build -tags "embed,timetzdata" -trimpath -ldflags $ldflags -o $exePath ./cmd/server
        }
        finally {
            $env:CGO_ENABLED = $oldCgo
            $env:GOOS = $oldGoos
            $env:GOARCH = $oldGoarch
        }
    }
    finally {
        Pop-Location
    }

    $hash = (Get-FileHash -LiteralPath $exePath -Algorithm SHA256).Hash.ToLowerInvariant()
    Set-Content -LiteralPath "$exePath.sha256" -Value "$hash  sub2api.exe" -Encoding ASCII
    $zipPath = Join-Path $OutputDir "sub2api-$version-windows-amd64.zip"
    if (Test-Path -LiteralPath $zipPath) {
        Remove-Item -LiteralPath $zipPath -Force
    }
    Compress-Archive -LiteralPath $exePath, "$exePath.sha256" -DestinationPath $zipPath

    if (-not $NoPush) {
        Invoke-Native $git push origin $CustomBranch
    }

    Write-Host "Build complete: $exePath" -ForegroundColor Green
    Write-Host "SHA256: $hash"
    Write-Host "Archive: $zipPath"
}
finally {
    Set-Location $originalLocation
}
