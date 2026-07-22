<template>
  <AppLayout>
    <div class="mx-auto max-w-[1500px] space-y-5">
      <header class="flex flex-col gap-4 border-b border-gray-200 pb-5 dark:border-dark-700 md:flex-row md:items-end md:justify-between">
        <div>
          <h1 class="text-2xl font-semibold text-gray-950 dark:text-white">{{ t('admin.modelPlaza.title') }}</h1>
          <p class="mt-1 text-sm text-gray-600 dark:text-gray-400">{{ t('admin.modelPlaza.description') }}</p>
        </div>
        <button type="button" class="btn btn-primary self-start" @click="openCreate">
          <Icon name="plus" size="md" class="mr-2" />
          {{ t('admin.modelPlaza.create') }}
        </button>
      </header>

      <div class="flex flex-col gap-3 rounded-lg border border-gray-200 bg-white p-3 dark:border-dark-700 dark:bg-dark-900 sm:flex-row sm:items-center">
        <div class="relative flex-1">
          <Icon name="search" size="md" class="absolute left-3 top-1/2 -translate-y-1/2 text-gray-400" />
          <input v-model.trim="search" class="input pl-10" type="search" :placeholder="t('modelPlaza.searchPlaceholder')" />
        </div>
        <button type="button" class="btn btn-secondary" :disabled="loading" :title="t('common.refresh')" @click="loadEntries">
          <Icon name="refresh" size="md" :class="loading ? 'animate-spin' : ''" />
        </button>
      </div>

      <div class="overflow-hidden rounded-lg border border-gray-200 bg-white dark:border-dark-700 dark:bg-dark-900">
        <div class="overflow-x-auto">
          <table class="min-w-[720px] divide-y divide-gray-200 text-sm dark:divide-dark-700">
            <thead class="bg-gray-50 dark:bg-dark-800">
              <tr class="text-left text-xs font-medium uppercase text-gray-500 dark:text-gray-400">
                <th class="px-4 py-3">{{ t('admin.modelPlaza.model') }}</th>
                <th class="px-4 py-3">{{ t('admin.modelPlaza.group') }}</th>
                <th class="px-4 py-3">{{ t('admin.modelPlaza.pricing') }}</th>
                <th class="px-4 py-3">{{ t('admin.modelPlaza.sortOrder') }}</th>
                <th class="px-4 py-3">{{ t('admin.modelPlaza.enabled') }}</th>
                <th class="px-4 py-3 text-right">{{ t('admin.modelPlaza.actions') }}</th>
              </tr>
            </thead>
            <tbody class="divide-y divide-gray-100 dark:divide-dark-800">
              <tr v-for="entry in filteredEntries" :key="entry.id" class="hover:bg-gray-50/70 dark:hover:bg-dark-800/60">
                <td class="px-4 py-3">
                  <div class="font-medium text-gray-950 dark:text-white">{{ entry.display_name || entry.model_name }}</div>
                  <div class="mt-0.5 font-mono text-xs text-gray-500">{{ entry.model_name }}</div>
                </td>
                <td class="px-4 py-3">
                  <div class="font-medium text-gray-800 dark:text-gray-200">{{ entry.group_name }}</div>
                  <div class="text-xs text-gray-500">{{ platformLabel(entry.platform) }} · {{ formatRate(entry.rate_multiplier) }}x</div>
                </td>
                <td class="whitespace-nowrap px-4 py-3 font-mono text-xs text-gray-700 dark:text-gray-300">
                  <template v-if="entry.pricing_available">
                    <div>{{ t('modelPlaza.input') }} ${{ formatPrice(entry.display_pricing.input_per_million) }}</div>
                    <div class="mt-1">{{ t('modelPlaza.output') }} ${{ formatPrice(entry.display_pricing.output_per_million) }}</div>
                  </template>
                  <span v-else class="text-gray-400">{{ t('modelPlaza.pricingUnavailable') }}</span>
                </td>
                <td class="px-4 py-3 font-mono text-gray-600 dark:text-gray-300">{{ entry.sort_order }}</td>
                <td class="px-4 py-3"><Toggle :model-value="entry.enabled" @update:model-value="toggleEntry(entry, $event)" /></td>
                <td class="px-4 py-3">
                  <div class="flex justify-end gap-1">
                    <button type="button" class="table-action" :title="t('common.edit')" @click="openEdit(entry)"><Icon name="edit" size="sm" /></button>
                    <button type="button" class="table-action text-red-600 hover:bg-red-50 dark:text-red-400 dark:hover:bg-red-950/30" :title="t('common.delete')" @click="askDelete(entry)"><Icon name="trash" size="sm" /></button>
                  </div>
                </td>
              </tr>
            </tbody>
          </table>
        </div>
        <div v-if="loading" class="p-10 text-center text-sm text-gray-500">{{ t('common.loading') }}</div>
        <div v-else-if="!filteredEntries.length" class="p-12 text-center text-sm text-gray-500">{{ t('admin.modelPlaza.noEntries') }}</div>
      </div>
    </div>

    <BaseDialog :show="showDialog" :title="editing ? t('admin.modelPlaza.edit') : t('admin.modelPlaza.create')" width="wide" @close="closeDialog">
      <form class="space-y-5" @submit.prevent="saveEntry">
        <div class="grid gap-4 md:grid-cols-2">
          <div>
            <label class="input-label">{{ t('admin.modelPlaza.group') }}</label>
            <Select v-model="form.group_id" :options="groupOptions" searchable :placeholder="t('admin.modelPlaza.selectGroupFirst')" @change="onGroupChange" />
          </div>
          <div>
            <label class="input-label">{{ t('admin.modelPlaza.model') }}</label>
            <Select v-model="form.model_name" :options="modelOptions" searchable creatable :disabled="!form.group_id || modelsLoading" :placeholder="t('admin.modelPlaza.selectModel')" @change="onModelChange" />
          </div>
          <div>
            <label class="input-label">{{ t('admin.modelPlaza.displayName') }}</label>
            <input v-model.trim="form.display_name" class="input" maxlength="255" :placeholder="t('admin.modelPlaza.displayNamePlaceholder')" />
          </div>
          <div>
            <label class="input-label">{{ t('admin.modelPlaza.tags') }}</label>
            <input v-model="tagsText" class="input" :placeholder="t('admin.modelPlaza.tagsPlaceholder')" />
          </div>
        </div>
        <div>
          <label class="input-label">{{ t('admin.modelPlaza.descriptionLabel') }}</label>
          <textarea v-model.trim="form.description" class="input min-h-24 resize-y" maxlength="2000" :placeholder="t('admin.modelPlaza.descriptionPlaceholder')"></textarea>
        </div>
        <div class="grid gap-4 md:grid-cols-[160px_1fr]">
          <div>
            <label class="input-label">{{ t('admin.modelPlaza.sortOrder') }}</label>
            <input v-model.number="form.sort_order" type="number" class="input" />
          </div>
          <div class="flex items-end pb-2">
            <label class="flex cursor-pointer items-center gap-3 text-sm font-medium text-gray-700 dark:text-gray-300">
              <Toggle v-model="form.enabled" />
              {{ t('admin.modelPlaza.enabled') }}
            </label>
          </div>
        </div>

        <div class="border-y border-gray-200 py-4 dark:border-dark-700">
          <div class="flex flex-col gap-3 sm:flex-row sm:items-center sm:justify-between">
            <div>
              <div class="text-sm font-medium text-gray-900 dark:text-white">{{ t('admin.modelPlaza.pricing') }}</div>
              <div class="mt-0.5 text-xs text-gray-500">{{ t('admin.modelPlaza.formula') }}</div>
            </div>
            <div v-if="preview" class="flex gap-6 font-mono text-sm">
              <span>{{ t('modelPlaza.input') }} <strong>${{ formatPrice(preview.input) }}</strong></span>
              <span>{{ t('modelPlaza.output') }} <strong>${{ formatPrice(preview.output) }}</strong></span>
            </div>
            <div v-else class="text-sm text-gray-400">{{ t('modelPlaza.pricingUnavailable') }}</div>
          </div>
        </div>
      </form>
      <template #footer>
        <button type="button" class="btn btn-secondary" @click="closeDialog">{{ t('common.cancel') }}</button>
        <button type="button" class="btn btn-primary" :disabled="saving || !form.group_id || !form.model_name" @click="saveEntry">
          {{ saving ? t('common.saving') : t('common.save') }}
        </button>
      </template>
    </BaseDialog>

    <ConfirmDialog
      :show="Boolean(deleting)"
      :title="t('common.delete')"
      :message="t('admin.modelPlaza.deleteConfirm', { name: deleting?.display_name || deleting?.model_name || '' })"
      :confirm-text="t('common.delete')"
      :cancel-text="t('common.cancel')"
      danger
      @confirm="deleteEntry"
      @cancel="deleting = null"
    />
  </AppLayout>
</template>

<script setup lang="ts">
import { computed, onMounted, reactive, ref } from 'vue'
import { useI18n } from 'vue-i18n'
import AppLayout from '@/components/layout/AppLayout.vue'
import BaseDialog from '@/components/common/BaseDialog.vue'
import ConfirmDialog from '@/components/common/ConfirmDialog.vue'
import Icon from '@/components/icons/Icon.vue'
import Select, { type SelectOption } from '@/components/common/Select.vue'
import Toggle from '@/components/common/Toggle.vue'
import { adminAPI } from '@/api/admin'
import type { ModelPlazaEntry } from '@/api/modelPlaza'
import type { ModelPlazaEntryInput } from '@/api/admin/modelPlaza'
import type { AdminGroup } from '@/types'
import { useAppStore } from '@/stores/app'
import { extractApiErrorMessage } from '@/utils/apiError'

const { t } = useI18n()
const appStore = useAppStore()
const entries = ref<ModelPlazaEntry[]>([])
const groups = ref<AdminGroup[]>([])
const catalogModels = ref<string[]>([])
const providerPricing = ref<{ input_price?: number; output_price?: number } | null>(null)
const search = ref('')
const loading = ref(false)
const modelsLoading = ref(false)
const saving = ref(false)
const showDialog = ref(false)
const editing = ref<ModelPlazaEntry | null>(null)
const deleting = ref<ModelPlazaEntry | null>(null)
const tagsText = ref('')
const form = reactive<ModelPlazaEntryInput>({ model_name: '', display_name: '', description: '', group_id: 0, tags: [], sort_order: 0, enabled: true })

const filteredEntries = computed(() => {
  const query = search.value.toLowerCase()
  if (!query) return entries.value
  return entries.value.filter((entry) => [entry.model_name, entry.display_name, entry.group_name, entry.platform, ...entry.tags].some((value) => value.toLowerCase().includes(query)))
})

const groupOptions = computed<SelectOption[]>(() => groups.value.map((group) => ({
  value: group.id,
  label: `${group.name} · ${platformLabel(group.platform)} · ${formatRate(group.rate_multiplier)}x`
})))

const modelOptions = computed<SelectOption[]>(() => {
  const models = new Set(catalogModels.value)
  if (form.model_name) models.add(form.model_name)
  return Array.from(models).sort().map((value) => ({ value, label: value }))
})

const selectedGroup = computed(() => groups.value.find((group) => group.id === Number(form.group_id)))
const preview = computed(() => {
  if (!providerPricing.value || !selectedGroup.value || providerPricing.value.input_price == null || providerPricing.value.output_price == null) return null
  const factor = selectedGroup.value.rate_multiplier / 5
  return {
    input: providerPricing.value.input_price * 1_000_000 * factor,
    output: providerPricing.value.output_price * 1_000_000 * factor
  }
})

async function loadEntries() {
  loading.value = true
  try {
    entries.value = await adminAPI.modelPlaza.list()
  } catch (error: unknown) {
    appStore.showError(extractApiErrorMessage(error, t('admin.modelPlaza.loadFailed')))
  } finally {
    loading.value = false
  }
}

async function loadGroups() {
  groups.value = await adminAPI.groups.getAll()
}

function resetForm() {
  Object.assign(form, { model_name: '', display_name: '', description: '', group_id: 0, tags: [], sort_order: entries.value.length, enabled: true })
  tagsText.value = ''
  catalogModels.value = []
  providerPricing.value = null
}

function openCreate() {
  editing.value = null
  resetForm()
  showDialog.value = true
}

async function openEdit(entry: ModelPlazaEntry) {
  editing.value = entry
  Object.assign(form, { model_name: entry.model_name, display_name: entry.display_name, description: entry.description, group_id: entry.group_id, tags: [...entry.tags], sort_order: entry.sort_order, enabled: entry.enabled })
  tagsText.value = entry.tags.join(', ')
  showDialog.value = true
  await loadModelsForGroup()
  await loadPricing()
}

function closeDialog() {
  showDialog.value = false
  editing.value = null
}

async function onGroupChange() {
  form.model_name = ''
  providerPricing.value = null
  await loadModelsForGroup()
}

async function onModelChange() {
  await loadPricing()
}

async function loadModelsForGroup() {
  const group = selectedGroup.value
  if (!group) return
  modelsLoading.value = true
  try {
    const result = await adminAPI.channels.syncPricingModels(group.platform)
    catalogModels.value = result.models || []
  } catch (error: unknown) {
    appStore.showError(extractApiErrorMessage(error, t('admin.modelPlaza.modelLoadFailed')))
  } finally {
    modelsLoading.value = false
  }
}

async function loadPricing() {
  providerPricing.value = null
  if (!form.model_name) return
  try {
    providerPricing.value = await adminAPI.channels.getModelDefaultPricing(form.model_name)
  } catch {
    providerPricing.value = null
  }
}

function buildInput(): ModelPlazaEntryInput {
  return {
    ...form,
    group_id: Number(form.group_id),
    tags: tagsText.value.split(',').map((tag) => tag.trim()).filter(Boolean).slice(0, 8)
  }
}

async function saveEntry() {
  if (!form.group_id || !form.model_name || saving.value) return
  saving.value = true
  try {
    const input = buildInput()
    if (editing.value) {
      await adminAPI.modelPlaza.update(editing.value.id, input)
      appStore.showSuccess(t('admin.modelPlaza.updateSuccess'))
    } else {
      await adminAPI.modelPlaza.create(input)
      appStore.showSuccess(t('admin.modelPlaza.createSuccess'))
    }
    closeDialog()
    await loadEntries()
  } catch (error: unknown) {
    appStore.showError(extractApiErrorMessage(error, t('admin.modelPlaza.saveFailed')))
  } finally {
    saving.value = false
  }
}

async function toggleEntry(entry: ModelPlazaEntry, enabled: boolean) {
  const input: ModelPlazaEntryInput = { model_name: entry.model_name, display_name: entry.display_name, description: entry.description, group_id: entry.group_id, tags: entry.tags, sort_order: entry.sort_order, enabled }
  try {
    const updated = await adminAPI.modelPlaza.update(entry.id, input)
    Object.assign(entry, updated)
  } catch (error: unknown) {
    appStore.showError(extractApiErrorMessage(error, t('admin.modelPlaza.saveFailed')))
  }
}

function askDelete(entry: ModelPlazaEntry) { deleting.value = entry }

async function deleteEntry() {
  if (!deleting.value) return
  try {
    await adminAPI.modelPlaza.remove(deleting.value.id)
    deleting.value = null
    appStore.showSuccess(t('admin.modelPlaza.deleteSuccess'))
    await loadEntries()
  } catch (error: unknown) {
    appStore.showError(extractApiErrorMessage(error, t('common.error')))
  }
}

function formatPrice(value: number | null): string {
  if (value == null) return '-'
  if (value === 0) return '0'
  if (value < 0.001) return value.toFixed(6).replace(/0+$/, '').replace(/\.$/, '')
  if (value < 1) return value.toFixed(4).replace(/0+$/, '').replace(/\.$/, '')
  return value.toFixed(2).replace(/\.00$/, '')
}

function formatRate(value: number): string { return value.toFixed(4).replace(/0+$/, '').replace(/\.$/, '') }
function platformLabel(value: string): string { return ({ openai: 'OpenAI', anthropic: 'Anthropic', gemini: 'Gemini', antigravity: 'Antigravity', grok: 'Grok' } as Record<string, string>)[value] || value }

onMounted(async () => {
  await Promise.all([loadEntries(), loadGroups()])
})
</script>

<style scoped>
.table-action {
  display: inline-flex;
  height: 2rem;
  width: 2rem;
  align-items: center;
  justify-content: center;
  border-radius: 6px;
  color: rgb(107 114 128);
  transition: background-color 150ms, color 150ms;
}
.table-action:hover { background: rgb(243 244 246); color: rgb(17 24 39); }
.dark .table-action:hover { background: rgb(31 41 55); color: rgb(243 244 246); }
</style>
