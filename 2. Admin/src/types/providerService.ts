export type ServiceApproval = 'pending' | 'approved' | 'rejected'

export type ProviderServiceItem = {
  audit?: import('../lib/audit').AuditTrail
  id: string
  provider_id: string
  name: string
  price: number
  images?: string[]
  image?: string
  approval_status: ServiceApproval
  rejection_note?: string
  submitted_by: 'provider' | 'staff'
  reviewed_by_name?: string
  reviewed_by_role?: string
  reviewed_at?: string
  provider_name?: string
  created_at: string
}

export type ProviderServiceListResult = {
  items: ProviderServiceItem[]
  total: number
  page: number
  limit: number
}

export type ProviderServiceRow = {
  name: string
  price: string
  images: string[]
}

export type ProviderServiceForm = {
  provider_id: string
  items: ProviderServiceRow[]
}
