export type ServiceProviderStatus = 'active' | 'inactive'

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
  status: ServiceProviderStatus
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

export type ServiceProviderForm = {
  activity_type_id: string
  name: string
  phone: string
  image: string
  clear_image: boolean
  region_id: string
  district_id: string
  mfy_id: string
  status: ServiceProviderStatus
  password: string
}
