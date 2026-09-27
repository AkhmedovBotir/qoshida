export type CatalogKind = 'product' | 'shop' | 'service'

export type CatalogItem = {
  kind: CatalogKind
  id: string
  name: string
  description?: string
  price: number
  quantity: number
  unit: string
  unit_size?: number
  images: string[]
  image?: string
  seller_name?: string
  seller_id?: string
  category_id?: string
  category_name?: string
  region_name?: string
  district_name?: string
  mfy_name?: string
}

export type CategoryItem = {
  id: string
  name: string
  slug?: string
  image_url?: string
  parent_id?: string
}

export type CatalogList = {
  items: CatalogItem[]
  total: number
  page: number
  limit: number
}

export type CartLine = {
  id: string
  kind: CatalogKind
  item_id: string
  quantity: number
  item?: CatalogItem
  line_total: number
}

export type CartPublic = {
  items: CartLine[]
  total: number
  count: number
}

export type OrderItem = {
  kind: CatalogKind
  item_id: string
  name: string
  unit: string
  quantity: number
  unit_price: number
  seller_name?: string
  image?: string
  line_total: number
}

export type OrderStatus = 'pending' | 'confirmed' | 'delivering' | 'done' | 'cancelled'

export type OrderPublic = {
  id: string
  status: OrderStatus
  total: number
  address: string
  note?: string
  created_at: string
  items: OrderItem[]
}

export type OrderList = {
  items: OrderPublic[]
  total: number
  page: number
  limit: number
}

export type RegionItem = {
  id: string
  name: string
  type: 'region' | 'district' | 'mfy'
  parent_id?: string | null
}

export type RegionList = {
  items: RegionItem[]
  total: number
  page: number
  limit: number
}
