import { Banknote, ChevronRight, MapPin, Package, ShieldCheck, Store, Truck } from 'lucide-react'
import { useEffect, useMemo, useState } from 'react'
import { Link, useParams } from 'react-router-dom'
import { AddCartControl } from '../components/catalog/AddCartControl'
import { ItemGallery } from '../components/catalog/ItemGallery'
import { KindBadge } from '../components/catalog/KindBadge'
import { ProductCard } from '../components/catalog/ProductCard'
import { EmptyState } from '../components/ui/EmptyState'
import { Skeleton } from '../components/ui/Skeleton'
import { useCart } from '../context/CartContext'
import { getCatalogItem, listCatalog } from '../lib/market'
import { formatMoney, formatQty, kindMeta, placeLine } from '../lib/media'
import type { CatalogItem, CatalogKind } from '../types/market'

export function ItemPage() {
  const { kind, id } = useParams()
  const { cart } = useCart()
  const [item, setItem] = useState<CatalogItem | null>(null)
  const [related, setRelated] = useState<CatalogItem[]>([])
  const [error, setError] = useState('')
  const [busy, setBusy] = useState(false)

  useEffect(() => {
    if (!kind || !id) return
    setItem(null)
    setRelated([])
    setError('')
    getCatalogItem(kind as CatalogKind, id)
      .then(setItem)
      .catch((err: unknown) => setError(err instanceof Error ? err.message : 'Topilmadi'))
  }, [kind, id])

  useEffect(() => {
    if (!item) return
    let cancelled = false
    listCatalog({ kind: item.kind, category_id: item.category_id, limit: 8 })
      .then((data) => {
        if (cancelled) return
        setRelated((data.items ?? []).filter((row) => !(row.kind === item.kind && row.id === item.id)).slice(0, 8))
      })
      .catch(() => {
        if (!cancelled) setRelated([])
      })
    return () => {
      cancelled = true
    }
  }, [item])

  const images = useMemo(() => {
    if (!item) return []
    const list = [...(item.images ?? [])]
    if (item.image && !list.includes(item.image)) list.unshift(item.image)
    return list
  }, [item])

  if (error) {
    return <EmptyState title="Topilmadi" text={error} actionTo="/catalog" actionLabel="Katalogga qaytish" />
  }

  if (!item) {
    return (
      <div className="grid gap-6 lg:grid-cols-[1.1fr_0.9fr]">
        <Skeleton height={360} radius={24} />
        <div className="space-y-3">
          <Skeleton width="40%" height={16} />
          <Skeleton width="80%" height={28} />
          <Skeleton width="50%" height={22} />
          <Skeleton height={120} radius={20} />
        </div>
      </div>
    )
  }

  const line = cart.items.find((row) => row.kind === item.kind && row.item_id === item.id)
  const meta = kindMeta(item.kind)
  const place = placeLine(item)
  const facts =
    item.kind === 'service'
      ? [
          { icon: ShieldCheck, title: 'Tasdiqlangan xizmat', text: 'Buyurtmadan keyin xizmat ko‘rsatuvchi siz bilan bog‘lanadi.' },
          { icon: Banknote, title: 'Naqd to‘lov', text: 'Xizmat bajarilgach naqd to‘lanadi.' },
        ]
      : item.kind === 'shop'
        ? [
            { icon: Store, title: 'Mahalla do‘koni', text: place || 'Do‘kon manzili buyurtmada ko‘rsatiladi.' },
            { icon: Truck, title: 'Yetkazib berish', text: 'Mahsulot do‘kondan manzilingizga yetkaziladi.' },
            { icon: Banknote, title: 'Naqd (COD)', text: 'To‘lov kuryerga yetkazib berganda.' },
          ]
        : [
            { icon: Package, title: 'Ombordan', text: item.quantity > 0 ? `Mavjud: ${formatQty(item.quantity)} ${item.unit}` : 'Hozircha omborda yo‘q' },
            { icon: Truck, title: 'Yetkazib berish', text: 'Buyurtma tasdiqlangach manzilingizga yuboriladi.' },
            { icon: Banknote, title: 'Naqd (COD)', text: 'To‘lov kuryerga yetkazib berganda.' },
          ]

  function BuyBox({ sticky, controls }: { sticky?: boolean; controls?: boolean }) {
    return (
      <div className={sticky ? 'market-card p-5 lg:sticky lg:top-24' : 'space-y-3'}>
        <KindBadge kind={item.kind} />
        <h1 className="mt-3 text-2xl font-extrabold tracking-tight text-slate-900 sm:text-3xl">{item.name}</h1>
        {item.seller_name ? (
          <p className="mt-1 text-sm text-slate-500">
            {item.kind === 'service' ? 'Xizmat ko‘rsatuvchi' : item.kind === 'shop' ? 'Do‘kon' : 'Sotuvchi'}:{' '}
            <span className="font-semibold text-slate-700">{item.seller_name}</span>
          </p>
        ) : null}
        {item.category_name ? <p className="text-xs text-slate-400">{item.category_name}</p> : null}
        {place ? (
          <p className="flex items-start gap-1.5 text-sm text-slate-500">
            <MapPin size={16} className="mt-0.5 shrink-0" />
            {place}
          </p>
        ) : null}
        <p className="price mt-3 text-3xl">{formatMoney(item.price)}</p>
        {item.kind !== 'service' ? (
          <p className="text-sm text-slate-500">
            {item.quantity > 0 ? `Omborda ${formatQty(item.quantity)} ${item.unit}` : 'Hozircha tugagan'}
          </p>
        ) : (
          <p className="text-sm text-slate-500">Narx 1 {item.unit || 'xizmat'} uchun</p>
        )}

        {controls ? (
          <>
            <AddCartControl item={item} variant="detail" busy={busy} onBusy={setBusy} />
            {line ? (
              <Link to="/cart" className="mt-3 block text-center text-sm font-semibold text-[#2E5D90]">
                Savatni ochish
              </Link>
            ) : null}
          </>
        ) : null}
      </div>
    )
  }

  return (
    <article className="space-y-6 pb-24 lg:pb-0">
      <nav className="flex flex-wrap items-center gap-1 text-xs text-slate-500">
        <Link to="/" className="hover:text-[#2E5D90]">
          Bosh sahifa
        </Link>
        <ChevronRight size={12} />
        <Link to={`/catalog?kind=${item.kind}`} className="hover:text-[#2E5D90]">
          {meta.title}
        </Link>
        <ChevronRight size={12} />
        <span className="line-clamp-1 max-w-[14rem] text-slate-700 sm:max-w-none">{item.name}</span>
      </nav>

      <div className="grid items-start gap-6 lg:grid-cols-[minmax(0,1.15fr)_minmax(280px,0.85fr)]">
        <ItemGallery images={images} alt={item.name} />
        <div className="hidden lg:block">
          <BuyBox sticky controls />
        </div>
        <div className="lg:hidden">
          <BuyBox />
        </div>
      </div>

      <section className="grid gap-3 sm:grid-cols-3">
        {facts.map((fact) => {
          const Icon = fact.icon
          return (
            <div key={fact.title} className="market-card flex gap-3 p-4">
              <span className="flex h-10 w-10 shrink-0 items-center justify-center rounded-xl bg-[#E8F0F8] text-[#2E5D90]">
                <Icon size={18} />
              </span>
              <div>
                <p className="text-sm font-bold text-slate-900">{fact.title}</p>
                <p className="mt-0.5 text-xs leading-5 text-slate-500">{fact.text}</p>
              </div>
            </div>
          )
        })}
      </section>

      {item.description ? (
        <section className="market-card p-5">
          <h2 className="text-base font-bold text-slate-900">Tavsif</h2>
          <p className="mt-2 whitespace-pre-wrap text-sm leading-7 text-slate-600">{item.description}</p>
        </section>
      ) : null}

      {related.length > 0 ? (
        <section>
          <h2 className="mb-3 text-base font-bold text-slate-900">O‘xshash {meta.title.toLowerCase()}</h2>
          <div className="grid grid-cols-2 gap-3 sm:grid-cols-3 lg:grid-cols-4">
            {related.map((row) => (
              <ProductCard key={`${row.kind}-${row.id}`} item={row} />
            ))}
          </div>
        </section>
      ) : null}

      <div className="fixed inset-x-0 bottom-0 z-40 border-t border-[#E6EDF4] bg-white/95 px-4 py-3 pb-[calc(0.75rem+env(safe-area-inset-bottom))] backdrop-blur-md lg:hidden">
        <div className="mx-auto flex max-w-6xl items-center gap-3">
          <div className="min-w-0 flex-1">
            <p className="truncate text-xs text-slate-500">{item.name}</p>
            <p className="price text-lg">{formatMoney(item.price)}</p>
          </div>
          <AddCartControl item={item} variant="bar" busy={busy} onBusy={setBusy} />
          {line ? (
            <Link to="/cart" className="text-sm font-semibold text-[#2E5D90]">
              Savat
            </Link>
          ) : null}
        </div>
      </div>
    </article>
  )
}
