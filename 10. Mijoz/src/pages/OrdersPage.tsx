import { ClipboardList } from 'lucide-react'
import { useEffect, useState } from 'react'
import { Link } from 'react-router-dom'
import { EmptyState } from '../components/ui/EmptyState'
import { Skeleton } from '../components/ui/Skeleton'
import { listOrders } from '../lib/market'
import { formatMoney, statusClass, statusLabel } from '../lib/media'
import type { OrderPublic } from '../types/market'

export function OrdersPage() {
  const [items, setItems] = useState<OrderPublic[]>([])
  const [loading, setLoading] = useState(true)

  useEffect(() => {
    listOrders()
      .then((data) => setItems(data.items ?? []))
      .catch(() => setItems([]))
      .finally(() => setLoading(false))
  }, [])

  if (loading) {
    return (
      <div className="space-y-3">
        <Skeleton height={28} width="40%" />
        <Skeleton height={96} radius={20} />
        <Skeleton height={96} radius={20} />
      </div>
    )
  }

  if (items.length === 0) {
    return (
      <EmptyState
        icon={<ClipboardList size={26} />}
        title="Buyurtmalar yo‘q"
        text="Birinchi xaridni katalogdan boshlang. Holat shu yerda kuzatiladi."
        actionTo="/catalog"
        actionLabel="Katalog"
      />
    )
  }

  return (
    <div className="space-y-3">
      <div>
        <h2 className="text-xl font-extrabold text-slate-900">Buyurtmalarim</h2>
        <p className="text-sm text-slate-500">{items.length} ta buyurtma</p>
      </div>
      {items.map((order) => (
        <Link key={order.id} to={`/orders/${order.id}`} className="market-card block p-4 transition hover:-translate-y-0.5">
          <div className="flex items-start justify-between gap-3">
            <div className="min-w-0">
              <p className="price text-base">{formatMoney(order.total)}</p>
              <p className="mt-1 text-xs text-slate-500">{new Date(order.created_at).toLocaleString('uz-UZ')}</p>
              <p className="mt-1 line-clamp-1 text-sm text-slate-600">{order.items.map((row) => row.name).join(', ')}</p>
            </div>
            <span className={`shrink-0 rounded-full px-2.5 py-1 text-[11px] font-semibold ${statusClass(order.status)}`}>
              {statusLabel(order.status)}
            </span>
          </div>
        </Link>
      ))}
    </div>
  )
}
