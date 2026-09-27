import type { KontragentItem, KontragentListResult } from '../types/kontragent'
import { api } from './api'

export function listKontragents(params: { page?: number; limit?: number; q?: string } = {}) {
  const query = new URLSearchParams({
    page: String(params.page ?? 1),
    limit: String(params.limit ?? 100),
  })
  if (params.q) query.set('q', params.q)
  return api<KontragentListResult>(`/api/v1/manager/kontragents?${query}`)
}

export type { KontragentItem }
