import { Package, Search, Sparkles, Store } from 'lucide-react'
import { FormEvent, useEffect, useState } from 'react'
import { Link, useNavigate } from 'react-router-dom'
import { ProductCard } from '../components/catalog/ProductCard'
import { EmptyState } from '../components/ui/EmptyState'
import { SkeletonCard } from '../components/ui/Skeleton'
import { listCatalog, listCategories } from '../lib/market'
import { kindMeta, mediaSrc } from '../lib/media'
import type { CatalogItem, CategoryItem } from '../types/market'

const TILES = [
  { kind: 'product' as const, icon: Package, text: 'Ombordan yetkazish' },
  { kind: 'shop' as const, icon: Store, text: 'Mahalla do‘konlari' },
  { kind: 'service' as const, icon: Sparkles, text: 'Usta va xizmatlar' },
]

export function HomePage() {
  const navigate = useNavigate()
  const [query, setQuery] = useState('')
  const [categories, setCategories] = useState<CategoryItem[]>([])
  const [items, setItems] = useState<CatalogItem[]>([])
  const [loading, setLoading] = useState(true)

  useEffect(() => {
    void listCategories({ roots: true }).then(setCategories).catch(() => setCategories([]))
  }, [])

  useEffect(() => {
    let cancelled = false
    setLoading(true)
    listCatalog({ limit: 12 })
      .then((data) => {
        if (!cancelled) setItems(data.items ?? [])
      })
      .catch(() => {
        if (!cancelled) setItems([])
      })
      .finally(() => {
        if (!cancelled) setLoading(false)
      })
    return () => {
      cancelled = true
    }
  }, [])

  function onSearch(event: FormEvent) {
    event.preventDefault()
    const q = query.trim()
    navigate(q ? `/catalog?q=${encodeURIComponent(q)}` : '/catalog')
  }

  return (
    <div className="space-y-6">
      <section className="overflow-hidden rounded-[1.75rem] bg-[#2E5D90] px-5 py-7 text-white shadow-[0_18px_40px_rgba(46,93,144,0.28)] sm:px-8 sm:py-9">
        <p className="text-sm text-white/75">Mahsulot · mahalla do‘koni · xizmat</p>
        <h2 className="mt-1 max-w-lg text-2xl font-extrabold tracking-tight sm:text-3xl">Qoshida marketplace</h2>
        <p className="mt-2 max-w-md text-sm leading-6 text-white/80">
          Tasdiqlangan e’lonlar. To‘lov yetkazib berganda naqd.
        </p>
        <form onSubmit={onSearch} className="relative mt-5 max-w-xl lg:hidden">
          <Search size={18} className="pointer-events-none absolute top-1/2 left-3.5 -translate-y-1/2 text-slate-400" />
          <input
            value={query}
            onChange={(e) => setQuery(e.target.value)}
            placeholder="Qidirish..."
            className="w-full rounded-2xl border-0 bg-white py-3.5 pr-3 pl-11 text-sm text-slate-900 outline-none"
          />
        </form>
      </section>

      <section className="grid grid-cols-1 gap-3 sm:grid-cols-3">
        {TILES.map((tile) => {
          const Icon = tile.icon
          const info = kindMeta(tile.kind)
          return (
            <Link
              key={tile.kind}
              to={`/catalog?kind=${tile.kind}`}
              className="market-card flex items-center gap-3 p-4 transition hover:-translate-y-0.5"
            >
              <span className="flex h-12 w-12 items-center justify-center rounded-2xl bg-[#E8F0F8] text-[#2E5D90]">
                <Icon size={22} />
              </span>
              <span>
                <span className="block font-bold text-slate-900">{info.title}</span>
                <span className="text-xs text-slate-500">{tile.text}</span>
              </span>
            </Link>
          )
        })}
      </section>

      {categories.length > 0 ? (
        <section>
          <div className="mb-3 flex items-center justify-between">
            <h3 className="text-base font-bold text-slate-900">Kategoriyalar</h3>
            <Link to="/catalog" className="text-sm font-semibold text-[#2E5D90]">
              Barchasi
            </Link>
          </div>
          <div className="no-scrollbar flex gap-2 overflow-x-auto pb-1 lg:grid lg:grid-cols-8 lg:overflow-visible">
            {categories.slice(0, 16).map((cat) => (
              <Link
                key={cat.id}
                to={`/catalog?category_id=${cat.id}`}
                className="flex w-[5.5rem] shrink-0 flex-col items-center gap-2 rounded-2xl border border-[#E6EDF4] bg-white p-2.5 lg:w-auto"
              >
                {cat.image_url ? (
                  <img src={mediaSrc(cat.image_url)} alt="" className="h-14 w-14 rounded-xl object-cover" />
                ) : (
                  <div className="flex h-14 w-14 items-center justify-center rounded-xl bg-[#E8F0F8] text-sm font-bold text-[#2E5D90]">
                    {cat.name.slice(0, 1)}
                  </div>
                )}
                <span className="line-clamp-2 text-center text-[11px] font-medium text-slate-700">{cat.name}</span>
              </Link>
            ))}
          </div>
        </section>
      ) : null}

      <section>
        <div className="mb-3 flex items-end justify-between gap-3">
          <div>
            <h3 className="text-base font-bold text-slate-900">Tavsiya etiladi</h3>
            <p className="text-xs text-slate-500">Mahsulot, do‘kon va xizmatlar bir joyda.</p>
          </div>
          <Link to="/catalog" className="text-sm font-semibold text-[#2E5D90]">
            Ko‘rish
          </Link>
        </div>
        <div className="grid grid-cols-2 gap-3 sm:grid-cols-3 lg:grid-cols-4">
          {loading
            ? [1, 2, 3, 4, 5, 6, 7, 8].map((key) => <SkeletonCard key={key} />)
            : items.map((item) => <ProductCard key={`${item.kind}-${item.id}`} item={item} />)}
        </div>
        {!loading && items.length === 0 ? (
          <EmptyState
            title="Hozircha e’lon yo‘q"
            text="Tasdiqlangan mahsulot, do‘kon va xizmatlar shu yerda chiqadi."
            actionTo="/catalog"
            actionLabel="Katalogni ochish"
          />
        ) : null}
      </section>
    </div>
  )
}
