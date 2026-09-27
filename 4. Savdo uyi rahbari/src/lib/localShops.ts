import { api } from './api'

export type LocalShopItem = {
  id: string
  name: string
}

type LocalShopListResult = {
  items: LocalShopItem[]
  total: number
}

export function listLocalShops(params: { page?: number; limit?: number; q?: string } = {}) {
  const query = new URLSearchParams({
    page: String(params.page ?? 1),
    limit: String(params.limit ?? 20),
  })
  if (params.q) query.set('q', params.q)
  return api<LocalShopListResult>(`/api/v1/shop-director/local-shops?${query}`)
}
