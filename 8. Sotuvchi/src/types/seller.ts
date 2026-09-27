export type SellerUser = {
  id: string
  shop_id: string
  first_name: string
  last_name: string
  phone: string
  status: 'active' | 'inactive'
  has_password: boolean
  shop_name?: string
  region_id?: string
  district_id?: string
  mfy_id?: string
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
