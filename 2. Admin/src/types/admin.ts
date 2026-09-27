export type AdminRole = 'general' | 'admin'

export type AdminUser = {
  id: string
  first_name: string
  last_name: string
  phone: string
  username: string
  role: AdminRole
  is_active: boolean
  created_at: string
}

export type AdminForm = {
  first_name: string
  last_name: string
  phone: string
  username: string
  password: string
}
