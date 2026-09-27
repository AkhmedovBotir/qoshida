export type ApiRegionType = 'region' | 'district' | 'mfy'
export type UIRegionType = 'viloyat' | 'tuman' | 'mfy'
export type RegionStatus = 'active' | 'inactive'

export type RegionItem = {
  id: string
  name: string
  type: ApiRegionType
  parent_id: string | null
  parent_name?: string
  code: string
  status: RegionStatus
}

export type RegionListResult = {
  items: RegionItem[]
  total: number
  page: number
  limit: number
}

export const REGION_TYPES = {
  viloyat: 'viloyat',
  tuman: 'tuman',
  mfy: 'mfy',
} as const

export const API_TYPE: Record<UIRegionType, ApiRegionType> = {
  viloyat: 'region',
  tuman: 'district',
  mfy: 'mfy',
}

export function getRegionId(item: { id?: string } | null | undefined) {
  return item?.id ?? ''
}
