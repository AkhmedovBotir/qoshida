export type DeliveryItem = {
  audit?: import('../lib/audit').AuditTrail
  id: string
  shop_id: string
  first_name: string
  last_name: string
  phone: string
  status: 'active' | 'inactive'
  has_password: boolean
  shop_name?: string
}

export type DeliveryListResult = {
  items: DeliveryItem[]
  total: number
  page: number
  limit: number
}
