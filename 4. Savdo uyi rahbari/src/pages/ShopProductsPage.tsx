import { Loader2, Plus, Search } from 'lucide-react'
import { motion } from 'motion/react'
import { useCallback, useEffect, useState } from 'react'
import { AttachProductModal } from '../components/shop-catalog/AttachProductModal'
import { ConfirmModal } from '../components/ui/ConfirmModal'
import { ImageThumb } from '../components/ui/ImageLightbox'
import { TableActions } from '../components/ui/TableActions'
import { ViewModal } from '../components/ui/ViewModal'
import { APP_COLOR } from '../constants/config'
import { auditFields } from '../lib/audit'
import { ApiError } from '../lib/api'
import { attachShopProduct, catalogImageSrc, deleteShopProduct, formatMoney, listShopProducts, updateShopProduct } from '../lib/shopCatalog'
import type { ShopProductForm, ShopProductItem } from '../types/shopCatalog'
import { toast } from '../lib/snack'

export function ShopProductsPage() {
  const [items, setItems] = useState<ShopProductItem[]>([])
  const [loading, setLoading] = useState(true)
  const [searchTerm, setSearchTerm] = useState('')
  const [appliedSearch, setAppliedSearch] = useState('')
  const [page, setPage] = useState(1)
  const [total, setTotal] = useState(0)
  const [formOpen, setFormOpen] = useState(false)
  const [editing, setEditing] = useState<ShopProductItem | null>(null)
  const [viewing, setViewing] = useState<ShopProductItem | null>(null)
  const [saving, setSaving] = useState(false)
  const [deleteTarget, setDeleteTarget] = useState<ShopProductItem | null>(null)
  const [deleting, setDeleting] = useState(false)

  const load = useCallback(async () => {
    try {
      const data = await listShopProducts({ page, limit: 20, q: appliedSearch })
      setItems(data.items ?? [])
      setTotal(data.total)
    } catch (err) {
      if (!(err instanceof ApiError)) toast.error('Yuklab bo‘lmadi')
    } finally {
      setLoading(false)
    }
  }, [page, appliedSearch])

  useEffect(() => {
    setLoading(true)
    void load()
  }, [load])

  async function handleSubmit(payload: ShopProductForm) {
    setSaving(true)
    try {
      if (editing) await updateShopProduct(editing.id, { sale_price: payload.sale_price, cost_price: payload.cost_price })
      else await attachShopProduct(payload)
      await load()
    } finally {
      setSaving(false)
    }
  }

  async function confirmDelete() {
    if (!deleteTarget) return
    setDeleting(true)
    try {
      await deleteShopProduct(deleteTarget.id)
      setDeleteTarget(null)
      await load()
    } catch (err) {
      if (!(err instanceof ApiError)) toast.error('O‘chirib bo‘lmadi')
      setDeleteTarget(null)
    } finally {
      setDeleting(false)
    }
  }

  return (
    <motion.div initial={{ opacity: 0 }} animate={{ opacity: 1 }} className="mx-auto max-w-6xl space-y-6">
      <div className="flex flex-col gap-4 sm:flex-row sm:items-start sm:justify-between">
        <div>
          <h2 className="text-2xl font-bold text-slate-900">Do‘kon mahsulotlari</h2>
          <p className="mt-1 text-sm text-slate-500">Shablondan mahalla do‘koniga mahsulot biriktirish</p>
        </div>
        <button
          type="button"
          onClick={() => {
            setEditing(null)
            setFormOpen(true)
          }}
          className="inline-flex items-center justify-center gap-2 self-start rounded-xl px-3.5 py-2.5 text-sm font-semibold text-white"
          style={{ backgroundColor: APP_COLOR }}
        >
          <Plus size={16} />
          Biriktirish
        </button>
      </div>

      <div className="rounded-2xl border border-slate-200 bg-white p-4">
        <div className="relative">
          <span className="pointer-events-none absolute inset-y-0 left-0 flex w-10 items-center justify-center text-slate-400">
            <Search size={18} />
          </span>
          <input
            value={searchTerm}
            onChange={(e) => setSearchTerm(e.target.value)}
            onKeyDown={(e) => {
              if (e.key === 'Enter') {
                setPage(1)
                setAppliedSearch(searchTerm.trim())
              }
            }}
            placeholder="Mahsulot yoki do‘kon..."
            className="w-full rounded-xl border border-slate-200 py-2.5 pr-4 pl-10 text-sm outline-none focus:border-slate-400"
          />
        </div>
      </div>

      {loading ? (
        <div className="flex h-64 items-center justify-center rounded-2xl border border-slate-200 bg-white">
          <Loader2 className="animate-spin" size={28} style={{ color: APP_COLOR }} />
        </div>
      ) : items.length === 0 ? (
        <p className="rounded-2xl border border-slate-200 bg-white py-12 text-center text-slate-500">Biriktirilgan mahsulot yo‘q</p>
      ) : (
        <div className="overflow-x-auto rounded-2xl border border-slate-200 bg-white">
          <table className="min-w-full text-left text-sm">
            <thead className="bg-slate-50 text-slate-500">
              <tr>
                <th className="px-4 py-3 font-medium">Rasm</th>
                <th className="px-4 py-3 font-medium">Mahsulot</th>
                <th className="px-4 py-3 font-medium">Do‘kon</th>
                <th className="px-4 py-3 font-medium">Miqdor</th>
                <th className="px-4 py-3 font-medium">Narx</th>
                <th className="px-4 py-3 text-right font-medium">Amallar</th>
              </tr>
            </thead>
            <tbody>
              {items.map((item) => (
                <tr key={item.id} className="border-t border-slate-100">
                  <td className="px-4 py-3">
                    {item.images?.[0] ? (
                      <ImageThumb
                        src={catalogImageSrc(item.images[0])}
                        images={item.images.map((src) => catalogImageSrc(src))}
                        className="h-12 w-12 rounded-lg object-cover"
                      />
                    ) : (
                      <span className="block h-12 w-12 rounded-lg bg-slate-100" />
                    )}
                  </td>
                  <td className="px-4 py-3">
                    <p className="font-medium text-slate-900">{item.name}</p>
                    <p className="text-xs text-slate-500">
                      {item.category_name} / {item.subcategory_name}
                    </p>
                  </td>
                  <td className="px-4 py-3 text-slate-700">{item.shop_name}</td>
                  <td className="px-4 py-3 text-slate-700">
                    {item.quantity} {item.unit}
                  </td>
                  <td className="px-4 py-3 text-slate-700">
                    {formatMoney(item.sale_price)} / {formatMoney(item.cost_price)}
                  </td>
                  <td className="px-4 py-3">
                    <TableActions
                      onView={() => setViewing(item)}
                      onEdit={() => {
                        setEditing(item)
                        setFormOpen(true)
                      }}
                      onDelete={() => setDeleteTarget(item)}
                    />
                  </td>
                </tr>
              ))}
            </tbody>
          </table>
        </div>
      )}

      <div className="flex items-center justify-between rounded-2xl border border-slate-200 bg-white px-4 py-3 text-sm">
        <p className="text-slate-600">
          Jami: <span className="font-semibold text-slate-900">{total}</span>
        </p>
        <div className="flex gap-2">
          <button type="button" disabled={page <= 1} onClick={() => setPage((v) => v - 1)} className="rounded-xl border border-slate-200 px-3 py-2 disabled:opacity-50">
            Oldingi
          </button>
          <button type="button" disabled={items.length < 20} onClick={() => setPage((v) => v + 1)} className="rounded-xl border border-slate-200 px-3 py-2 disabled:opacity-50">
            Keyingi
          </button>
        </div>
      </div>

      {viewing ? (
        <ViewModal
          title="Do‘kon mahsuloti"
          fields={[
            { label: 'Nomi', value: viewing.name },
            { label: 'Tavsif', value: viewing.description },
            { label: 'Do‘kon', value: viewing.shop_name },
            { label: 'Kategoriya', value: `${viewing.category_name} / ${viewing.subcategory_name}` },
            { label: 'Miqdor', value: `${viewing.quantity} ${viewing.unit} × ${viewing.unit_size}` },
            { label: 'Sotuv narxi', value: formatMoney(viewing.sale_price) },
            { label: 'Asl narxi', value: formatMoney(viewing.cost_price) },
            {
              label: 'Rasmlar',
              value: viewing.images?.length ? (
                <div className="flex flex-wrap gap-2">
                  {viewing.images.map((src) => (
                    <ImageThumb
                      key={src}
                      src={catalogImageSrc(src)}
                      images={viewing.images.map((item) => catalogImageSrc(item))}
                      className="h-16 w-16 rounded-lg object-cover"
                    />
                  ))}
                </div>
              ) : (
                'Yo‘q'
              ),
            },
            ...auditFields(viewing.audit),
          ]}
          onClose={() => setViewing(null)}
          onEdit={() => {
            setEditing(viewing)
            setViewing(null)
            setFormOpen(true)
          }}
        />
      ) : null}

      {formOpen ? (
        <AttachProductModal
          key={editing?.id ?? 'new'}
          item={editing}
          saving={saving}
          onClose={() => {
            if (!saving) {
              setFormOpen(false)
              setEditing(null)
            }
          }}
          onSubmit={handleSubmit}
        />
      ) : null}

      <ConfirmModal
        open={Boolean(deleteTarget)}
        title="Biriktirishni olib tashlash"
        onClose={() => {
          if (!deleting) setDeleteTarget(null)
        }}
        onConfirm={() => {
          void confirmDelete()
        }}
        loading={deleting}
      >
        <strong>{deleteTarget?.name}</strong> do‘kondan olib tashlansinmi?
      </ConfirmModal>
    </motion.div>
  )
}
