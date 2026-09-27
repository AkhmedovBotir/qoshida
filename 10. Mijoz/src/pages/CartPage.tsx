import { ShoppingBag, Trash2 } from 'lucide-react'
import { Link, useNavigate } from 'react-router-dom'
import { KindBadge } from '../components/catalog/KindBadge'
import { QtyStepper } from '../components/catalog/QtyStepper'
import { EmptyState } from '../components/ui/EmptyState'
import { FastImage } from '../components/ui/FastImage'
import { useAuth } from '../context/AuthContext'
import { useCart } from '../context/CartContext'
import { formatMoney, itemPath, mediaSrc, qtyStep } from '../lib/media'

export function CartPage() {
  const { cart, setQuantity } = useCart()
  const { user } = useAuth()
  const navigate = useNavigate()

  if (cart.items.length === 0) {
    return (
      <EmptyState
        icon={<ShoppingBag size={26} />}
        title="Savat bo‘sh"
        text="Katalogdan mahsulot, do‘kon yoki xizmat qo‘shing — keyin shu yerda ko‘rinadi."
        actionTo="/catalog"
        actionLabel="Katalogga o‘tish"
      />
    )
  }

  return (
    <div className="grid items-start gap-5 lg:grid-cols-[minmax(0,1fr)_320px]">
      <section className="space-y-3">
        <div>
          <h2 className="text-xl font-extrabold text-slate-900">Savat</h2>
          <p className="text-sm text-slate-500">{cart.count} ta tur</p>
        </div>
        {cart.items.map((line) => {
          const src = mediaSrc(line.item?.image || line.item?.images?.[0])
          const step = qtyStep(line.item ? { kind: line.kind, unit: line.item.unit } : { kind: line.kind, unit: 'dona' })
          return (
            <article key={line.id} className="market-card flex gap-3 p-3 sm:p-4">
              <Link to={itemPath(line.kind, line.item_id)} className="shrink-0">
                {src ? (
                  <FastImage uri={src} alt={line.item?.name ?? ''} width={88} height={88} radius={16} />
                ) : (
                  <div className="flex h-[88px] w-[88px] items-center justify-center rounded-2xl bg-slate-100 text-[10px] text-slate-400">
                    Rasm
                  </div>
                )}
              </Link>
              <div className="min-w-0 flex-1">
                <KindBadge kind={line.kind} />
                <Link to={itemPath(line.kind, line.item_id)} className="mt-1 block line-clamp-2 font-semibold text-slate-900">
                  {line.item?.name ?? 'Mahsulot'}
                </Link>
                <p className="truncate text-xs text-slate-500">{line.item?.seller_name}</p>
                <p className="price mt-1 text-sm">
                  {formatMoney(line.line_total || (line.item?.price ?? 0) * line.quantity)}
                </p>
                <div className="mt-2 flex items-center gap-2">
                  <QtyStepper
                    size="sm"
                    value={line.quantity}
                    step={step}
                    onChange={(next) => void setQuantity(line.kind, line.item_id, next, line.item)}
                  />
                  <button
                    type="button"
                    className="ml-auto flex h-8 w-8 items-center justify-center rounded-xl text-slate-400 hover:bg-red-50 hover:text-red-600"
                    onClick={() => void setQuantity(line.kind, line.item_id, 0, line.item)}
                    aria-label="O‘chirish"
                  >
                    <Trash2 size={16} />
                  </button>
                </div>
              </div>
            </article>
          )
        })}
      </section>

      <aside className="market-card p-5 lg:sticky lg:top-24">
        <h3 className="font-bold text-slate-900">Jami</h3>
        <div className="mt-3 flex items-center justify-between text-sm">
          <span className="text-slate-500">Mahsulotlar</span>
          <span className="font-semibold">{cart.count} tur</span>
        </div>
        <div className="mt-2 flex items-center justify-between">
          <span className="text-slate-500">To‘lov</span>
          <span className="price text-xl">{formatMoney(cart.total)}</span>
        </div>
        <p className="mt-3 rounded-xl bg-[#E8F0F8] px-3 py-2 text-xs leading-5 text-[#244A73]">
          To‘lov yetkazib berganda naqd (COD). Hozir kartadan yechilmaydi.
        </p>
        <button
          type="button"
          onClick={() => navigate(user ? '/checkout' : `/login?from=${encodeURIComponent('/checkout')}`)}
          className="mt-4 w-full rounded-2xl bg-[#2E5D90] py-3 text-sm font-semibold text-white hover:bg-[#244A73]"
        >
          Buyurtma berish
        </button>
        {!user ? (
          <p className="mt-2 text-center text-xs text-slate-500">Buyurtma uchun avval kiring yoki ro‘yxatdan o‘ting.</p>
        ) : null}
      </aside>
    </div>
  )
}
