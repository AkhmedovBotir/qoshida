export type ManagerType = 'region' | 'district'
export type ManagerStatus = 'active' | 'inactive'

export type ManagerUser = {
  audit?: import('../lib/audit').AuditTrail
  id: string
  type: ManagerType
  region_id: string
  district_id: string | null
  first_name: string
  last_name: string
  phone: string
  username: string
  status: ManagerStatus
  has_password: boolean
  region_name?: string
  district_name?: string
}

export type PhoneCheck = {
  exists: boolean
  active: boolean
  has_password: boolean
  needs_setup: boolean
  message: string
}

export type ManagerListResult = {
  items: ManagerUser[]
  total: number
  page: number
  limit: number
}

export type DistrictItem = {
  id: string
  name: string
}
