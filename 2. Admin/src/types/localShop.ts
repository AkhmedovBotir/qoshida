export type LocalShopStatus = 'active' | 'inactive'

export type LocalShopItem = {
  audit?: import('../lib/audit').AuditTrail
  id: string
  name: string
  phone: string
  image: string
  region_id: string
  district_id: string
  mfy_id: string
  status: LocalShopStatus
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

export type LocalShopForm = {
  name: string
  phone: string
  image: string
  clear_image: boolean
  region_id: string
  district_id: string
  mfy_id: string
  status: LocalShopStatus
  password: string
}
