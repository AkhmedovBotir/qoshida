import type { Identification, IdentificationListResult } from '../types/identification'
import { api } from './api'

export const IDENT_BASE = '/api/v1/manager/provider-identifications'

export function listIdentifications(params: { page?: number; limit?: number; q?: string; status?: string } = {}) {
  const query = new URLSearchParams({
    page: String(params.page ?? 1),
    limit: String(params.limit ?? 20),
  })
  if (params.q) query.set('q', params.q)
  if (params.status) query.set('status', params.status)
  return api<IdentificationListResult>(`${IDENT_BASE}?${query}`)
}

export function getIdentification(id: string) {
  return api<Identification>(`${IDENT_BASE}/${id}`)
}

export function approveIdentification(id: string) {
  return api<Identification>(`${IDENT_BASE}/${id}/approve`, { method: 'POST', body: {} })
}

export function rejectIdentification(id: string, note: string) {
  return api<Identification>(`${IDENT_BASE}/${id}/reject`, { method: 'POST', body: { note } })
}
