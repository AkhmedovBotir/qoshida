import { FormEvent, useState } from 'react'
import { APP_COLOR } from '../../constants/config'
import { ApiError } from '../../lib/api'
import type { CommentStatus, CommentTemplateItem } from '../../types/commentTemplate'
import { Select } from '../ui/Select'
import { toast } from '../../lib/snack'

type CommentTemplateFormModalProps = {
  item: CommentTemplateItem | null
  saving: boolean
  onClose: () => void
  onSubmit: (payload: { comment: string; status: CommentStatus }) => Promise<void>
}

export function CommentTemplateFormModal({ item, saving, onClose, onSubmit }: CommentTemplateFormModalProps) {
  const isEdit = Boolean(item)
  const [comment, setComment] = useState(item?.comment ?? '')
  const [status, setStatus] = useState<CommentStatus>(item?.status ?? 'active')
  const [errors, setErrors] = useState<Record<string, string>>({})

  async function handleSubmit(event: FormEvent) {
    event.preventDefault()
    const next: Record<string, string> = {}
    if (!comment.trim()) next.comment = 'Kommentariya matni majburiy'
    setErrors(next)
    if (Object.keys(next).length) return

    try {
      await onSubmit({ comment: comment.trim(), status })
      onClose()
    } catch (err) {
      if (!(err instanceof ApiError)) toast.error('Saqlab bo‘lmadi')
    }
  }

  return (
    <div className="fixed inset-0 z-50 flex items-center justify-center bg-slate-950/40 px-4">
      <form onSubmit={handleSubmit} className="w-full max-w-xl rounded-2xl bg-white p-5 shadow-xl">
        <h3 className="text-lg font-bold text-slate-900">{isEdit ? 'Shablonni tahrirlash' : 'Yangi shablon'}</h3>

        <label className="mt-4 block text-sm font-medium text-slate-700">
          Kommentariya
          <textarea
            value={comment}
            onChange={(e) => setComment(e.target.value)}
            rows={4}
            placeholder="Shablon matnini kiriting"
            className="mt-1.5 w-full rounded-xl border border-slate-200 px-3 py-2.5 outline-none focus:border-slate-400"
          />
          {errors.comment ? <p className="mt-1 text-xs text-red-600">{errors.comment}</p> : null}
        </label>

        <Select
          className="mt-3"
          label="Holat"
          value={status}
          onChange={(v) => setStatus(v as CommentStatus)}
          options={[
            { value: 'active', label: 'Faol' },
            { value: 'inactive', label: 'Nofaol' },
          ]}
        />

        <div className="mt-5 flex justify-end gap-2">
          <button
            type="button"
            disabled={saving}
            onClick={onClose}
            className="rounded-xl border border-slate-200 px-4 py-2 text-sm font-medium text-slate-700 disabled:opacity-60"
          >
            Bekor qilish
          </button>
          <button
            type="submit"
            disabled={saving}
            className="rounded-xl px-4 py-2 text-sm font-semibold text-white disabled:opacity-60"
            style={{ backgroundColor: APP_COLOR }}
          >
            {saving ? 'Saqlanmoqda...' : 'Saqlash'}
          </button>
        </div>
      </form>
    </div>
  )
}
