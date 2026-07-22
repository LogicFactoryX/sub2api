import { apiClient } from '../client'
import type { ModelPlazaEntry } from '../modelPlaza'

export interface ModelPlazaEntryInput {
  model_name: string
  display_name: string
  description: string
  group_id: number
  tags: string[]
  sort_order: number
  enabled: boolean
}

export async function list(): Promise<ModelPlazaEntry[]> {
  const { data } = await apiClient.get<ModelPlazaEntry[]>('/admin/model-plaza')
  return data
}

export async function create(input: ModelPlazaEntryInput): Promise<ModelPlazaEntry> {
  const { data } = await apiClient.post<ModelPlazaEntry>('/admin/model-plaza', input)
  return data
}

export async function update(id: number, input: ModelPlazaEntryInput): Promise<ModelPlazaEntry> {
  const { data } = await apiClient.put<ModelPlazaEntry>(`/admin/model-plaza/${id}`, input)
  return data
}

export async function remove(id: number): Promise<void> {
  await apiClient.delete(`/admin/model-plaza/${id}`)
}

export default { list, create, update, remove }
