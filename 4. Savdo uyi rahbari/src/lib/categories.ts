import { API_URL } from '../constants/config'
import { api } from './api'
import type { CategoryItem, CategoryListResult, CategoryStatus } from '../types/category'

const BASE = '/api/v1/shop-director/categories'

export function categoryImageSrc(url?: string | null) {
  if (!url) return ''
  if (url.startsWith('http://') || url.startsWith('https://') || url.startsWith('data:')) return url
  return `${API_URL}${url}`
}

export async function listCategories(): Promise<CategoryItem[]> {
  const items: CategoryItem[] = []
  let page = 1
  const limit = 100
  while (true) {
    const query = new URLSearchParams({ page: String(page), limit: String(limit) })
    const data = await api<CategoryListResult>(`${BASE}?${query}`)
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
  return api<CategoryItem>(BASE, { method: 'POST', body, timeoutMs: 30_000 })
}

export function updateCategory(id: string, body: CategoryWrite) {
  return api<CategoryItem>(`${BASE}/${id}`, { method: 'PUT', body, timeoutMs: 30_000 })
}

export function deleteCategory(id: string) {
  return api(`${BASE}/${id}`, { method: 'DELETE' })
}
