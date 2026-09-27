import { API_URL } from '../constants/config'
import type { ProductForm, ProductItem, ProductListResult } from '../types/product'
import { api } from './api'

const BASE = '/api/v1/manager/products'

export function productImageSrc(url?: string | null) {
  if (!url) return ''
  if (url.startsWith('http://') || url.startsWith('https://') || url.startsWith('data:')) return url
  return `${API_URL}${url}`
}

export function listProducts(params: { page?: number; limit?: number; q?: string; approval_status?: string } = {}) {
  const query = new URLSearchParams({
    page: String(params.page ?? 1),
    limit: String(params.limit ?? 20),
  })
  if (params.q) query.set('q', params.q)
  if (params.approval_status) query.set('approval_status', params.approval_status)
  return api<ProductListResult>(`${BASE}?${query}`)
}

function toWritePayload(body: ProductForm) {
  return {
    kontragent_id: body.kontragent_id,
    category_id: body.category_id,
    subcategory_id: body.subcategory_id,
    name: body.name,
    description: body.description,
    sale_price: Number(body.sale_price),
    cost_price: Number(body.cost_price),
    quantity: Number(body.quantity),
    unit: body.unit,
    unit_size: Number(body.unit_size),
    commission_percent: Number(body.commission_percent),
    images: body.images,
    status: body.status,
  }
}

export function createProduct(body: ProductForm) {
  return api<ProductItem>(BASE, { method: 'POST', body: toWritePayload(body), timeoutMs: 120_000 })
}

export function updateProduct(id: string, body: ProductForm) {
  return api<ProductItem>(`${BASE}/${id}`, { method: 'PUT', body: toWritePayload(body), timeoutMs: 120_000 })
}

export function deleteProduct(id: string) {
  return api(`${BASE}/${id}`, { method: 'DELETE' })
}

export function approveProduct(id: string) {
  return api<ProductItem>(`${BASE}/${id}/approve`, { method: 'POST', body: {} })
}

export function rejectProduct(id: string, note: string) {
  return api<ProductItem>(`${BASE}/${id}/reject`, { method: 'POST', body: { note } })
}

export function formatMoney(value: number) {
  return new Intl.NumberFormat('uz-UZ').format(value)
}

export function commissionAmount(sale: number, cost: number, percent: number) {
  const margin = Math.max(0, sale - cost)
  return Math.round((margin * percent) / 100)
}
