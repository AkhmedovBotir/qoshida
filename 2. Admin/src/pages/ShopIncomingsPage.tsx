import { Loader2, Plus, Search } from 'lucide-react'
import { motion } from 'motion/react'
import { useCallback, useEffect, useState } from 'react'
import { IncomingFormModal } from '../components/shop-catalog/IncomingFormModal'
import { TableActions } from '../components/ui/TableActions'
import { ViewModal } from '../components/ui/ViewModal'
import { APP_COLOR } from '../constants/config'
import { auditFields } from '../lib/audit'
import { ApiError } from '../lib/api'
import { createShopIncoming, listShopIncomings } from '../lib/shopCatalog'
import type { ShopIncomingItem } from '../types/shopCatalog'
import { toast } from '../lib/snack'

function formatAt(at?: string) {
  if (!at) return ''
  const date = new Date(at)
  if (Number.isNaN(date.getTime())) return at
  return date.toLocaleString('uz-UZ')
}

export function ShopIncomingsPage() {
  const [items, setItems] = useState<ShopIncomingItem[]>([])
  const [loading, setLoading] = useState(true)
  const [searchTerm, setSearchTerm] = useState('')
  const [appliedSearch, setAppliedSearch] = useState('')
  const [page, setPage] = useState(1)
  const [total, setTotal] = useState(0)
  const [formOpen, setFormOpen] = useState(false)
  const [viewing, setViewing] = useState<ShopIncomingItem | null>(null)
  const [saving, setSaving] = useState(false)

  const load = useCallback(async () => {
    try {
      const data = await listShopIncomings({ page, limit: 20, q: appliedSearch })
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

  async function handleSubmit(payload: { shop_product_id: string; quantity: string }) {
    setSaving(true)
    try {
      await createShopIncoming(payload)
      await load()
    } finally {
      setSaving(false)
    }
  }

  return (
    <motion.div initial={{ opacity: 0 }} animate={{ opacity: 1 }} className="mx-auto max-w-6xl space-y-6">
      <div className="flex flex-col gap-4 sm:flex-row sm:items-start sm:justify-between">
        <div>
          <h2 className="text-2xl font-bold text-slate-900">Kirim</h2>
          <p className="mt-1 text-sm text-slate-500">Do‘kon mahsulotiga kirim qilinganda ombor miqdori oshadi</p>
        </div>
        <button
          type="button"
          onClick={() => setFormOpen(true)}
          className="inline-flex items-center justify-center gap-2 self-start rounded-xl px-3.5 py-2.5 text-sm font-semibold text-white"
          style={{ backgroundColor: APP_COLOR }}
        >
          <Plus size={16} />
          Kirim qilish
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
        <p className="rounded-2xl border border-slate-200 bg-white py-12 text-center text-slate-500">Kirimlar yo‘q</p>
      ) : (
        <div className="overflow-x-auto rounded-2xl border border-slate-200 bg-white">
          <table className="min-w-full text-left text-sm">
            <thead className="bg-slate-50 text-slate-500">
              <tr>
                <th className="px-4 py-3 font-medium">Mahsulot</th>
                <th className="px-4 py-3 font-medium">Do‘kon</th>
                <th className="px-4 py-3 font-medium">Miqdor</th>
                <th className="px-4 py-3 font-medium">Vaqt</th>
                <th className="px-4 py-3 text-right font-medium">Amallar</th>
              </tr>
            </thead>
            <tbody>
              {items.map((item) => (
                <tr key={item.id} className="border-t border-slate-100">
                  <td className="px-4 py-3 font-medium text-slate-900">{item.product_name}</td>
                  <td className="px-4 py-3 text-slate-700">{item.shop_name}</td>
                  <td className="px-4 py-3 text-slate-700">
                    +{item.quantity} {item.unit}
                  </td>
                  <td className="px-4 py-3 text-slate-700">{formatAt(item.created_at)}</td>
                  <td className="px-4 py-3">
                    <TableActions onView={() => setViewing(item)} />
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
          title="Kirim"
          fields={[
            { label: 'Mahsulot', value: viewing.product_name },
            { label: 'Do‘kon', value: viewing.shop_name },
            { label: 'Miqdor', value: `${viewing.quantity} ${viewing.unit ?? ''}` },
            { label: 'Vaqt', value: formatAt(viewing.created_at) },
            ...auditFields(viewing.audit),
          ]}
          onClose={() => setViewing(null)}
        />
      ) : null}

      {formOpen ? (
        <IncomingFormModal
          saving={saving}
          onClose={() => {
            if (!saving) setFormOpen(false)
          }}
          onSubmit={handleSubmit}
        />
      ) : null}
    </motion.div>
  )
}
