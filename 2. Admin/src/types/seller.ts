export type SellerStatus = 'active' | 'inactive'

export type SellerItem = {
  audit?: import('../lib/audit').AuditTrail
  id: string
  shop_id: string
  first_name: string
  last_name: string
  phone: string
  status: SellerStatus
  has_password: boolean
  shop_name?: string
  region_id?: string
  district_id?: string
  mfy_id?: string
  region_name?: string
  district_name?: string
  mfy_name?: string
}

export type SellerListResult = {
  items: SellerItem[]
  total: number
  page: number
  limit: number
}

export type SellerForm = {
  shop_id: string
  region_id: string
  district_id: string
  mfy_id: string
  first_name: string
  last_name: string
  phone: string
  status: SellerStatus
  password: string
}
