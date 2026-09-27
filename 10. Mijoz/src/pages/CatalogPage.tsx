import { Search } from 'lucide-react'
import { useEffect, useState } from 'react'
import { useSearchParams } from 'react-router-dom'
import { KindChips } from '../components/catalog/KindChips'
import { ProductCard } from '../components/catalog/ProductCard'
import { EmptyState } from '../components/ui/EmptyState'
import { SkeletonCard } from '../components/ui/Skeleton'
import { listCatalog, listCategories } from '../lib/market'
import { kindMeta } from '../lib/media'
import type { CatalogItem, CatalogKind, CategoryItem } from '../types/market'

export function CatalogPage() {
  const [params, setParams] = useSearchParams()
  const q = params.get('q') ?? ''
  const kind = (params.get('kind') ?? '') as CatalogKind | ''
  const categoryId = params.get('category_id') ?? ''
  const [search, setSearch] = useState(q)
  const [categories, setCategories] = useState<CategoryItem[]>([])
  const [items, setItems] = useState<CatalogItem[]>([])
  const [total, setTotal] = useState(0)
  const [page, setPage] = useState(1)
  const [loading, setLoading] = useState(true)

  useEffect(() => {
    setSearch(q)
  }, [q])

  useEffect(() => {
    void listCategories({ roots: true }).then(setCategories).catch(() => setCategories([]))
  }, [])

  useEffect(() => {
    let cancelled = false
    setLoading(page === 1)
    listCatalog({
      q,
      kind,
      category_id: categoryId,
      page,
      limit: 20,
    })
      .then((data) => {
        if (cancelled) return
        setItems((prev) => (page === 1 ? data.items ?? [] : [...prev, ...(data.items ?? [])]))
        setTotal(data.total ?? 0)
      })
      .catch(() => {
        if (!cancelled) {
          setItems([])
          setTotal(0)
        }
      })
      .finally(() => {
        if (!cancelled) setLoading(false)
      })
    return () => {
      cancelled = true
    }
  }, [q, kind, categoryId, page])

  function patch(next: Record<string, string>) {
    const copy = new URLSearchParams(params)
    Object.entries(next).forEach(([key, value]) => {
      if (value) copy.set(key, value)
      else copy.delete(key)
    })
    setPage(1)
    setParams(copy)
  }

  const meta = kindMeta(kind)
  const activeCategory = categories.find((cat) => cat.id === categoryId)

  return (
    <div className="space-y-4">
      <div>
        <h1 className="text-xl font-extrabold text-slate-900 sm:text-2xl">{meta.title}</h1>
        <p className="mt-1 text-sm text-slate-500">{meta.hint}</p>
      </div>

      <form
        onSubmit={(event) => {
          event.preventDefault()
          patch({ q: search.trim() })
        }}
        className="relative"
      >
        <Search size={18} className="pointer-events-none absolute top-1/2 left-3.5 -translate-y-1/2 text-slate-400" />
        <input
          value={search}
          onChange={(e) => setSearch(e.target.value)}
          placeholder="Mahsulot, do‘kon yoki xizmat..."
          className="w-full rounded-2xl border border-[#E6EDF4] bg-white py-3.5 pr-3 pl-11 text-sm outline-none focus:border-[#2E5D90]"
        />
      </form>

      <KindChips value={kind} onChange={(value) => patch({ kind: value })} />

      {categories.length > 0 ? (
        <div className="no-scrollbar flex gap-2 overflow-x-auto pb-1">
          <button
            type="button"
            onClick={() => patch({ category_id: '' })}
            className={`shrink-0 rounded-full px-3 py-1.5 text-xs font-semibold ${
              !categoryId ? 'bg-slate-900 text-white' : 'bg-white text-slate-600 ring-1 ring-[#E6EDF4]'
            }`}
          >
            Barcha turlar
          </button>
          {categories.map((cat) => (
            <button
              key={cat.id}
              type="button"
              onClick={() => patch({ category_id: cat.id })}
              className={`shrink-0 rounded-full px-3 py-1.5 text-xs font-semibold ${
                categoryId === cat.id ? 'bg-slate-900 text-white' : 'bg-white text-slate-600 ring-1 ring-[#E6EDF4]'
              }`}
            >
              {cat.name}
            </button>
          ))}
        </div>
      ) : null}

      <p className="text-xs text-slate-500">
        {total} ta e’lon{activeCategory ? ` · ${activeCategory.name}` : ''}
        {q ? ` · “${q}”` : ''}
      </p>

      <div className="grid grid-cols-2 gap-3 sm:grid-cols-3 lg:grid-cols-4">
        {loading
          ? [1, 2, 3, 4, 5, 6, 7, 8].map((key) => <SkeletonCard key={key} />)
          : items.map((item) => <ProductCard key={`${item.kind}-${item.id}`} item={item} />)}
      </div>
      {!loading && items.length === 0 ? (
        <EmptyState title="Mos e’lon topilmadi" text="Boshqa tur yoki qidiruv so‘zini sinab ko‘ring." />
      ) : null}
      {total > page * 20 ? (
        <button
          type="button"
          onClick={() => setPage((v) => v + 1)}
          className="w-full rounded-2xl bg-white py-3 text-sm font-semibold text-slate-700 ring-1 ring-[#E6EDF4]"
        >
          Yana ko‘rsatish
        </button>
      ) : null}
    </div>
  )
}
