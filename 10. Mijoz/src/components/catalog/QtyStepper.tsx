import { Minus, Plus } from 'lucide-react'
import { cn } from '../../lib/cn'
import { formatQty } from '../../lib/media'

export function QtyStepper({
  value,
  step,
  onChange,
  size = 'md',
  disabled,
}: {
  value: number
  step: number
  onChange: (next: number) => void
  size?: 'sm' | 'md'
  disabled?: boolean
}) {
  const compact = size === 'sm'
  return (
    <div
      className={cn(
        'flex items-center',
        compact
          ? 'rounded-full bg-white p-0.5 shadow-[0_8px_20px_rgba(18,32,51,0.16)] ring-1 ring-black/5'
          : 'rounded-2xl bg-[#F2F5F8] p-1',
      )}
    >
      <button
        type="button"
        disabled={disabled}
        className={cn(
          'flex items-center justify-center text-slate-700 disabled:opacity-40',
          compact ? 'h-8 w-8 rounded-full bg-[#F7F4EE]' : 'h-11 w-11 rounded-xl bg-white shadow-sm',
        )}
        onClick={() => onChange(Math.max(0, Math.round((value - step) * 1000) / 1000))}
        aria-label="Kamaytirish"
      >
        <Minus size={compact ? 14 : 16} />
      </button>
      <span className={cn('text-center text-sm font-bold text-slate-900', compact ? 'min-w-8' : 'min-w-16')}>
        {formatQty(value)}
      </span>
      <button
        type="button"
        disabled={disabled}
        className={cn(
          'flex items-center justify-center text-slate-700 disabled:opacity-40',
          compact ? 'h-8 w-8 rounded-full bg-[#F7F4EE]' : 'h-11 w-11 rounded-xl bg-white shadow-sm',
        )}
        onClick={() => onChange(Math.round((value + step) * 1000) / 1000)}
        aria-label="Ko‘paytirish"
      >
        <Plus size={compact ? 14 : 16} />
      </button>
    </div>
  )
}
