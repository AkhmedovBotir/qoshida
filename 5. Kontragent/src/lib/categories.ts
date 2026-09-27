import { api } from './api'
import type { CategoryItem, CategoryListResult } from '../types/category'

const BASE = '/api/v1/kontragent/categories'

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
