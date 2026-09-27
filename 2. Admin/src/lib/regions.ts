import { api } from './api'
import type { ApiRegionType, RegionItem, RegionListResult, RegionStatus } from '../types/region'

type ListParams = {
  type?: ApiRegionType
  parent_id?: string
  q?: string
}

export async function listRegions(params: ListParams = {}): Promise<RegionItem[]> {
  const items: RegionItem[] = []
  let page = 1
  const limit = 100

  while (true) {
    const query = new URLSearchParams({ page: String(page), limit: String(limit) })
    if (params.type) query.set('type', params.type)
    if (params.parent_id) query.set('parent_id', params.parent_id)
    if (params.q) query.set('q', params.q)

    const data = await api<RegionListResult>(`/api/v1/regions?${query}`)
    items.push(...(data.items ?? []))
    if (items.length >= data.total || (data.items ?? []).length === 0) break
    page += 1
  }

  return items
}

export function createRegion(body: {
  name: string
  type: ApiRegionType
  parent_id: string | null
  code: string
  status: RegionStatus
}) {
  return api<RegionItem>('/api/v1/regions', { method: 'POST', body })
}

export function updateRegion(
  id: string,
  body: {
    name: string
    type: ApiRegionType
    parent_id: string | null
    code: string
    status: RegionStatus
  },
) {
  return api<RegionItem>(`/api/v1/regions/${id}`, { method: 'PUT', body })
}

export function deleteRegion(id: string) {
  return api(`/api/v1/regions/${id}`, { method: 'DELETE' })
}

export function importRegions() {
  return api<{ inserted: number; updated: number; total: number }>('/api/v1/regions/import', {
    method: 'POST',
    body: {},
    timeoutMs: 120_000,
  })
}
