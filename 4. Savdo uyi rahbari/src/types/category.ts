export type CategoryStatus = 'active' | 'inactive'

export type CategoryItem = {
  audit?: import('../lib/audit').AuditTrail
  id: string
  name: string
  slug: string
  parent_id: string | null
  parent_name?: string
  image_url?: string
  censored: boolean
  status: CategoryStatus
}

export type CategoryListResult = {
  items: CategoryItem[]
  total: number
  page: number
  limit: number
}

export function getCategoryId(item: { id?: string } | null | undefined) {
  return item?.id ?? ''
}
