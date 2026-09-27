export type CommentStatus = 'active' | 'inactive'

export type CommentTemplateItem = {
  id: string
  comment: string
  sort_order: number
  status: CommentStatus
  created_at?: string
  updated_at?: string
}

export type CommentTemplateListResult = {
  items: CommentTemplateItem[]
  total: number
  page: number
  limit: number
  total_pages: number
}

export function getCommentTemplateId(item: { id?: string } | null | undefined) {
  return item?.id ?? ''
}
