export type LocalShopItem = {
  audit?: import('../lib/audit').AuditTrail
  id: string
  name: string
  phone: string
  image: string
  region_id: string
  district_id: string
  mfy_id: string
  status: 'active' | 'inactive'
  has_password: boolean
  region_name?: string
  district_name?: string
  mfy_name?: string
}

export type LocalShopListResult = {
  items: LocalShopItem[]
  total: number
  page: number
  limit: number
}

export type AreaItem = {
  id: string
  name: string
}
