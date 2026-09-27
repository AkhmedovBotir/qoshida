export type KontragentItem = {
  audit?: import('../lib/audit').AuditTrail
  id: string
  activity_type_id: string
  name: string
  inn: string
  region_id: string
  district_id: string
  mfy_id: string
  phone: string
  logo: string
  status: 'active' | 'inactive'
  has_password: boolean
  activity_name?: string
  region_name?: string
  district_name?: string
  mfy_name?: string
}

export type KontragentListResult = {
  items: KontragentItem[]
  total: number
  page: number
  limit: number
}

export type AreaItem = {
  id: string
  name: string
}
