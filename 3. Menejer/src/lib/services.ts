import { API_URL } from '../constants/config'
import type { ServiceProviderListResult } from '../types/serviceProvider'
import type { ProviderServiceForm, ProviderServiceItem, ProviderServiceListResult, ProviderServiceRow } from '../types/providerService'
import { api } from './api'

export const OWN_SERVICES = false
export const SERVICES_BASE = '/api/v1/manager/provider-services'
export const PROVIDERS_PATH = '/api/v1/manager/service-providers'

export function serviceImageSrc(url?: string | null) {
  if (!url) return ''
  if (url.startsWith('http://') || url.startsWith('https://') || url.startsWith('data:')) return url
  return `${API_URL}${url}`
}

export function formatMoney(value: number) {
  return new Intl.NumberFormat('uz-UZ').format(value)
}

export function listServices(params: { page?: number; limit?: number; q?: string; provider_id?: string; approval_status?: string } = {}) {
  const query = new URLSearchParams({
    page: String(params.page ?? 1),
    limit: String(params.limit ?? 20),
  })
  if (params.q) query.set('q', params.q)
  if (params.provider_id) query.set('provider_id', params.provider_id)
  if (params.approval_status) query.set('approval_status', params.approval_status)
  return api<ProviderServiceListResult>(`${SERVICES_BASE}?${query}`)
}

export function listProvidersForSelect() {
  return api<ServiceProviderListResult>(`${PROVIDERS_PATH}?page=1&limit=200`)
}

export function createServices(body: ProviderServiceForm) {
  return api<{ items: ProviderServiceItem[] }>(SERVICES_BASE, {
    method: 'POST',
    timeoutMs: 120_000,
    body: {
      provider_id: body.provider_id,
      items: body.items.map((item) => ({
        name: item.name.trim(),
        price: Number(item.price),
        images: item.images ?? [],
      })),
    },
  })
}

export function updateService(id: string, body: ProviderServiceRow) {
  return api<ProviderServiceItem>(`${SERVICES_BASE}/${id}`, {
    method: 'PUT',
    timeoutMs: 120_000,
    body: { name: body.name.trim(), price: Number(body.price), images: body.images ?? [] },
  })
}

export function deleteService(id: string) {
  return api(`${SERVICES_BASE}/${id}`, { method: 'DELETE' })
}

export function approveService(id: string) {
  return api<ProviderServiceItem>(`${SERVICES_BASE}/${id}/approve`, { method: 'POST', body: {} })
}

export function rejectService(id: string, note: string) {
  return api<ProviderServiceItem>(`${SERVICES_BASE}/${id}/reject`, { method: 'POST', body: { note } })
}
