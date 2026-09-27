export type ManagerType = 'region' | 'district'
export type ManagerStatus = 'active' | 'inactive'

export type ManagerItem = {
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

export type ManagerListResult = {
  items: ManagerItem[]
  total: number
  page: number
  limit: number
}

export type ManagerForm = {
  type: ManagerType
  region_id: string
  district_id: string
  first_name: string
  last_name: string
  phone: string
  username: string
  status: ManagerStatus
  password: string
}
