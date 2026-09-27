export type ServiceProviderItem = {
  audit?: import('../lib/audit').AuditTrail
  id: string
  activity_type_id: string
  name: string
  phone: string
  image: string
  region_id: string
  district_id: string
  mfy_id: string
  status: 'active' | 'inactive'
  has_password: boolean
  activity_name?: string
  region_name?: string
  district_name?: string
  mfy_name?: string
}

export type ServiceProviderListResult = {
  items: ServiceProviderItem[]
  total: number
  page: number
  limit: number
}

export type AreaItem = {
  id: string
  name: string
}
