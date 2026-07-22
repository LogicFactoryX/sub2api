import { apiClient } from './client'

export interface ModelPlazaPricing {
  input_per_million: number | null
  output_per_million: number | null
  cache_write_per_million: number | null
  cache_read_per_million: number | null
}

export interface ModelPlazaEntry {
  id: number
  model_name: string
  display_name: string
  description: string
  group_id: number
  group_name: string
  platform: string
  group_status: string
  rate_multiplier: number
  is_exclusive: boolean
  is_member_group: boolean
  tags: string[]
  sort_order: number
  enabled: boolean
  pricing_available: boolean
  original_pricing: ModelPlazaPricing
  display_pricing: ModelPlazaPricing
  created_at: string
  updated_at: string
}

export async function list(): Promise<ModelPlazaEntry[]> {
  const { data } = await apiClient.get<ModelPlazaEntry[]>('/model-plaza')
  return data
}

export default { list }
