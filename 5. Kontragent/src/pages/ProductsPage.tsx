import { Loader2, Plus, Search } from 'lucide-react'
import { motion } from 'motion/react'
import { useCallback, useEffect, useState } from 'react'
import { ProductFormModal } from '../components/products/ProductFormModal'
import { ConfirmModal } from '../components/ui/ConfirmModal'
import { ImageThumb } from '../components/ui/ImageLightbox'
import { Select } from '../components/ui/Select'
import { TableActions } from '../components/ui/TableActions'
import { ViewModal } from '../components/ui/ViewModal'
import { APP_COLOR } from '../constants/config'
import { auditFields } from '../lib/audit'
import { ApiError } from '../lib/api'
import { createProduct, deleteProduct, formatMoney, listProducts, productImageSrc, updateProduct } from '../lib/products'
import type { ProductApproval, ProductForm, ProductItem } from '../types/product'
import { toast } from '../lib/snack'

const approvalLabel: Record<ProductApproval, string> = {
  pending: 'Kutilmoqda',
  approved: 'Tasdiqlangan',
  rejected: 'Bekor qilingan',
}

export function ProductsPage() {
  const [items, setItems] = useState<ProductItem[]>([])
  const [loading, setLoading] = useState(true)
  const [searchTerm, setSearchTerm] = useState('')
  const [appliedSearch, setAppliedSearch] = useState('')
  const [approvalFilter, setApprovalFilter] = useState('')
  const [page, setPage] = useState(1)
  const [total, setTotal] = useState(0)
  const [formOpen, setFormOpen] = useState(false)
  const [editing, setEditing] = useState<ProductItem | null>(null)
  const [viewing, setViewing] = useState<ProductItem | null>(null)
  const [saving, setSaving] = useState(false)
  const [deleteTarget, setDeleteTarget] = useState<ProductItem | null>(null)
  const [deleting, setDeleting] = useState(false)

  const load = useCallback(async () => {
    try {
      const data = await listProducts({ page, limit: 20, q: appliedSearch, approval_status: approvalFilter })
      setItems(data.items ?? [])
      setTotal(data.total)
    } catch (err) {
      if (!(err instanceof ApiError)) toast.error('Yuklab bo‘lmadi')
    } finally {
      setLoading(false)
    }
  }, [page, appliedSearch, approvalFilter])

  useEffect(() => {
    setLoading(true)
    void load()
  }, [load])

  async function handleSubmit(payload: ProductForm) {
    setSaving(true)
    try {
      if (editing) await updateProduct(editing.id, payload)
      else await createProduct(payload)
      await load()
    } finally {
      setSaving(false)
    }
  }

  async function confirmDelete() {
    if (!deleteTarget) return
    setDeleting(true)
    try {
      await deleteProduct(deleteTarget.id)
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
          <h2 className="text-2xl font-bold text-slate-900">Mahsulotlar</h2>
          <p className="mt-1 text-sm text-slate-500">Mahsulot qo‘shilgach yuqori qatlam tasdiqlashi kerak. Bekor qilinsa sabab ko‘rinadi.</p>
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
          Qo‘shish
        </button>
      </div>

      <div className="rounded-2xl border border-slate-200 bg-white p-4">
        <div className="flex flex-col gap-3 sm:flex-row">
          <div className="relative flex-1">
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
              placeholder="Nomi yoki kontragent..."
              className="w-full rounded-xl border border-slate-200 py-2.5 pr-4 pl-10 text-sm outline-none focus:border-slate-400"
            />
          </div>
          <Select
            value={approvalFilter}
            onChange={(v) => {
              setPage(1)
              setApprovalFilter(v)
            }}
            placeholder="Barcha holat"
            options={[
              { value: '', label: 'Barcha holat' },
              { value: 'pending', label: 'Kutilmoqda' },
              { value: 'approved', label: 'Tasdiqlangan' },
              { value: 'rejected', label: 'Bekor qilingan' },
            ]}
          />
        </div>
      </div>

      {loading ? (
        <div className="flex h-64 items-center justify-center rounded-2xl border border-slate-200 bg-white">
          <Loader2 className="animate-spin" size={28} style={{ color: APP_COLOR }} />
        </div>
      ) : items.length === 0 ? (
        <p className="rounded-2xl border border-slate-200 bg-white py-12 text-center text-slate-500">Mahsulotlar topilmadi</p>
      ) : (
        <div className="overflow-x-auto rounded-2xl border border-slate-200 bg-white">
          <table className="min-w-full text-left text-sm">
            <thead className="bg-slate-50 text-slate-500">
              <tr>
                <th className="px-4 py-3 font-medium">Rasm</th>
                <th className="px-4 py-3 font-medium">Mahsulot</th>
                <th className="px-4 py-3 font-medium">Narx</th>
                <th className="px-4 py-3 font-medium">Komissiya</th>
                <th className="px-4 py-3 font-medium">Holat</th>
                <th className="px-4 py-3 text-right font-medium">Amallar</th>
              </tr>
            </thead>
            <tbody>
              {items.map((item) => (
                <tr key={item.id} className="border-t border-slate-100">
                  <td className="px-4 py-3">
                    {item.images?.[0] ? (
                      <ImageThumb
                        src={productImageSrc(item.images[0])}
                        images={item.images.map((src) => productImageSrc(src))}
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
                  <td className="px-4 py-3 text-slate-700">
                    {formatMoney(item.sale_price)} · {item.quantity} {item.unit}
                  </td>
                  <td className="px-4 py-3 text-slate-700">
                    {item.commission_percent}% ({formatMoney(item.commission_amount)})
                  </td>
                  <td className="px-4 py-3">
                    <ApprovalBadge status={item.approval_status} />
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
          title="Mahsulot"
          fields={[
            { label: 'Nomi', value: viewing.name },
            { label: 'Tavsif', value: viewing.description },
            { label: 'Kontragent', value: viewing.kontragent_name },
            { label: 'Kategoriya', value: `${viewing.category_name} / ${viewing.subcategory_name}` },
            { label: 'Sotuv narxi', value: formatMoney(viewing.sale_price) },
            { label: 'Asl narxi', value: formatMoney(viewing.cost_price) },
            { label: 'Miqdor', value: `${viewing.quantity} ${viewing.unit} × ${viewing.unit_size}` },
            { label: 'Komissiya', value: `${viewing.commission_percent}% (${formatMoney(viewing.commission_amount)})` },
            { label: 'Tasdiq', value: approvalLabel[viewing.approval_status] },
            ...(viewing.rejection_note ? [{ label: 'Bekor sababi', value: viewing.rejection_note }] : []),
            {
              label: 'Rasmlar',
              value: viewing.images?.length ? (
                <div className="flex flex-wrap gap-2">
                  {viewing.images.map((src) => (
                    <ImageThumb
                      key={src}
                      src={productImageSrc(src)}
                      images={viewing.images.map((item) => productImageSrc(item))}
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
        <ProductFormModal
          key={editing?.id ?? 'new'}
          item={editing}
          saving={saving}
          lockKontragent
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
        title="Mahsulotni o‘chirish"
        onClose={() => {
          if (!deleting) setDeleteTarget(null)
        }}
        onConfirm={() => {
          void confirmDelete()
        }}
        loading={deleting}
      >
        <strong>{deleteTarget?.name}</strong> o‘chirilsinmi?
      </ConfirmModal>
    </motion.div>
  )
}

function ApprovalBadge({ status }: { status: ProductApproval }) {
  const cls =
    status === 'approved'
      ? 'bg-emerald-50 text-emerald-700'
      : status === 'rejected'
        ? 'bg-red-50 text-red-700'
        : 'bg-amber-50 text-amber-700'
  return <span className={`rounded-full px-2.5 py-1 text-xs font-semibold ${cls}`}>{approvalLabel[status]}</span>
}
