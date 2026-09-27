export type DeliveryStatus = 'active' | 'inactive'

export type DeliveryItem = {
  audit?: import('../lib/audit').AuditTrail
  id: string
  shop_id: string
  first_name: string
  last_name: string
  phone: string
  status: DeliveryStatus
  has_password: boolean
  shop_name?: string
  region_id?: string
  district_id?: string
  mfy_id?: string
  region_name?: string
  district_name?: string
  mfy_name?: string
}

export type DeliveryListResult = {
  items: DeliveryItem[]
  total: number
  page: number
  limit: number
}

export type DeliveryForm = {
  shop_id: string
  region_id: string
  district_id: string
  mfy_id: string
  first_name: string
  last_name: string
  phone: string
  status: DeliveryStatus
  password: string
}
