import { Check, Loader2, Plus, Search, X } from 'lucide-react'
import { motion } from 'motion/react'
import { useCallback, useEffect, useState } from 'react'
import { ServiceFormModal } from '../components/services/ServiceFormModal'
import { ConfirmModal } from '../components/ui/ConfirmModal'
import { ImageThumb } from '../components/ui/ImageLightbox'
import { Select } from '../components/ui/Select'
import { TableActions } from '../components/ui/TableActions'
import { ViewModal } from '../components/ui/ViewModal'
import { APP_COLOR } from '../constants/config'
import { auditFields } from '../lib/audit'
import { ApiError } from '../lib/api'
import {
  OWN_SERVICES,
  approveService,
  createServices,
  deleteService,
  formatMoney,
  listProvidersForSelect,
  listServices,
  rejectService,
  serviceImageSrc,
  updateService,
} from '../lib/services'
import type { ProviderServiceForm, ProviderServiceItem, ServiceApproval } from '../types/providerService'
import { toast } from '../lib/snack'

const approvalLabel: Record<ServiceApproval, string> = {
  pending: 'Kutilmoqda',
  approved: 'Tasdiqlangan',
  rejected: 'Bekor qilingan',
}

export function ServicesPage() {
  const [items, setItems] = useState<ProviderServiceItem[]>([])
  const [providers, setProviders] = useState<{ id: string; name: string }[]>([])
  const [loading, setLoading] = useState(true)
  const [searchTerm, setSearchTerm] = useState('')
  const [appliedSearch, setAppliedSearch] = useState('')
  const [providerFilter, setProviderFilter] = useState('')
  const [approvalFilter, setApprovalFilter] = useState('')
  const [page, setPage] = useState(1)
  const [total, setTotal] = useState(0)
  const [formOpen, setFormOpen] = useState(false)
  const [editing, setEditing] = useState<ProviderServiceItem | null>(null)
  const [viewing, setViewing] = useState<ProviderServiceItem | null>(null)
  const [saving, setSaving] = useState(false)
  const [deleteTarget, setDeleteTarget] = useState<ProviderServiceItem | null>(null)
  const [deleting, setDeleting] = useState(false)
  const [approveTarget, setApproveTarget] = useState<ProviderServiceItem | null>(null)
  const [rejectTarget, setRejectTarget] = useState<ProviderServiceItem | null>(null)
  const [rejectNote, setRejectNote] = useState('')
  const [reviewing, setReviewing] = useState(false)

  const load = useCallback(async () => {
    try {
      const data = await listServices({
        page,
        limit: 20,
        q: appliedSearch,
        provider_id: providerFilter,
        approval_status: approvalFilter,
      })
      setItems(data.items ?? [])
      setTotal(data.total)
    } catch (err) {
      if (!(err instanceof ApiError)) toast.error('Yuklab bo‘lmadi')
    } finally {
      setLoading(false)
    }
  }, [page, appliedSearch, providerFilter, approvalFilter])

  useEffect(() => {
    setLoading(true)
    void load()
  }, [load])

  useEffect(() => {
    if (OWN_SERVICES) return
    void listProvidersForSelect()
      .then((data) => setProviders((data.items ?? []).map((item) => ({ id: item.id, name: item.name }))))
      .catch(() => setProviders([]))
  }, [])

  async function handleSubmit(payload: ProviderServiceForm) {
    setSaving(true)
    try {
      if (editing) {
        const row = payload.items[0]
        await updateService(editing.id, row)
      } else {
        await createServices(payload)
      }
      setFormOpen(false)
      setEditing(null)
      await load()
    } finally {
      setSaving(false)
    }
  }

  async function confirmDelete() {
    if (!deleteTarget) return
    setDeleting(true)
    try {
      await deleteService(deleteTarget.id)
      setDeleteTarget(null)
      await load()
    } catch (err) {
      if (!(err instanceof ApiError)) toast.error('O‘chirib bo‘lmadi')
      setDeleteTarget(null)
    } finally {
      setDeleting(false)
    }
  }

  async function confirmApprove() {
    if (!approveTarget) return
    setReviewing(true)
    try {
      await approveService(approveTarget.id)
      setApproveTarget(null)
      await load()
    } catch (err) {
      if (!(err instanceof ApiError)) toast.error('Tasdiqlab bo‘lmadi')
      setApproveTarget(null)
    } finally {
      setReviewing(false)
    }
  }

  async function confirmReject() {
    if (!rejectTarget) return
    if (rejectNote.trim().length < 3) {
      toast.error('Bekor qilish sababi kamida 3 belgi bo‘lishi kerak')
      return
    }
    setReviewing(true)
    try {
      await rejectService(rejectTarget.id, rejectNote.trim())
      setRejectTarget(null)
      setRejectNote('')
      await load()
    } catch (err) {
      if (!(err instanceof ApiError)) toast.error('Bekor qilib bo‘lmadi')
    } finally {
      setReviewing(false)
    }
  }

  return (
    <motion.div initial={{ opacity: 0 }} animate={{ opacity: 1 }} className="mx-auto max-w-6xl space-y-6">
      <div className="flex flex-col gap-4 sm:flex-row sm:items-start sm:justify-between">
        <div>
          <h2 className="text-2xl font-bold text-slate-900">Xizmatlar</h2>
          <p className="mt-1 text-sm text-slate-500">
            {OWN_SERVICES
              ? 'Xizmat qo‘shilgach yuqori qatlam tasdiqlashi kerak. Bekor qilinsa sabab ko‘rinadi.'
              : 'Xizmat ko‘rsatuvchi qo‘shsa tasdiqlang yoki sabab yozib bekor qiling'}
          </p>
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
              placeholder={OWN_SERVICES ? 'Xizmat nomi...' : 'Xizmat yoki ko‘rsatuvchi...'}
              className="w-full rounded-xl border border-slate-200 py-2.5 pr-4 pl-10 text-sm outline-none focus:border-slate-400"
            />
          </div>
          <Select
            className="sm:w-52"
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
          {!OWN_SERVICES ? (
            <Select
              className="sm:w-72"
              value={providerFilter}
              onChange={(v) => {
                setPage(1)
                setProviderFilter(v)
              }}
              placeholder="Barcha ko‘rsatuvchilar"
              searchable
              options={[
                { value: '', label: 'Barcha ko‘rsatuvchilar' },
                ...providers.map((item) => ({ value: item.id, label: item.name })),
              ]}
            />
          ) : null}
        </div>
      </div>

      {loading ? (
        <div className="flex h-64 items-center justify-center rounded-2xl border border-slate-200 bg-white">
          <Loader2 className="animate-spin" size={28} style={{ color: APP_COLOR }} />
        </div>
      ) : items.length === 0 ? (
        <p className="rounded-2xl border border-slate-200 bg-white py-12 text-center text-slate-500">Xizmatlar topilmadi</p>
      ) : (
        <div className="overflow-x-auto rounded-2xl border border-slate-200 bg-white">
          <table className="min-w-full text-left text-sm">
            <thead className="bg-slate-50 text-slate-500">
              <tr>
                <th className="px-4 py-3 font-medium">Rasm</th>
                <th className="px-4 py-3 font-medium">Xizmat</th>
                <th className="px-4 py-3 font-medium">Narxi</th>
                {!OWN_SERVICES ? <th className="px-4 py-3 font-medium">Xizmat ko‘rsatuvchi</th> : null}
                <th className="px-4 py-3 font-medium">Holat</th>
                <th className="px-4 py-3 text-right font-medium">Amallar</th>
              </tr>
            </thead>
            <tbody>
              {items.map((item) => (
                <tr key={item.id} className="border-t border-slate-100">
                  <td className="px-4 py-3">
                    {item.image ? (
                      <ImageThumb
                        src={serviceImageSrc(item.image)}
                        images={(item.images ?? []).map((src) => serviceImageSrc(src))}
                        className="h-12 w-12 rounded-lg object-cover"
                      />
                    ) : (
                      <span className="text-xs text-slate-400">Yo‘q</span>
                    )}
                  </td>
                  <td className="px-4 py-3 font-medium text-slate-900">{item.name}</td>
                  <td className="px-4 py-3 text-slate-700">{formatMoney(item.price)}</td>
                  {!OWN_SERVICES ? <td className="px-4 py-3 text-slate-700">{item.provider_name}</td> : null}
                  <td className="px-4 py-3">
                    <ApprovalBadge status={item.approval_status} />
                  </td>
                  <td className="px-4 py-3">
                    <TableActions
                      extra={
                        !OWN_SERVICES && item.approval_status === 'pending' ? (
                          <>
                            <button
                              type="button"
                              title="Tasdiqlash"
                              onClick={() => setApproveTarget(item)}
                              className="rounded-lg border border-emerald-100 p-2 text-emerald-700 hover:bg-emerald-50"
                            >
                              <Check size={15} />
                            </button>
                            <button
                              type="button"
                              title="Bekor qilish"
                              onClick={() => {
                                setRejectNote('')
                                setRejectTarget(item)
                              }}
                              className="rounded-lg border border-red-100 p-2 text-red-600 hover:bg-red-50"
                            >
                              <X size={15} />
                            </button>
                          </>
                        ) : null
                      }
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

      <div className="flex flex-col gap-3 rounded-2xl border border-slate-200 bg-white px-4 py-3 text-sm sm:flex-row sm:items-center sm:justify-between">
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
          title={viewing.name}
          fields={[
            { label: 'Xizmat', value: viewing.name },
            { label: 'Narxi', value: formatMoney(viewing.price) },
            ...(viewing.images?.length
              ? [
                  {
                    label: 'Rasmlar',
                    value: (
                      <div className="flex flex-wrap gap-2">
                        {viewing.images.map((src) => (
                          <ImageThumb
                            key={src}
                            src={serviceImageSrc(src)}
                            images={viewing.images!.map((itemSrc) => serviceImageSrc(itemSrc))}
                            className="h-16 w-16 rounded-lg object-cover"
                          />
                        ))}
                      </div>
                    ),
                  },
                ]
              : []),
            ...(!OWN_SERVICES ? [{ label: 'Xizmat ko‘rsatuvchi', value: viewing.provider_name ?? '' }] : []),
            { label: 'Holat', value: approvalLabel[viewing.approval_status] },
            ...(viewing.rejection_note ? [{ label: 'Bekor sababi', value: viewing.rejection_note }] : []),
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
        <ServiceFormModal
          item={editing}
          providers={providers}
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
        title="Xizmatni o‘chirish"
        loading={deleting}
        onClose={() => {
          if (!deleting) setDeleteTarget(null)
        }}
        onConfirm={() => {
          void confirmDelete()
        }}
      >
        {deleteTarget ? (
          <p>
            <span className="font-semibold">{deleteTarget.name}</span> o‘chirilsinmi?
          </p>
        ) : null}
      </ConfirmModal>

      {!OWN_SERVICES ? (
        <ConfirmModal
          open={Boolean(approveTarget)}
          title="Xizmatni tasdiqlash"
          confirmLabel="Tasdiqlash"
          loading={reviewing}
          onClose={() => {
            if (!reviewing) setApproveTarget(null)
          }}
          onConfirm={() => {
            void confirmApprove()
          }}
        >
          {approveTarget ? (
            <p>
              <span className="font-semibold">{approveTarget.name}</span> tasdiqlansinmi?
            </p>
          ) : null}
        </ConfirmModal>
      ) : null}

      {!OWN_SERVICES && rejectTarget ? (
        <div className="fixed inset-0 z-50 flex items-center justify-center bg-slate-950/40 px-4">
          <div className="w-full max-w-md rounded-2xl bg-white p-5 shadow-xl">
            <h3 className="text-lg font-bold text-slate-900">Xizmatni bekor qilish</h3>
            <p className="mt-2 text-sm text-slate-600">
              <strong>{rejectTarget.name}</strong> uchun sabab yozing.
            </p>
            <textarea
              value={rejectNote}
              onChange={(e) => setRejectNote(e.target.value)}
              rows={4}
              className="mt-3 w-full rounded-xl border border-slate-200 px-3 py-2.5 text-sm outline-none focus:border-slate-400"
              placeholder="Bekor qilish sababi"
            />
            <div className="mt-5 flex justify-end gap-2">
              <button type="button" disabled={reviewing} onClick={() => setRejectTarget(null)} className="rounded-xl border border-slate-200 px-4 py-2 text-sm">
                Yopish
              </button>
              <button
                type="button"
                disabled={reviewing}
                onClick={() => {
                  void confirmReject()
                }}
                className="rounded-xl bg-red-600 px-4 py-2 text-sm font-semibold text-white disabled:opacity-60"
              >
                {reviewing ? 'Saqlanmoqda...' : 'Bekor qilish'}
              </button>
            </div>
          </div>
        </div>
      ) : null}
    </motion.div>
  )
}

function ApprovalBadge({ status }: { status: ServiceApproval }) {
  const cls =
    status === 'approved'
      ? 'bg-emerald-50 text-emerald-700'
      : status === 'rejected'
        ? 'bg-red-50 text-red-700'
        : 'bg-amber-50 text-amber-700'
  return <span className={`rounded-full px-2.5 py-1 text-xs font-semibold ${cls}`}>{approvalLabel[status]}</span>
}
