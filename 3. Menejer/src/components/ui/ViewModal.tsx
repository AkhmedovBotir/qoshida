import type { ReactNode } from 'react'
import { APP_COLOR } from '../../constants/config'

export type ViewField = {
  label: string
  value: ReactNode
}

type ViewModalProps = {
  title: string
  fields: ViewField[]
  onClose: () => void
  onEdit?: () => void
}

export function ViewModal({ title, fields, onClose, onEdit }: ViewModalProps) {
  return (
    <div className="fixed inset-0 z-50 flex items-center justify-center bg-slate-950/40 px-4">
      <div className="max-h-[90vh] w-full max-w-lg overflow-y-auto rounded-2xl bg-white p-5 shadow-xl">
        <h3 className="text-lg font-bold text-slate-900">{title}</h3>
        <dl className="mt-4 space-y-3">
          {fields.map((field) => (
            <div key={field.label}>
              <dt className="text-xs font-medium text-slate-500">{field.label}</dt>
              <dd className="mt-1 whitespace-pre-line text-sm font-medium text-slate-900">{field.value || '—'}</dd>
            </div>
          ))}
        </dl>
        <div className="mt-5 flex justify-end gap-2">
          <button type="button" onClick={onClose} className="rounded-xl border border-slate-200 px-4 py-2 text-sm">
            Yopish
          </button>
          {onEdit ? (
            <button
              type="button"
              onClick={onEdit}
              className="rounded-xl px-4 py-2 text-sm font-semibold text-white"
              style={{ backgroundColor: APP_COLOR }}
            >
              Tahrirlash
            </button>
          ) : null}
        </div>
      </div>
    </div>
  )
}
