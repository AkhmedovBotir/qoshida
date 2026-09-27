import { api } from './api'
import type { ActivityStatus, ActivityTypeItem, ActivityTypeListResult } from '../types/activityType'

export async function listActivityTypes(q = '', status = ''): Promise<ActivityTypeItem[]> {
  const items: ActivityTypeItem[] = []
  let page = 1
  const limit = 100

  while (true) {
    const query = new URLSearchParams({ page: String(page), limit: String(limit) })
    if (q) query.set('q', q)
    if (status) query.set('status', status)
    const data = await api<ActivityTypeListResult>(`/api/v1/activity-types?${query}`)
    items.push(...(data.items ?? []))
    if (items.length >= data.total || (data.items ?? []).length === 0) break
    page += 1
  }

  return items
}

export function createActivityType(body: { name: string; icon: string; status: ActivityStatus }) {
  return api<ActivityTypeItem>('/api/v1/activity-types', { method: 'POST', body })
}

export function updateActivityType(id: string, body: { name: string; icon: string; status: ActivityStatus }) {
  return api<ActivityTypeItem>(`/api/v1/activity-types/${id}`, { method: 'PUT', body })
}

export function deleteActivityType(id: string) {
  return api(`/api/v1/activity-types/${id}`, { method: 'DELETE' })
}

export function importActivityTypes() {
  return api<{ inserted: number; updated: number; total: number }>('/api/v1/activity-types/import', {
    method: 'POST',
    body: {},
    timeoutMs: 60_000,
  })
}
