import type { ManagerForm, ManagerItem, ManagerListResult, ManagerType } from '../types/manager'
import { api } from './api'

export function listManagers(params: { page?: number; limit?: number; q?: string; type?: ManagerType | '' } = {}) {
  const query = new URLSearchParams({
    page: String(params.page ?? 1),
    limit: String(params.limit ?? 20),
  })
  if (params.q) query.set('q', params.q)
  if (params.type) query.set('type', params.type)
  return api<ManagerListResult>(`/api/v1/managers?${query}`)
}

export function createManager(body: Omit<ManagerForm, 'password'> & { password?: string }) {
  return api<ManagerItem>('/api/v1/managers', { method: 'POST', body })
}

export function updateManager(id: string, body: ManagerForm) {
  return api<ManagerItem>(`/api/v1/managers/${id}`, { method: 'PUT', body })
}

export function deleteManager(id: string) {
  return api(`/api/v1/managers/${id}`, { method: 'DELETE' })
}
