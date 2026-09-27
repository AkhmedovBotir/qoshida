import type { CommentStatus, CommentTemplateItem, CommentTemplateListResult } from '../types/commentTemplate'
import { api } from './api'

export function listCommentTemplates(params: { page?: number; limit?: number; q?: string } = {}) {
  const query = new URLSearchParams({
    page: String(params.page ?? 1),
    limit: String(params.limit ?? 10),
  })
  if (params.q) query.set('q', params.q)
  return api<CommentTemplateListResult>(`/api/v1/comment-templates?${query}`)
}

export function getCommentTemplate(id: string) {
  return api<CommentTemplateItem>(`/api/v1/comment-templates/${id}`)
}

export function createCommentTemplate(body: { comment: string; status: CommentStatus }) {
  return api<CommentTemplateItem>('/api/v1/comment-templates', { method: 'POST', body })
}

export function updateCommentTemplate(id: string, body: { comment: string; status: CommentStatus }) {
  return api<CommentTemplateItem>(`/api/v1/comment-templates/${id}`, { method: 'PUT', body })
}

export function deleteCommentTemplate(id: string) {
  return api(`/api/v1/comment-templates/${id}`, { method: 'DELETE' })
}

export function reorderCommentTemplates(fromId: string, toId: string) {
  return api('/api/v1/comment-templates/reorder', {
    method: 'PATCH',
    body: { from_id: fromId, to_id: toId },
  })
}
