import { cn } from '../../lib/cn'
import { kindLabel } from '../../lib/media'
import type { CatalogKind } from '../../types/market'

const styles: Record<CatalogKind, string> = {
  product: 'bg-sky-50 text-sky-800',
  shop: 'bg-emerald-50 text-emerald-800',
  service: 'bg-violet-50 text-violet-800',
}

export function KindBadge({ kind, className }: { kind: CatalogKind; className?: string }) {
  return (
    <span
      className={cn(
        'inline-flex rounded-full px-2.5 py-0.5 text-[11px] font-bold tracking-wide',
        styles[kind],
        className,
      )}
    >
      {kindLabel(kind)}
    </span>
  )
}
