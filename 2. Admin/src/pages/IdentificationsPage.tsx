import { Check, FileText, Loader2, Search, X } from 'lucide-react'
import { toast } from '../lib/snack'
import { motion } from 'motion/react'
import { useCallback, useEffect, useState } from 'react'
import { ConfirmModal } from '../components/ui/ConfirmModal'
import { ImageThumb } from '../components/ui/ImageLightbox'
import { Select } from '../components/ui/Select'
import { TableActions } from '../components/ui/TableActions'
import { APP_COLOR, API_URL } from '../constants/config'
import { auditFields } from '../lib/audit'
import { ApiError } from '../lib/api'
import { approveIdentification, listIdentifications, rejectIdentification } from '../lib/identifications'
import {
  identDocKindLabel,
  identStatusLabel,
  type Identification,
} from '../types/identification'

function fileUrl(path?: string) {
  if (!path) return ''
  return `${API_URL}${path}`
}

export function IdentificationsPage() {
  const [items, setItems] = useState<Identification[]>([])
  const [loading, setLoading] = useState(true)

  const [searchTerm, setSearchTerm] = useState('')
  const [appliedSearch, setAppliedSearch] = useState('')
  const [statusFilter, setStatusFilter] = useState('pending')
  const [page, setPage] = useState(1)
  const [total, setTotal] = useState(0)
  const [viewing, setViewing] = useState<Identification | null>(null)
  const [approveTarget, setApproveTarget] = useState<Identification | null>(null)
  const [rejectTarget, setRejectTarget] = useState<Identification | null>(null)
  const [rejectNote, setRejectNote] = useState('')
  const [reviewing, setReviewing] = useState(false)

  const load = useCallback(async () => {

    try {
      const data = await listIdentifications({
        page,
        limit: 20,
        q: appliedSearch,
        status: statusFilter,
      })
      setItems(data.items ?? [])
      setTotal(data.total)
    } catch (err) {
      if (!(err instanceof ApiError)) toast.error('Yuklab bo‘lmadi')
    } finally {
      setLoading(false)
    }
  }, [page, appliedSearch, statusFilter])

  useEffect(() => {
    setLoading(true)
    void load()
  }, [load])

  async function confirmApprove() {
    if (!approveTarget) return
    setReviewing(true)
    try {
      await approveIdentification(approveTarget.id)
      setApproveTarget(null)
      setViewing(null)
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
      await rejectIdentification(rejectTarget.id, rejectNote.trim())
      setRejectTarget(null)
      setRejectNote('')
      setViewing(null)
      await load()
    } catch (err) {
      if (!(err instanceof ApiError)) toast.error('Bekor qilib bo‘lmadi')
    } finally {
      setReviewing(false)
    }
  }

  return (
    <motion.div initial={{ opacity: 0 }} animate={{ opacity: 1 }} className="mx-auto max-w-6xl space-y-6">
      <div>
        <h2 className="text-2xl font-bold text-slate-900">Identifikatsiya</h2>
        <p className="mt-1 text-sm text-slate-500">Xizmat ko‘rsatuvchi yuborgan ma’lumot va hujjatlarni tasdiqlang yoki sabab yozib bekor qiling</p>
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
              placeholder="Ism, telefon yoki JSHSHIR..."
              className="w-full rounded-xl border border-slate-200 py-2.5 pr-4 pl-10 text-sm outline-none focus:border-slate-400"
            />
          </div>
          <Select
            className="sm:w-52"
            value={statusFilter}
            onChange={(v) => {
              setPage(1)
              setStatusFilter(v)
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
        <p className="rounded-2xl border border-slate-200 bg-white py-12 text-center text-slate-500">Arizalar topilmadi</p>
      ) : (
        <div className="overflow-x-auto rounded-2xl border border-slate-200 bg-white">
          <table className="min-w-full text-left text-sm">
            <thead className="bg-slate-50 text-slate-500">
              <tr>
                <th className="px-4 py-3 font-medium">Xizmat ko‘rsatuvchi</th>
                <th className="px-4 py-3 font-medium">F.I.Sh</th>
                <th className="px-4 py-3 font-medium">Hudud</th>
                <th className="px-4 py-3 font-medium">Holat</th>
                <th className="px-4 py-3 text-right font-medium">Amallar</th>
              </tr>
            </thead>
            <tbody>
              {items.map((item) => (
                <tr key={item.id} className="border-t border-slate-100">
                  <td className="px-4 py-3">
                    <p className="font-medium text-slate-900">{item.provider_name}</p>
                    <p className="text-xs text-slate-500">{item.phone}</p>
                  </td>
                  <td className="px-4 py-3 text-slate-700">{item.full_name}</td>
                  <td className="px-4 py-3 text-slate-700">
                    {[item.region_name, item.district_name, item.mfy_name].filter(Boolean).join(' / ')}
                  </td>
                  <td className="px-4 py-3">
                    <StatusBadge status={item.status} />
                  </td>
                  <td className="px-4 py-3">
                    <TableActions
                      extra={
                        item.status === 'pending' ? (
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

      {viewing ? <IdentView item={viewing} onClose={() => setViewing(null)} onApprove={() => setApproveTarget(viewing)} onReject={() => { setRejectNote(''); setRejectTarget(viewing) }} /> : null}

      <ConfirmModal
        open={Boolean(approveTarget)}
        title="Identifikatsiyani tasdiqlash"
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
            <span className="font-semibold">{approveTarget.full_name || approveTarget.provider_name}</span> tasdiqlansinmi?
          </p>
        ) : null}
      </ConfirmModal>

      {rejectTarget ? (
        <div className="fixed inset-0 z-50 flex items-center justify-center bg-slate-950/40 px-4">
          <div className="w-full max-w-md rounded-2xl bg-white p-5 shadow-xl">
            <h3 className="text-lg font-bold text-slate-900">Identifikatsiyani bekor qilish</h3>
            <p className="mt-2 text-sm text-slate-600">
              <strong>{rejectTarget.full_name || rejectTarget.provider_name}</strong> uchun sabab yozing.
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

function StatusBadge({ status }: { status: Identification['status'] }) {
  const cls =
    status === 'approved'
      ? 'bg-emerald-50 text-emerald-700'
      : status === 'rejected'
        ? 'bg-red-50 text-red-700'
        : 'bg-amber-50 text-amber-700'
  const label = status === 'none' ? 'Topshirilmagan' : identStatusLabel[status]
  return <span className={`rounded-full px-2.5 py-1 text-xs font-semibold ${cls}`}>{label}</span>
}

function IdentView({
  item,
  onClose,
  onApprove,
  onReject,
}: {
  item: Identification
  onClose: () => void
  onApprove: () => void
  onReject: () => void
}) {
  const fields = [
    { label: 'Xizmat ko‘rsatuvchi', value: item.provider_name },
    { label: 'Telefon', value: item.phone },
    { label: 'F.I.Sh', value: item.full_name },
    { label: 'JSHSHIR', value: item.pinfl },
    { label: 'Pasport', value: item.passport },
    { label: 'Tug‘ilgan sana', value: item.birth_date },
    { label: 'Manzil', value: item.address },
    { label: 'Ish tajribasi', value: `${item.experience_years ?? 0} yil` },
    { label: 'Faoliyat', value: item.activity_name },
    { label: 'Hudud', value: [item.region_name, item.district_name, item.mfy_name].filter(Boolean).join(' / ') },
    { label: 'O‘zi haqida', value: item.about },
    ...(item.rejection_note ? [{ label: 'Bekor sababi', value: item.rejection_note }] : []),
    ...auditFields(item.audit),
  ]

  return (
    <div className="fixed inset-0 z-50 flex items-center justify-center bg-slate-950/40 px-4">
      <div className="max-h-[90vh] w-full max-w-2xl overflow-y-auto rounded-2xl bg-white p-5 shadow-xl">
        <div className="flex items-start justify-between gap-3">
          <div>
            <h3 className="text-lg font-bold text-slate-900">{item.full_name || item.provider_name}</h3>
            <div className="mt-2">
              <StatusBadge status={item.status} />
            </div>
          </div>
          <button type="button" onClick={onClose} className="rounded-lg p-1 text-slate-400 hover:bg-slate-100">
            <X size={18} />
          </button>
        </div>
        <dl className="mt-4 grid gap-3 sm:grid-cols-2">
          {fields.map((field) => (
            <div key={field.label} className={field.label === 'O‘zi haqida' || field.label === 'Manzil' ? 'sm:col-span-2' : ''}>
              <dt className="text-xs font-medium text-slate-500">{field.label}</dt>
              <dd className="mt-1 whitespace-pre-line text-sm font-medium text-slate-900">{field.value || '—'}</dd>
            </div>
          ))}
        </dl>
        <div className="mt-4 grid gap-4 sm:grid-cols-2">
          <div>
            <p className="text-xs font-medium text-slate-500">Pasport rasmi</p>
            {item.passport_image ? (
              <ImageThumb src={fileUrl(item.passport_image)} className="mt-2 h-28 w-28 rounded-xl object-cover" />
            ) : (
              <p className="mt-2 text-sm text-slate-400">—</p>
            )}
          </div>
          <div>
            <p className="text-xs font-medium text-slate-500">Selfi</p>
            {item.selfie_image ? (
              <ImageThumb src={fileUrl(item.selfie_image)} className="mt-2 h-28 w-28 rounded-xl object-cover" />
            ) : (
              <p className="mt-2 text-sm text-slate-400">—</p>
            )}
          </div>
        </div>
        <div className="mt-4">
          <p className="text-xs font-medium text-slate-500">Hujjatlar</p>
          <ul className="mt-2 space-y-2">
            {(item.documents ?? []).map((doc) => (
              <li key={doc.id || doc.path}>
                <a
                  href={fileUrl(doc.path)}
                  target="_blank"
                  rel="noreferrer"
                  className="inline-flex items-center gap-2 text-sm font-medium text-sky-800"
                >
                  <FileText size={16} />
                  {identDocKindLabel[doc.kind]} — {doc.title || doc.file_name}
                </a>
              </li>
            ))}
          </ul>
        </div>
        <div className="mt-5 flex justify-end gap-2">
          {item.status === 'pending' ? (
            <>
              <button type="button" onClick={onReject} className="rounded-xl border border-red-200 px-4 py-2 text-sm font-semibold text-red-700">
                Bekor qilish
              </button>
              <button type="button" onClick={onApprove} className="rounded-xl bg-emerald-600 px-4 py-2 text-sm font-semibold text-white">
                Tasdiqlash
              </button>
            </>
          ) : (
            <button type="button" onClick={onClose} className="rounded-xl border border-slate-200 px-4 py-2 text-sm">
              Yopish
            </button>
          )}
        </div>
      </div>
    </div>
  )
}
