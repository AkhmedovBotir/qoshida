export type ActivityStatus = 'active' | 'inactive'

export type ActivityTypeItem = {
  id: string
  source_id?: string
  name: string
  icon: string
  status: ActivityStatus
}

export type ActivityTypeListResult = {
  items: ActivityTypeItem[]
  total: number
  page: number
  limit: number
}

export function getActivityTypeId(item: { id?: string } | null | undefined) {
  return item?.id ?? ''
}
