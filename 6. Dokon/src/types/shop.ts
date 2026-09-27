export type ShopUser = {
  id: string
  name: string
  phone: string
  image: string
  region_id: string
  district_id: string
  mfy_id: string
  status: 'active' | 'inactive'
  has_password: boolean
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
