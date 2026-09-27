import { API_URL } from '../constants/config'
import type {
  ShopIncomingItem,
  ShopIncomingListResult,
  ShopProductForm,
  ShopProductItem,
  ShopProductListResult,
  ShopTemplateListResult,
} from '../types/shopCatalog'
import { api } from './api'

const TEMPLATES = '/api/v1/local-shop/shop-templates'
const PRODUCTS = '/api/v1/local-shop/shop-products'
const INCOMINGS = '/api/v1/local-shop/shop-incomings'

export function catalogImageSrc(url?: string | null) {
  if (!url) return ''
  if (url.startsWith('http://') || url.startsWith('https://') || url.startsWith('data:')) return url
  return `${API_URL}${url}`
}

export function formatMoney(value: number) {
  return new Intl.NumberFormat('uz-UZ').format(value)
}

export function listShopTemplates(params: { page?: number; limit?: number; q?: string } = {}) {
  const query = new URLSearchParams({
    page: String(params.page ?? 1),
    limit: String(params.limit ?? 20),
  })
  if (params.q) query.set('q', params.q)
  return api<ShopTemplateListResult>(`${TEMPLATES}?${query}`)
}

export function listShopProducts(params: { page?: number; limit?: number; q?: string } = {}) {
  const query = new URLSearchParams({
    page: String(params.page ?? 1),
    limit: String(params.limit ?? 20),
  })
  if (params.q) query.set('q', params.q)
  return api<ShopProductListResult>(`${PRODUCTS}?${query}`)
}

export function attachShopProduct(body: ShopProductForm) {
  return api<ShopProductItem>(PRODUCTS, {
    method: 'POST',
    body: {
      shop_id: body.shop_id,
      template_id: body.template_id,
      quantity: Number(body.quantity),
      sale_price: Number(body.sale_price),
      cost_price: Number(body.cost_price),
    },
  })
}

export function updateShopProduct(id: string, body: { sale_price: string; cost_price: string }) {
  return api<ShopProductItem>(`${PRODUCTS}/${id}`, {
    method: 'PUT',
    body: {
      sale_price: Number(body.sale_price),
      cost_price: Number(body.cost_price),
    },
  })
}

export function deleteShopProduct(id: string) {
  return api(`${PRODUCTS}/${id}`, { method: 'DELETE' })
}

export function listShopIncomings(params: { page?: number; limit?: number; q?: string } = {}) {
  const query = new URLSearchParams({
    page: String(params.page ?? 1),
    limit: String(params.limit ?? 20),
  })
  if (params.q) query.set('q', params.q)
  return api<ShopIncomingListResult>(`${INCOMINGS}?${query}`)
}

export function createShopIncoming(body: { shop_product_id: string; quantity: string }) {
  return api<ShopIncomingItem>(INCOMINGS, {
    method: 'POST',
    body: {
      shop_product_id: body.shop_product_id,
      quantity: Number(body.quantity),
    },
  })
}
