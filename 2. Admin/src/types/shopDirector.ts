export type ShopDirectorStatus = 'active' | 'inactive'

export type ShopDirectorItem = {
  audit?: import('../lib/audit').AuditTrail
  id: string
  region_id: string
  district_id: string
  mfy_id: string
  first_name: string
  last_name: string
  phone: string
  status: ShopDirectorStatus
  has_password: boolean
  region_name?: string
  district_name?: string
  mfy_name?: string
}

export type ShopDirectorListResult = {
  items: ShopDirectorItem[]
  total: number
  page: number
  limit: number
}

export type ShopDirectorForm = {
  region_id: string
  district_id: string
  mfy_id: string
  first_name: string
  last_name: string
  phone: string
  status: ShopDirectorStatus
  password: string
}
