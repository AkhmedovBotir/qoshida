export type ProductApproval = 'pending' | 'approved' | 'rejected'
export type ProductUnit = 'dona' | 'litr' | 'kg'
export type ProductStatus = 'active' | 'inactive'

export type ProductItem = {
  audit?: import('../lib/audit').AuditTrail
  id: string
  kontragent_id: string
  category_id: string
  subcategory_id: string
  name: string
  description: string
  sale_price: number
  cost_price: number
  quantity: number
  unit: ProductUnit
  unit_size: number
  commission_percent: number
  commission_amount: number
  images: string[]
  approval_status: ProductApproval
  rejection_note?: string
  submitted_by: 'kontragent' | 'staff'
  status: ProductStatus
  reviewed_by_name?: string
  reviewed_by_role?: string
  reviewed_at?: string
  kontragent_name?: string
  category_name?: string
  subcategory_name?: string
  region_name?: string
  district_name?: string
  mfy_name?: string
}

export type ProductListResult = {
  items: ProductItem[]
  total: number
  page: number
  limit: number
}

export type ProductForm = {
  kontragent_id: string
  category_id: string
  subcategory_id: string
  name: string
  description: string
  sale_price: string
  cost_price: string
  quantity: string
  unit: ProductUnit
  unit_size: string
  commission_percent: string
  images: string[]
  status: ProductStatus
}
