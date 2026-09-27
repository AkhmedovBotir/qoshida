import { Eye, Pencil, Trash2 } from 'lucide-react'
import type { ReactNode } from 'react'

type TableActionsProps = {
  onView?: () => void
  onEdit?: () => void
  onDelete?: () => void
  extra?: ReactNode
}

function ActionButton({
  label,
  onClick,
  danger,
  children,
}: {
  label: string
  onClick: () => void
  danger?: boolean
  children: ReactNode
}) {
  return (
    <button
      type="button"
      title={label}
      aria-label={label}
      onClick={onClick}
      className={
        danger
          ? 'rounded-lg border border-red-100 p-2 text-red-600 hover:bg-red-50'
          : 'rounded-lg border border-slate-200 p-2 text-slate-600 hover:bg-slate-50'
      }
    >
      {children}
    </button>
  )
}

export function TableActions({ onView, onEdit, onDelete, extra }: TableActionsProps) {
  return (
    <div className="flex justify-end gap-1.5">
      {extra}
      {onView ? (
        <ActionButton label="Ko‘rish" onClick={onView}>
          <Eye size={15} />
        </ActionButton>
      ) : null}
      {onEdit ? (
        <ActionButton label="Tahrirlash" onClick={onEdit}>
          <Pencil size={15} />
        </ActionButton>
      ) : null}
      {onDelete ? (
        <ActionButton label="O‘chirish" onClick={onDelete} danger>
          <Trash2 size={15} />
        </ActionButton>
      ) : null}
    </div>
  )
}
