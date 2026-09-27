import type { ReactNode } from 'react'
import { APP_COLOR } from '../../constants/config'

type ConfirmModalProps = {
  open: boolean
  title: string
  children: ReactNode
  confirmLabel?: string
  cancelLabel?: string
  loading?: boolean
  tone?: 'danger' | 'primary'
  onClose: () => void
  onConfirm: () => void
}

export function ConfirmModal({
  open,
  title,
  children,
  confirmLabel = 'O‘chirish',
  cancelLabel = 'Bekor qilish',
  loading = false,
  tone = 'danger',
  onClose,
  onConfirm,
}: ConfirmModalProps) {
  if (!open) return null

  return (
    <div className="fixed inset-0 z-50 flex items-center justify-center bg-slate-950/40 px-4">
      <div className="w-full max-w-md rounded-2xl bg-white p-5 shadow-xl">
        <h3 className="text-lg font-bold text-slate-900">{title}</h3>
        <div className="mt-3 text-sm text-slate-600">{children}</div>
        <div className="mt-5 flex justify-end gap-2">
          <button
            type="button"
            onClick={onClose}
            disabled={loading}
            className="rounded-xl border border-slate-200 px-4 py-2 text-sm font-medium text-slate-700 disabled:opacity-60"
          >
            {cancelLabel}
          </button>
          <button
            type="button"
            onClick={onConfirm}
            disabled={loading}
            className={
              tone === 'primary'
                ? 'rounded-xl px-4 py-2 text-sm font-semibold text-white disabled:opacity-60'
                : 'rounded-xl bg-red-600 px-4 py-2 text-sm font-semibold text-white disabled:opacity-60'
            }
            style={tone === 'primary' ? { backgroundColor: APP_COLOR } : undefined}
          >
            {loading ? 'Kutilmoqda...' : confirmLabel}
          </button>
        </div>
      </div>
    </div>
  )
}
