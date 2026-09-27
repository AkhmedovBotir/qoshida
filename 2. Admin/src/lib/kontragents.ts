import type { KontragentForm, KontragentItem, KontragentListResult } from '../types/kontragent'
import { api } from './api'

export function listKontragents(params: { page?: number; limit?: number; q?: string } = {}) {
  const query = new URLSearchParams({
    page: String(params.page ?? 1),
    limit: String(params.limit ?? 20),
  })
  if (params.q) query.set('q', params.q)
  return api<KontragentListResult>(`/api/v1/kontragents?${query}`)
}

export function createKontragent(body: KontragentForm) {
  return api<KontragentItem>('/api/v1/kontragents', { method: 'POST', body })
}

export function updateKontragent(id: string, body: KontragentForm) {
  return api<KontragentItem>(`/api/v1/kontragents/${id}`, { method: 'PUT', body })
}

export function deleteKontragent(id: string) {
  return api(`/api/v1/kontragents/${id}`, { method: 'DELETE' })
}
