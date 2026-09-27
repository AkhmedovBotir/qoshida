import type { SellerForm, SellerItem, SellerListResult } from '../types/seller'
import { api } from './api'

export function listSellers(params: { page?: number; limit?: number; q?: string } = {}) {
  const query = new URLSearchParams({
    page: String(params.page ?? 1),
    limit: String(params.limit ?? 20),
  })
  if (params.q) query.set('q', params.q)
  return api<SellerListResult>(`/api/v1/sellers?${query}`)
}

function toWritePayload(body: SellerForm) {
  return {
    shop_id: body.shop_id,
    first_name: body.first_name,
    last_name: body.last_name,
    phone: body.phone,
    status: body.status,
    password: body.password,
  }
}

export function createSeller(body: SellerForm) {
  return api<SellerItem>('/api/v1/sellers', { method: 'POST', body: toWritePayload(body) })
}

export function updateSeller(id: string, body: SellerForm) {
  return api<SellerItem>(`/api/v1/sellers/${id}`, { method: 'PUT', body: toWritePayload(body) })
}

export function deleteSeller(id: string) {
  return api(`/api/v1/sellers/${id}`, { method: 'DELETE' })
}
