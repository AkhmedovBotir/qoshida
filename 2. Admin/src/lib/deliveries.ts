import type { DeliveryForm, DeliveryItem, DeliveryListResult } from '../types/delivery'
import { api } from './api'

export function listDeliveries(params: { page?: number; limit?: number; q?: string } = {}) {
  const query = new URLSearchParams({
    page: String(params.page ?? 1),
    limit: String(params.limit ?? 20),
  })
  if (params.q) query.set('q', params.q)
  return api<DeliveryListResult>(`/api/v1/deliveries?${query}`)
}

function toWritePayload(body: DeliveryForm) {
  return {
    shop_id: body.shop_id,
    first_name: body.first_name,
    last_name: body.last_name,
    phone: body.phone,
    status: body.status,
    password: body.password,
  }
}

export function createDelivery(body: DeliveryForm) {
  return api<DeliveryItem>('/api/v1/deliveries', { method: 'POST', body: toWritePayload(body) })
}

export function updateDelivery(id: string, body: DeliveryForm) {
  return api<DeliveryItem>(`/api/v1/deliveries/${id}`, { method: 'PUT', body: toWritePayload(body) })
}

export function deleteDelivery(id: string) {
  return api(`/api/v1/deliveries/${id}`, { method: 'DELETE' })
}
