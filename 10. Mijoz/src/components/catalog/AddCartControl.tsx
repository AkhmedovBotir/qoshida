import { Plus, ShoppingBag } from 'lucide-react'
import { cn } from '../../lib/cn'
import { useCart } from '../../context/CartContext'
import { qtyStep } from '../../lib/media'
import type { CatalogItem } from '../../types/market'
import { QtyStepper } from './QtyStepper'

export function AddCartControl({
  item,
  variant = 'card',
  busy,
  onBusy,
}: {
  item: CatalogItem
  variant?: 'card' | 'detail' | 'bar'
  busy?: boolean
  onBusy?: (value: boolean) => void
}) {
  const { add, cart, setQuantity } = useCart()
  const line = cart.items.find((row) => row.kind === item.kind && row.item_id === item.id)
  const step = qtyStep(item)

  async function addOne() {
    onBusy?.(true)
    try {
      await add(item)
    } finally {
      onBusy?.(false)
    }
  }

  if (variant === 'card') {
    if (line) {
      return (
        <div
          className="absolute right-2 bottom-2 z-10"
          onClick={(event) => {
            event.preventDefault()
            event.stopPropagation()
          }}
        >
          <QtyStepper
            size="sm"
            value={line.quantity}
            step={step}
            onChange={(next) => void setQuantity(item.kind, item.id, next, item)}
          />
        </div>
      )
    }
    return (
      <button
        type="button"
        aria-label="Savatga qo‘shish"
        onClick={(event) => {
          event.preventDefault()
          event.stopPropagation()
          void addOne()
        }}
        className="absolute right-2 bottom-2 z-10 flex h-10 w-10 items-center justify-center rounded-full bg-white text-[#C2410C] shadow-[0_8px_20px_rgba(18,32,51,0.18)] ring-1 ring-black/5 transition hover:scale-105 active:scale-95"
      >
        <Plus size={20} strokeWidth={2.4} />
      </button>
    )
  }

  if (line) {
    return (
      <div className={cn('flex items-center gap-3', variant === 'bar' && 'justify-end')}>
        <QtyStepper
          size={variant === 'bar' ? 'sm' : 'md'}
          value={line.quantity}
          step={step}
          onChange={(next) => void setQuantity(item.kind, item.id, next, item)}
        />
      </div>
    )
  }

  return (
    <button
      type="button"
      disabled={busy}
      onClick={() => void addOne()}
      className={cn(
        'inline-flex items-center justify-center gap-2 rounded-full bg-[#F4B942] font-bold text-[#3A2A08] shadow-[0_10px_24px_rgba(244,185,66,0.35)] transition hover:bg-[#efae2e] active:scale-[0.98] disabled:opacity-60',
        variant === 'bar' ? 'min-h-11 px-5 text-sm' : 'mt-3 min-h-12 w-full text-[15px]',
      )}
    >
      <ShoppingBag size={18} />
      {busy ? 'Qo‘shilmoqda...' : 'Savatga'}
    </button>
  )
}
