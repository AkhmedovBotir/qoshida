import type { ShopDirectorForm, ShopDirectorItem, ShopDirectorListResult } from '../types/shopDirector'
import { api } from './api'

export function listShopDirectors(params: { page?: number; limit?: number; q?: string } = {}) {
  const query = new URLSearchParams({
    page: String(params.page ?? 1),
    limit: String(params.limit ?? 20),
  })
  if (params.q) query.set('q', params.q)
  return api<ShopDirectorListResult>(`/api/v1/shop-directors?${query}`)
}

export function createShopDirector(body: ShopDirectorForm) {
  return api<ShopDirectorItem>('/api/v1/shop-directors', { method: 'POST', body })
}

export function updateShopDirector(id: string, body: ShopDirectorForm) {
  return api<ShopDirectorItem>(`/api/v1/shop-directors/${id}`, { method: 'PUT', body })
}

export function deleteShopDirector(id: string) {
  return api(`/api/v1/shop-directors/${id}`, { method: 'DELETE' })
}
