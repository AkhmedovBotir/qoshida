import { AlertCircle, CheckCircle2, Info, X } from 'lucide-react'
import { useEffect, useState } from 'react'
import { createPortal } from 'react-dom'
import { dismissSnack, subscribeSnacks, type SnackItem } from '../../lib/snack'
import { cn } from '../../lib/cn'

const styles: Record<SnackItem['kind'], string> = {
  success: 'bg-[#122033] text-white',
  error: 'bg-[#B42318] text-white',
  info: 'bg-[#1B3A5C] text-white',
}

function Icon({ kind }: { kind: SnackItem['kind'] }) {
  if (kind === 'success') return <CheckCircle2 size={18} />
  if (kind === 'error') return <AlertCircle size={18} />
  return <Info size={18} />
}

export function SnackbarHost() {
  const [items, setItems] = useState<SnackItem[]>([])

  useEffect(() => subscribeSnacks(setItems), [])

  if (typeof document === 'undefined') return null

  return createPortal(
    <div
      className="pointer-events-none fixed z-[100] flex w-[min(22rem,calc(100vw-1.5rem))] flex-col gap-2"
      style={{
        right: 'max(0.75rem, env(safe-area-inset-right, 0px))',
        bottom: 'calc(var(--snack-offset, 1rem) + env(safe-area-inset-bottom, 0px))',
      }}
    >
      {items.map((item) => (
        <div
          key={item.id}
          className={cn(
            'pointer-events-auto flex items-start gap-2.5 rounded-2xl px-3.5 py-3 text-sm shadow-[0_12px_32px_rgba(18,32,51,0.22)]',
            styles[item.kind],
          )}
          role="status"
        >
          <span className="mt-0.5 shrink-0 opacity-90">
            <Icon kind={item.kind} />
          </span>
          <p className="min-w-0 flex-1 leading-5 font-medium">{item.message}</p>
          <button
            type="button"
            className="shrink-0 rounded-full p-0.5 opacity-70 hover:opacity-100"
            onClick={() => dismissSnack(item.id)}
            aria-label="Yopish"
          >
            <X size={16} />
          </button>
        </div>
      ))}
    </div>,
    document.body,
  )
}
