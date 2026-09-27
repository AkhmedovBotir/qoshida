import type { ReactNode } from 'react'
import { Link } from 'react-router-dom'

export function EmptyState({
  title,
  text,
  actionTo,
  actionLabel,
  icon,
}: {
  title: string
  text: string
  actionTo?: string
  actionLabel?: string
  icon?: ReactNode
}) {
  return (
    <div className="market-card px-6 py-12 text-center">
      {icon ? (
        <div className="mx-auto mb-4 flex h-14 w-14 items-center justify-center rounded-2xl bg-[#E8F0F8] text-[#2E5D90]">
          {icon}
        </div>
      ) : null}
      <h2 className="text-lg font-bold text-slate-900">{title}</h2>
      <p className="mx-auto mt-1 max-w-sm text-sm leading-6 text-slate-500">{text}</p>
      {actionTo && actionLabel ? (
        <Link
          to={actionTo}
          className="mt-5 inline-flex min-h-11 items-center justify-center rounded-2xl bg-[#2E5D90] px-5 text-sm font-semibold text-white hover:bg-[#244A73]"
        >
          {actionLabel}
        </Link>
      ) : null}
    </div>
  )
}
