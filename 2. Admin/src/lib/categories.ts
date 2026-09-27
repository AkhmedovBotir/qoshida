import { API_URL } from '../constants/config'
import { api } from './api'
import type { CategoryItem, CategoryListResult, CategoryStatus } from '../types/category'

export function categoryImageSrc(url?: string | null) {
  if (!url) return ''
  if (url.startsWith('http://') || url.startsWith('https://') || url.startsWith('data:')) return url
  return `${API_URL}${url}`
}

type ListParams = {
  parent_id?: string
  roots?: boolean
  q?: string
}

export async function listCategories(params: ListParams = {}): Promise<CategoryItem[]> {
  const items: CategoryItem[] = []
  let page = 1
  const limit = 100

  while (true) {
    const query = new URLSearchParams({ page: String(page), limit: String(limit) })
    if (params.parent_id) query.set('parent_id', params.parent_id)
    if (params.roots) query.set('roots', '1')
    if (params.q) query.set('q', params.q)

    const data = await api<CategoryListResult>(`/api/v1/categories?${query}`)
    items.push(...(data.items ?? []))
    if (items.length >= data.total || (data.items ?? []).length === 0) break
    page += 1
  }

  return items
}

export type CategoryWrite = {
  name: string
  slug: string
  parent_id: string | null
  censored: boolean
  status: CategoryStatus
  image?: string
  clear_image?: boolean
}

export function createCategory(body: CategoryWrite) {
  return api<CategoryItem>('/api/v1/categories', { method: 'POST', body, timeoutMs: 30_000 })
}

export function updateCategory(id: string, body: CategoryWrite) {
  return api<CategoryItem>(`/api/v1/categories/${id}`, { method: 'PUT', body, timeoutMs: 30_000 })
}

export function deleteCategory(id: string) {
  return api(`/api/v1/categories/${id}`, { method: 'DELETE' })
}

export function importCategories() {
  return api<{ inserted: number; updated: number; total: number; images: number }>('/api/v1/categories/import', {
    method: 'POST',
    body: {},
    timeoutMs: 120_000,
  })
}
