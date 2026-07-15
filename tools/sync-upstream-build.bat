@echo off
setlocal
powershell.exe -NoProfile -ExecutionPolicy Bypass -File "%~dp0sync-upstream-build.ps1" %*
exit /b %errorlevel%
