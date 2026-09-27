export type CustomerUser = {
  id: string
  name: string
  first_name?: string
  last_name?: string
  birth_date?: string
  phone: string
  region_id?: string
  district_id?: string
  mfy_id?: string
  address?: string
  avatar?: string
  status: 'active' | 'inactive'
  region_name?: string
  district_name?: string
  mfy_name?: string
  created_at: string
}

export type PhoneCheck = {
  exists: boolean
  active: boolean
  has_password: boolean
  needs_setup: boolean
  can_register: boolean
  message: string
}

export type SendCodeResult = {
  status: string
  message: string
  purpose: string
  exists: boolean
}

export type VerifyResult = {
  status: string
  exists: boolean
  purpose: string
  customer?: CustomerUser
}

export type AuthResult = {
  customer: CustomerUser
}

export type ProfilePayload = {
  first_name: string
  last_name: string
  birth_date: string
  region_id?: string
  district_id?: string
  mfy_id?: string
}
