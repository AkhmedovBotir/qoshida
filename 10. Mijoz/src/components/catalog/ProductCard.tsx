import { Link } from 'react-router-dom'
import { FastImage } from '../ui/FastImage'
import { formatMoney, itemPath, mediaSrc } from '../../lib/media'
import type { CatalogItem } from '../../types/market'
import { AddCartControl } from './AddCartControl'
import { KindBadge } from './KindBadge'

export function ProductCard({ item }: { item: CatalogItem }) {
  const src = mediaSrc(item.image || item.images?.[0])
  const href = itemPath(item.kind, item.id)

  return (
    <article className="group flex flex-col overflow-hidden rounded-[1.25rem] border border-[#E6EDF4] bg-white shadow-[0_8px_24px_rgba(18,32,51,0.05)] transition hover:-translate-y-0.5 hover:shadow-[0_14px_32px_rgba(18,32,51,0.08)]">
      <Link to={href} className="relative block">
        {src ? (
          <FastImage uri={src} alt={item.name} height="100%" radius={0} className="aspect-[4/3] w-full" />
        ) : (
          <div className="flex aspect-[4/3] items-center justify-center bg-slate-100 text-xs font-medium text-slate-400">
            Rasm yo‘q
          </div>
        )}
        <KindBadge kind={item.kind} className="absolute top-2.5 left-2.5 shadow-sm" />
        <AddCartControl item={item} variant="card" />
      </Link>
      <div className="flex flex-1 flex-col gap-1.5 p-3.5">
        <Link to={href} className="line-clamp-2 min-h-10 text-sm font-semibold text-slate-900 group-hover:text-[#2E5D90]">
          {item.name}
        </Link>
        <p className="truncate text-xs text-slate-500">{item.seller_name || item.category_name || 'Qoshida'}</p>
        <p className="price mt-auto text-[15px]">{formatMoney(item.price)}</p>
      </div>
    </article>
  )
}
