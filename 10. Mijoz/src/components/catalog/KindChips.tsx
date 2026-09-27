import { cn } from '../../lib/cn'
import type { CatalogKind } from '../../types/market'

export const KIND_OPTIONS: { value: CatalogKind | ''; label: string }[] = [
  { value: '', label: 'Barchasi' },
  { value: 'product', label: 'Mahsulotlar' },
  { value: 'shop', label: 'Do‘konlar' },
  { value: 'service', label: 'Xizmatlar' },
]

export function KindChips({
  value,
  onChange,
}: {
  value: CatalogKind | ''
  onChange: (value: CatalogKind | '') => void
}) {
  return (
    <div className="no-scrollbar flex gap-2 overflow-x-auto pb-1">
      {KIND_OPTIONS.map((item) => (
        <button
          key={item.label}
          type="button"
          onClick={() => onChange(item.value)}
          className={cn(
            'shrink-0 rounded-full px-4 py-2 text-sm font-semibold transition',
            value === item.value ? 'bg-[#2E5D90] text-white shadow-sm' : 'bg-white text-slate-600 ring-1 ring-[#E6EDF4]',
          )}
        >
          {item.label}
        </button>
      ))}
    </div>
  )
}
