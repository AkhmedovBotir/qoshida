export type KontragentUser = {
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

export type PhoneCheck = {
  exists: boolean
  active: boolean
  has_password: boolean
  needs_setup: boolean
  message: string
}
