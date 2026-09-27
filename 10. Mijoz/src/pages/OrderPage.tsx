import { ChevronRight } from 'lucide-react'
import { useEffect, useState } from 'react'
import { Link, useParams } from 'react-router-dom'
import { KindBadge } from '../components/catalog/KindBadge'
import { EmptyState } from '../components/ui/EmptyState'
import { FastImage } from '../components/ui/FastImage'
import { Skeleton } from '../components/ui/Skeleton'
import { getOrder } from '../lib/market'
import { formatMoney, itemPath, mediaSrc, statusClass, statusLabel } from '../lib/media'
import type { OrderPublic } from '../types/market'

export function OrderPage() {
  const { id } = useParams()
  const [order, setOrder] = useState<OrderPublic | null>(null)
  const [error, setError] = useState('')

  useEffect(() => {
    if (!id) return
    getOrder(id)
      .then(setOrder)
      .catch((err: unknown) => setError(err instanceof Error ? err.message : 'Topilmadi'))
  }, [id])

  if (error) return <EmptyState title="Topilmadi" text={error} actionTo="/orders" actionLabel="Buyurtmalar" />
  if (!order) {
    return (
      <div className="space-y-3">
        <Skeleton height={32} width="50%" />
        <Skeleton height={140} radius={20} />
      </div>
    )
  }

  return (
    <div className="grid items-start gap-5 lg:grid-cols-[minmax(0,1fr)_300px]">
      <div className="space-y-4">
        <nav className="flex items-center gap-1 text-xs text-slate-500">
          <Link to="/orders" className="hover:text-[#2E5D90]">
            Buyurtmalar
          </Link>
          <ChevronRight size={12} />
          <span>Tafsilot</span>
        </nav>
        <div className="flex items-start justify-between gap-3">
          <div>
            <h2 className="text-xl font-extrabold text-slate-900">Buyurtma</h2>
            <p className="text-xs text-slate-500">{new Date(order.created_at).toLocaleString('uz-UZ')}</p>
          </div>
          <span className={`rounded-full px-2.5 py-1 text-[11px] font-semibold ${statusClass(order.status)}`}>
            {statusLabel(order.status)}
          </span>
        </div>
        {order.items.map((row) => (
          <Link
            key={`${row.kind}-${row.item_id}-${row.name}`}
            to={itemPath(row.kind, row.item_id)}
            className="market-card flex gap-3 p-3"
          >
            {row.image ? <FastImage uri={mediaSrc(row.image)} alt={row.name} width={72} height={72} radius={14} /> : null}
            <div className="min-w-0 flex-1">
              <KindBadge kind={row.kind} />
              <p className="mt-1 font-semibold text-slate-900">{row.name}</p>
              <p className="text-xs text-slate-500">
                {row.quantity} {row.unit} · {formatMoney(row.unit_price)}
              </p>
              <p className="price text-sm">{formatMoney(row.line_total)}</p>
            </div>
          </Link>
        ))}
      </div>
      <aside className="market-card space-y-3 p-5 lg:sticky lg:top-24">
        <p className="text-sm text-slate-500">Yetkazish manzili</p>
        <p className="text-sm font-medium text-slate-900">{order.address}</p>
        {order.note ? <p className="text-sm text-slate-600">{order.note}</p> : null}
        <div className="border-t border-[#E6EDF4] pt-3">
          <p className="text-sm text-slate-500">Jami</p>
          <p className="price text-2xl">{formatMoney(order.total)}</p>
          <p className="mt-1 text-xs text-slate-500">To‘lov: yetkazib berganda naqd</p>
        </div>
      </aside>
    </div>
  )
}
