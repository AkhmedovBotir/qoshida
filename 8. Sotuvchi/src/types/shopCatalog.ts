export type ProductUnit = 'dona' | 'litr' | 'kg'

export type ShopTemplateItem = {
  audit?: import('../lib/audit').AuditTrail
  id: string
  category_id: string
  subcategory_id: string
  name: string
  description: string
  unit: ProductUnit
  unit_size: number
  images: string[]
  category_name?: string
  subcategory_name?: string
}

export type ShopTemplateListResult = {
  items: ShopTemplateItem[]
  total: number
  page: number
  limit: number
}

export type ShopTemplateForm = {
  category_id: string
  subcategory_id: string
  name: string
  description: string
  unit: ProductUnit
  unit_size: string
  images: string[]
}

export type ShopProductItem = {
  audit?: import('../lib/audit').AuditTrail
  id: string
  shop_id: string
  template_id: string
  quantity: number
  sale_price: number
  cost_price: number
  name: string
  description: string
  category_id: string
  subcategory_id: string
  category_name?: string
  subcategory_name?: string
  unit: ProductUnit
  unit_size: number
  images: string[]
  shop_name?: string
  region_name?: string
  district_name?: string
  mfy_name?: string
}

export type ShopProductListResult = {
  items: ShopProductItem[]
  total: number
  page: number
  limit: number
}

export type ShopProductForm = {
  shop_id: string
  template_id: string
  quantity: string
  sale_price: string
  cost_price: string
}

export type ShopIncomingItem = {
  audit?: import('../lib/audit').AuditTrail
  id: string
  shop_product_id: string
  quantity: number
  product_name?: string
  shop_id?: string
  shop_name?: string
  unit?: ProductUnit
  unit_size?: number
  created_at: string
}

export type ShopIncomingListResult = {
  items: ShopIncomingItem[]
  total: number
  page: number
  limit: number
}
