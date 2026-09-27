import type { LocalShopForm, LocalShopItem, LocalShopListResult } from '../types/localShop'
import { api } from './api'

export function listLocalShops(
  params: { page?: number; limit?: number; q?: string; region_id?: string; district_id?: string; mfy_id?: string } = {},
) {
  const query = new URLSearchParams({
    page: String(params.page ?? 1),
    limit: String(params.limit ?? 20),
  })
  if (params.q) query.set('q', params.q)
  if (params.region_id) query.set('region_id', params.region_id)
  if (params.district_id) query.set('district_id', params.district_id)
  if (params.mfy_id) query.set('mfy_id', params.mfy_id)
  return api<LocalShopListResult>(`/api/v1/local-shops?${query}`)
}

export function createLocalShop(body: LocalShopForm) {
  return api<LocalShopItem>('/api/v1/local-shops', { method: 'POST', body })
}

export function updateLocalShop(id: string, body: LocalShopForm) {
  return api<LocalShopItem>(`/api/v1/local-shops/${id}`, { method: 'PUT', body })
}

export function deleteLocalShop(id: string) {
  return api(`/api/v1/local-shops/${id}`, { method: 'DELETE' })
}
