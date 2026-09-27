import type { ServiceProviderForm, ServiceProviderItem, ServiceProviderListResult } from '../types/serviceProvider'
import { api } from './api'

export function listServiceProviders(params: { page?: number; limit?: number; q?: string } = {}) {
  const query = new URLSearchParams({
    page: String(params.page ?? 1),
    limit: String(params.limit ?? 20),
  })
  if (params.q) query.set('q', params.q)
  return api<ServiceProviderListResult>(`/api/v1/service-providers?${query}`)
}

export function createServiceProvider(body: ServiceProviderForm) {
  return api<ServiceProviderItem>('/api/v1/service-providers', { method: 'POST', body })
}

export function updateServiceProvider(id: string, body: ServiceProviderForm) {
  return api<ServiceProviderItem>(`/api/v1/service-providers/${id}`, { method: 'PUT', body })
}

export function deleteServiceProvider(id: string) {
  return api(`/api/v1/service-providers/${id}`, { method: 'DELETE' })
}
