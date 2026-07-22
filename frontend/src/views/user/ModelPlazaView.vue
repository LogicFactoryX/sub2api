<template>
  <AppLayout>
    <div class="mx-auto max-w-[1600px] space-y-5">
      <header class="flex flex-col gap-3 border-b border-gray-200 pb-5 dark:border-dark-700 md:flex-row md:items-end md:justify-between">
        <div>
          <div class="mb-2 flex items-center gap-2 text-sm font-medium text-primary-600 dark:text-primary-400">
            <Icon name="sparkles" size="sm" />
            <span>{{ t('modelPlaza.total', { count: entries.length }) }}</span>
          </div>
          <h1 class="text-2xl font-semibold text-gray-950 dark:text-white">{{ t('modelPlaza.title') }}</h1>
          <p class="mt-1 max-w-2xl text-sm text-gray-600 dark:text-gray-400">{{ t('modelPlaza.description') }}</p>
        </div>
        <button type="button" class="btn btn-secondary self-start" :disabled="loading" :title="t('common.refresh')" @click="loadEntries">
          <Icon name="refresh" size="md" :class="loading ? 'animate-spin' : ''" />
        </button>
      </header>

      <div class="flex flex-col gap-4 rounded-lg border border-gray-200 bg-white p-3 dark:border-dark-700 dark:bg-dark-900 lg:flex-row lg:items-center">
        <div class="relative min-w-0 flex-1">
          <Icon name="search" size="md" class="absolute left-3 top-1/2 -translate-y-1/2 text-gray-400" />
          <input v-model.trim="search" type="search" class="input pl-10" :placeholder="t('modelPlaza.searchPlaceholder')" />
        </div>
        <div class="flex flex-col gap-3 sm:flex-row">
          <Select v-model="platform" class="w-full sm:w-44" :options="platformOptions" />
          <Select v-model="groupID" class="w-full sm:w-56" :options="groupOptions" />
        </div>
      </div>

      <div v-if="loading" class="grid gap-4 md:grid-cols-2 xl:grid-cols-3">
        <div v-for="index in 6" :key="index" class="h-64 animate-pulse rounded-lg border border-gray-200 bg-white dark:border-dark-700 dark:bg-dark-900"></div>
      </div>

      <div v-else-if="filteredEntries.length" class="grid gap-4 md:grid-cols-2 xl:grid-cols-3">
        <article
          v-for="entry in filteredEntries"
          :key="entry.id"
          class="model-tile group flex min-h-64 flex-col border border-gray-200 bg-white p-5 transition-colors hover:border-primary-300 dark:border-dark-700 dark:bg-dark-900 dark:hover:border-primary-700"
        >
          <div class="flex items-start justify-between gap-3">
            <div class="flex min-w-0 items-start gap-3">
              <div class="flex h-10 w-10 flex-none items-center justify-center rounded-md" :class="platformIconClass(entry.platform)">
                <Icon name="cpu" size="lg" />
              </div>
              <div class="min-w-0">
                <h2 class="truncate font-mono text-sm font-semibold text-gray-950 dark:text-white" :title="entry.model_name">
                  {{ entry.display_name || entry.model_name }}
                </h2>
                <p v-if="entry.display_name" class="mt-0.5 truncate font-mono text-xs text-gray-500 dark:text-gray-400">{{ entry.model_name }}</p>
              </div>
            </div>
            <button type="button" class="icon-button" :title="t('modelPlaza.copyModel')" @click="copyModel(entry.model_name)">
              <Icon name="copy" size="sm" />
            </button>
          </div>

          <div v-if="entry.pricing_available" class="mt-5 grid grid-cols-2 gap-x-5 gap-y-3">
            <PriceMetric :label="t('modelPlaza.input')" :value="entry.display_pricing.input_per_million ?? undefined" />
            <PriceMetric :label="t('modelPlaza.output')" :value="entry.display_pricing.output_per_million ?? undefined" />
            <PriceMetric v-if="entry.display_pricing.cache_write_per_million != null" :label="t('modelPlaza.cacheWrite')" :value="entry.display_pricing.cache_write_per_million" compact />
            <PriceMetric v-if="entry.display_pricing.cache_read_per_million != null" :label="t('modelPlaza.cacheRead')" :value="entry.display_pricing.cache_read_per_million" compact />
          </div>
          <div v-else class="mt-5 rounded-md bg-gray-50 px-3 py-2 text-sm text-gray-500 dark:bg-dark-800 dark:text-gray-400">
            {{ t('modelPlaza.pricingUnavailable') }}
          </div>

          <p class="mt-5 line-clamp-3 min-h-[3.75rem] text-sm leading-5 text-gray-600 dark:text-gray-300">
            {{ entry.description || entry.model_name }}
          </p>

          <div class="mt-auto flex flex-wrap items-center gap-2 border-t border-gray-100 pt-4 text-xs dark:border-dark-800">
            <span class="badge badge-gray">{{ platformLabel(entry.platform) }}</span>
            <span class="badge" :class="groupBadgeClass(entry)">{{ entry.group_name }}</span>
            <span class="font-mono text-gray-500 dark:text-gray-400">{{ t('modelPlaza.multiplier', { value: formatMultiplier(entry.rate_multiplier) }) }}</span>
            <span v-for="tag in entry.tags" :key="tag" class="text-gray-500 dark:text-gray-400">#{{ tag }}</span>
          </div>
        </article>
      </div>

      <div v-else class="flex min-h-72 flex-col items-center justify-center border-y border-gray-200 text-center dark:border-dark-700">
        <Icon name="cube" size="xl" class="mb-3 text-gray-400" />
        <p class="font-medium text-gray-800 dark:text-gray-200">{{ t('modelPlaza.empty') }}</p>
      </div>
    </div>
  </AppLayout>
</template>

<script setup lang="ts">
import { computed, defineComponent, h, onMounted, ref } from 'vue'
import { useI18n } from 'vue-i18n'
import AppLayout from '@/components/layout/AppLayout.vue'
import Icon from '@/components/icons/Icon.vue'
import Select, { type SelectOption } from '@/components/common/Select.vue'
import modelPlazaAPI, { type ModelPlazaEntry } from '@/api/modelPlaza'
import { useAppStore } from '@/stores/app'
import { extractApiErrorMessage } from '@/utils/apiError'

const { t } = useI18n()
const appStore = useAppStore()
const entries = ref<ModelPlazaEntry[]>([])
const loading = ref(false)
const search = ref('')
const platform = ref<string>('')
const groupID = ref<number | ''>('')

const PriceMetric = defineComponent({
  props: { label: { type: String, required: true }, value: { type: Number, default: null }, compact: Boolean },
  setup(props) {
    return () => h('div', { class: 'min-w-0' }, [
      h('div', { class: 'text-xs text-gray-500 dark:text-gray-400' }, props.label),
      h('div', { class: [props.compact ? 'text-base' : 'text-lg', 'mt-0.5 truncate font-mono font-semibold text-gray-950 dark:text-white'] },
        props.value == null ? '-' : `$${formatPrice(props.value)}`),
      h('div', { class: 'text-[11px] text-gray-400 dark:text-gray-500' }, t('modelPlaza.perMillion'))
    ])
  }
})

const platformOptions = computed<SelectOption[]>(() => [
  { value: '', label: t('modelPlaza.allPlatforms') },
  ...Array.from(new Set(entries.value.map((entry) => entry.platform))).sort().map((value) => ({ value, label: platformLabel(value) }))
])

const groupOptions = computed<SelectOption[]>(() => [
  { value: '', label: t('modelPlaza.allGroups') },
  ...Array.from(new Map(entries.value.map((entry) => [entry.group_id, entry.group_name])).entries())
    .sort((a, b) => a[1].localeCompare(b[1]))
    .map(([value, label]) => ({ value, label }))
])

const filteredEntries = computed(() => {
  const query = search.value.toLowerCase()
  return entries.value.filter((entry) => {
    if (platform.value && entry.platform !== platform.value) return false
    if (groupID.value !== '' && entry.group_id !== Number(groupID.value)) return false
    if (!query) return true
    return [entry.model_name, entry.display_name, entry.description, entry.group_name, entry.platform, ...entry.tags]
      .some((value) => value.toLowerCase().includes(query))
  })
})

async function loadEntries() {
  loading.value = true
  try {
    entries.value = await modelPlazaAPI.list()
  } catch (error: unknown) {
    appStore.showError(extractApiErrorMessage(error, t('modelPlaza.loadFailed')))
  } finally {
    loading.value = false
  }
}

async function copyModel(model: string) {
  await navigator.clipboard.writeText(model)
  appStore.showSuccess(t('modelPlaza.copied'))
}

function formatPrice(value: number): string {
  if (value === 0) return '0'
  if (value < 0.001) return value.toFixed(6).replace(/0+$/, '').replace(/\.$/, '')
  if (value < 1) return value.toFixed(4).replace(/0+$/, '').replace(/\.$/, '')
  return value.toFixed(2).replace(/\.00$/, '')
}

function formatMultiplier(value: number): string {
  return value.toFixed(4).replace(/0+$/, '').replace(/\.$/, '')
}

function platformLabel(value: string): string {
  const labels: Record<string, string> = { openai: 'OpenAI', anthropic: 'Anthropic', gemini: 'Gemini', antigravity: 'Antigravity', grok: 'Grok' }
  return labels[value] || value
}

function platformIconClass(value: string): string {
  const classes: Record<string, string> = {
    openai: 'bg-emerald-50 text-emerald-700 dark:bg-emerald-950/40 dark:text-emerald-300',
    anthropic: 'bg-amber-50 text-amber-700 dark:bg-amber-950/40 dark:text-amber-300',
    gemini: 'bg-blue-50 text-blue-700 dark:bg-blue-950/40 dark:text-blue-300',
    antigravity: 'bg-violet-50 text-violet-700 dark:bg-violet-950/40 dark:text-violet-300',
    grok: 'bg-gray-100 text-gray-800 dark:bg-gray-800 dark:text-gray-200'
  }
  return classes[value] || classes.grok
}

function groupBadgeClass(entry: ModelPlazaEntry): string {
  if (entry.is_member_group) return 'badge-purple'
  if (entry.is_exclusive) return 'badge-warning'
  return 'badge-success'
}

onMounted(loadEntries)
</script>

<style scoped>
.model-tile { border-radius: 8px; }
.icon-button {
  display: inline-flex;
  height: 2rem;
  width: 2rem;
  flex: none;
  align-items: center;
  justify-content: center;
  border-radius: 6px;
  color: rgb(107 114 128);
  transition: background-color 150ms, color 150ms;
}
.icon-button:hover { background: rgb(243 244 246); color: rgb(17 24 39); }
.dark .icon-button:hover { background: rgb(31 41 55); color: rgb(243 244 246); }
</style>
