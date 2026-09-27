import { Loader2, Plus, Search } from 'lucide-react'
import { motion } from 'motion/react'
import { useCallback, useEffect, useState } from 'react'
import { DeliveryFormModal } from '../components/deliveries/DeliveryFormModal'
import { ConfirmModal } from '../components/ui/ConfirmModal'
import { displayUzPhone } from '../components/ui/PhoneInput'
import { TableActions } from '../components/ui/TableActions'
import { auditFields } from '../lib/audit'
import { ViewModal } from '../components/ui/ViewModal'
import { APP_COLOR } from '../constants/config'
import { createDelivery, deleteDelivery, listDeliveries, updateDelivery } from '../lib/deliveries'
import { ApiError } from '../lib/api'
import type { DeliveryForm, DeliveryItem } from '../types/delivery'
import { toast } from '../lib/snack'

export function DeliveriesPage() {
  const [items, setItems] = useState<DeliveryItem[]>([])
  const [loading, setLoading] = useState(true)
  const [searchTerm, setSearchTerm] = useState('')
  const [appliedSearch, setAppliedSearch] = useState('')
  const [page, setPage] = useState(1)
  const [total, setTotal] = useState(0)
  const [formOpen, setFormOpen] = useState(false)
  const [editing, setEditing] = useState<DeliveryItem | null>(null)
  const [viewing, setViewing] = useState<DeliveryItem | null>(null)
  const [saving, setSaving] = useState(false)
  const [deleteTarget, setDeleteTarget] = useState<DeliveryItem | null>(null)
  const [deleting, setDeleting] = useState(false)

  const load = useCallback(async () => {
    try {
      const data = await listDeliveries({ page, limit: 20, q: appliedSearch })
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

  async function handleSubmit(payload: DeliveryForm) {
    setSaving(true)
    try {
      if (editing) await updateDelivery(editing.id, payload)
      else await createDelivery(payload)
      await load()
    } finally {
      setSaving(false)
    }
  }

  async function confirmDelete() {
    if (!deleteTarget) return
    setDeleting(true)
    try {
      await deleteDelivery(deleteTarget.id)
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
          <h2 className="text-2xl font-bold text-slate-900">Yetkazuvchilar</h2>
          <p className="mt-1 text-sm text-slate-500">Istalgan viloyat, tuman, MFY va do‘kon bo‘yicha ro‘yxatga oling</p>
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
            placeholder="Ism, telefon yoki do‘kon..."
            className="w-full rounded-xl border border-slate-200 py-2.5 pr-4 pl-10 text-sm outline-none focus:border-slate-400"
          />
        </div>
      </div>

      {loading ? (
        <div className="flex h-64 items-center justify-center rounded-2xl border border-slate-200 bg-white">
          <Loader2 className="animate-spin" size={28} style={{ color: APP_COLOR }} />
        </div>
      ) : items.length === 0 ? (
        <p className="rounded-2xl border border-slate-200 bg-white py-12 text-center text-slate-500">Yetkazuvchilar topilmadi</p>
      ) : (
        <div className="overflow-x-auto rounded-2xl border border-slate-200 bg-white">
          <table className="min-w-full text-left text-sm">
            <thead className="bg-slate-50 text-slate-500">
              <tr>
                <th className="px-4 py-3 font-medium">Yetkazuvchi</th>
                <th className="px-4 py-3 font-medium">Do‘kon</th>
                <th className="px-4 py-3 font-medium">Hudud</th>
                <th className="px-4 py-3 font-medium">Telefon</th>
                <th className="px-4 py-3 font-medium">Parol</th>
                <th className="px-4 py-3 text-right font-medium">Amallar</th>
              </tr>
            </thead>
            <tbody>
              {items.map((item) => (
                <tr key={item.id} className="border-t border-slate-100">
                  <td className="px-4 py-3 font-medium text-slate-900">
                    {item.first_name} {item.last_name}
                  </td>
                  <td className="px-4 py-3 text-slate-700">{item.shop_name}</td>
                  <td className="px-4 py-3 text-slate-700">
                    {item.region_name} / {item.district_name} / {item.mfy_name}
                  </td>
                  <td className="px-4 py-3 text-slate-700">{displayUzPhone(item.phone)}</td>
                  <td className="px-4 py-3 text-xs font-medium">
                    {item.has_password ? (
                      <span className="text-emerald-700">O‘rnatilgan</span>
                    ) : (
                      <span className="text-amber-700">Kutilmoqda</span>
                    )}
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
          title="Yetkazuvchi"
          fields={[
            { label: 'Ism', value: `${viewing.first_name} ${viewing.last_name}` },
            { label: 'Do‘kon', value: viewing.shop_name },
            { label: 'Hudud', value: `${viewing.region_name} / ${viewing.district_name} / ${viewing.mfy_name}` },
            { label: 'Telefon', value: displayUzPhone(viewing.phone) },
            { label: 'Parol', value: viewing.has_password ? 'O‘rnatilgan' : 'Kutilmoqda' },
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
        <DeliveryFormModal
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
        title="Yetkazuvchini o‘chirish"
        onClose={() => {
          if (!deleting) setDeleteTarget(null)
        }}
        onConfirm={() => {
          void confirmDelete()
        }}
        loading={deleting}
      >
        <strong>
          {deleteTarget?.first_name} {deleteTarget?.last_name}
        </strong>{' '}
        o‘chirilsinmi?
      </ConfirmModal>
    </motion.div>
  )
}
