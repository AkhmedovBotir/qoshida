import type {
  CartPublic,
  CatalogItem,
  CatalogKind,
  CatalogList,
  CategoryItem,
  OrderList,
  OrderPublic,
  RegionList,
} from '../types/market'
import { api } from './api'

export type CatalogQuery = {
  q?: string
  kind?: CatalogKind | ''
  category_id?: string
  region_id?: string
  page?: number
  limit?: number
}

function toQuery(params: Record<string, string | number | undefined>) {
  const query = new URLSearchParams()
  Object.entries(params).forEach(([key, value]) => {
    if (value === undefined || value === '') return
    query.set(key, String(value))
  })
  return query.toString()
}

export function listCatalog(params: CatalogQuery = {}) {
  const qs = toQuery({
    q: params.q,
    kind: params.kind,
    category_id: params.category_id,
    region_id: params.region_id,
    page: params.page ?? 1,
    limit: params.limit ?? 20,
  })
  return api<CatalogList>(`/api/v1/public/catalog?${qs}`)
}

export function getCatalogItem(kind: CatalogKind, id: string) {
  return api<CatalogItem>(`/api/v1/public/catalog/${kind}/${id}`)
}

export function listCategories(params: { parent_id?: string; roots?: boolean } = {}) {
  const qs = toQuery({
    parent_id: params.parent_id,
    roots: params.roots ? '1' : params.parent_id ? '0' : '1',
  })
  return api<CategoryItem[]>(`/api/v1/public/categories?${qs}`)
}

export function listPublicRegions(params: { type?: string; parent_id?: string; limit?: number } = {}) {
  const qs = toQuery({
    type: params.type,
    parent_id: params.parent_id,
    page: 1,
    limit: params.limit ?? 200,
  })
  return api<RegionList>(`/api/v1/public/regions?${qs}`)
}

export function getCart() {
  return api<CartPublic>('/api/v1/mijoz/cart')
}

export function upsertCart(
  body: { kind: CatalogKind; item_id: string; quantity: number },
  opts?: { silent?: boolean },
) {
  return api<CartPublic>('/api/v1/mijoz/cart', { method: 'PUT', body, silent: opts?.silent })
}

export function checkout(body: {
  address: string
  note?: string
  region_id?: string
  district_id?: string
  mfy_id?: string
}) {
  return api<OrderPublic>('/api/v1/mijoz/checkout', { method: 'POST', body })
}

export function listOrders(page = 1) {
  return api<OrderList>(`/api/v1/mijoz/orders?page=${page}&limit=20`)
}

export function getOrder(id: string) {
  return api<OrderPublic>(`/api/v1/mijoz/orders/${id}`)
}

export function getProfile() {
  return api<import('../types/customer').CustomerUser>('/api/v1/mijoz/profile')
}

export function updateProfile(body: {
  first_name: string
  last_name: string
  birth_date: string
  region_id?: string
  district_id?: string
  mfy_id?: string
  avatar?: string
  clear_avatar?: boolean
}) {
  return api<import('../types/customer').CustomerUser>('/api/v1/mijoz/profile', {
    method: 'PUT',
    body,
    timeoutMs: 30_000,
  })
}
