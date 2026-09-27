import { Plus } from 'lucide-react'
import { FormEvent, useEffect, useState } from 'react'
import { PasswordInput } from '../components/ui/PasswordInput'
import { PhoneInput, displayUzPhone } from '../components/ui/PhoneInput'
import { ImageThumb } from '../components/ui/ImageLightbox'
import { Select } from '../components/ui/Select'
import { TableActions } from '../components/ui/TableActions'
import { ConfirmModal } from '../components/ui/ConfirmModal'
import { auditFields } from '../lib/audit'
import { ViewModal } from '../components/ui/ViewModal'
import { APP_COLOR, API_URL } from '../constants/config'
import { useAuth } from '../context/AuthContext'
import { api, ApiError } from '../lib/api'
import type { AreaItem, LocalShopItem, LocalShopListResult } from '../types/localShop'
import { toast } from '../lib/snack'

type FormState = {
  district_id: string
  mfy_id: string
  name: string
  phone: string
  status: 'active' | 'inactive'
  password: string
  image: string
  clear_image: boolean
}

const empty: FormState = {
  district_id: '',
  mfy_id: '',
  name: '',
  phone: '+998',
  status: 'active',
  password: '',
  image: '',
  clear_image: false,
}

function readImage(file: File): Promise<string> {
  return new Promise((resolve, reject) => {
    if (file.size > 10 * 1024 * 1024) {
      reject(new Error('Rasm 10 MB dan oshmasin'))
      return
    }
    const reader = new FileReader()
    reader.onload = () => resolve(String(reader.result ?? ''))
    reader.onerror = () => reject(new Error('Rasmni o‘qib bo‘lmadi'))
    reader.readAsDataURL(file)
  })
}

export function LocalShopsPage() {
  const { user } = useAuth()
  const isDistrict = user?.type === 'district'
  const [items, setItems] = useState<LocalShopItem[]>([])
  const [districts, setDistricts] = useState<AreaItem[]>([])
  const [mfys, setMfys] = useState<AreaItem[]>([])
  const [open, setOpen] = useState(false)
  const [editing, setEditing] = useState<LocalShopItem | null>(null)
  const [viewing, setViewing] = useState<LocalShopItem | null>(null)
  const [form, setForm] = useState<FormState>(empty)
  const [saving, setSaving] = useState(false)

  const [deleteTarget, setDeleteTarget] = useState<LocalShopItem | null>(null)
  const [deleting, setDeleting] = useState(false)

  async function load() {
    try {
      const [list, districtList] = await Promise.all([
        api<LocalShopListResult>('/api/v1/manager/local-shops'),
        api<AreaItem[]>('/api/v1/manager/districts'),
      ])
      setItems(list.items ?? [])
      setDistricts(districtList)
      if (isDistrict && user?.district_id) {
        setForm((prev) => ({ ...prev, district_id: prev.district_id || user.district_id || '' }))
      }
    } catch (err) {
      if (!(err instanceof ApiError)) toast.error('Yuklab bo‘lmadi')
    }
  }

  useEffect(() => {
    void load()
  }, [])

  useEffect(() => {
    const districtId = isDistrict ? user?.district_id : form.district_id
    if (!districtId) {
      setMfys([])
      return
    }
    const query = isDistrict ? '' : `?district_id=${districtId}`
    void api<AreaItem[]>(`/api/v1/manager/mfys${query}`)
      .then(setMfys)
      .catch(() => setMfys([]))
  }, [form.district_id, isDistrict, user?.district_id])

  function openCreate() {
    setEditing(null)
    setForm({
      ...empty,
      district_id: isDistrict ? user?.district_id ?? '' : '',
    })
    setOpen(true)
  }

  function openEdit(item: LocalShopItem) {
    setEditing(item)
    setForm({
      district_id: item.district_id,
      mfy_id: item.mfy_id,
      name: item.name,
      phone: item.phone,
      status: item.status,
      password: '',
      image: '',
      clear_image: false,
    })
    setOpen(true)
  }

  async function onSubmit(event: FormEvent) {
    event.preventDefault()
    setSaving(true)
    try {
      const body = {
        ...form,
        region_id: user?.region_id,
        district_id: isDistrict ? user?.district_id : form.district_id,
      }
      if (editing) {
        await api(`/api/v1/manager/local-shops/${editing.id}`, { method: 'PUT', body })
      } else {
        await api('/api/v1/manager/local-shops', { method: 'POST', body })
      }
      setOpen(false)
      await load()
    } catch (err) {
      if (!(err instanceof ApiError)) toast.error('Saqlab bo‘lmadi')
    } finally {
      setSaving(false)
    }
  }

  async function confirmDelete() {
    if (!deleteTarget) return
    setDeleting(true)
    try {
      await api(`/api/v1/manager/local-shops/${deleteTarget.id}`, { method: 'DELETE' })
      setDeleteTarget(null)
      await load()
    } catch (err) {
      if (!(err instanceof ApiError)) toast.error('O‘chirib bo‘lmadi')
    } finally {
      setDeleting(false)
    }
  }

  const preview = form.image || (!form.clear_image && editing?.image ? `${API_URL}${editing.image}` : '')

  return (
    <div className="mx-auto max-w-6xl space-y-5">
      <div className="flex flex-col gap-4 sm:flex-row sm:items-start sm:justify-between">
        <div>
          <h2 className="text-2xl font-bold text-slate-900">Mahalla do‘konlari</h2>
          <p className="mt-1 text-sm text-slate-500">
            {isDistrict ? 'Faqat o‘z tumaningizdagi MFYlar' : 'Faqat o‘z viloyatingizdagi tumanlar'}
          </p>
        </div>
        <button
          type="button"
          onClick={openCreate}
          className="inline-flex items-center justify-center gap-2 self-start rounded-xl px-3.5 py-2.5 text-sm font-semibold text-white"
          style={{ backgroundColor: APP_COLOR }}
        >
          <Plus size={16} />
          Qo‘shish
        </button>
      </div>

      <div className="overflow-x-auto rounded-2xl border border-slate-200 bg-white">
        <table className="min-w-full text-left text-sm">
          <thead className="bg-slate-50 text-slate-500">
            <tr>
              <th className="px-4 py-3 font-medium">Do‘kon</th>
              <th className="px-4 py-3 font-medium">Hudud</th>
              <th className="px-4 py-3 font-medium">Telefon</th>
              <th className="px-4 py-3 font-medium">Parol</th>
              <th className="px-4 py-3 text-right font-medium">Amallar</th>
            </tr>
          </thead>
          <tbody>
            {items.length === 0 ? (
              <tr>
                <td colSpan={5} className="px-4 py-10 text-center text-slate-500">
                  Hozircha mahalla do‘konlari yo‘q
                </td>
              </tr>
            ) : (
              items.map((item) => (
                <tr key={item.id} className="border-t border-slate-100">
                  <td className="px-4 py-3 font-medium text-slate-900">
                    <span className="inline-flex items-center gap-2">
                      {item.image ? <ImageThumb src={`${API_URL}${item.image}`} className="h-8 w-8 rounded-lg object-cover" /> : null}
                      <span>{item.name}</span>
                    </span>
                  </td>
                  <td className="px-4 py-3 text-slate-700">
                    {item.district_name} / {item.mfy_name}
                  </td>
                  <td className="px-4 py-3 text-slate-700">{displayUzPhone(item.phone)}</td>
                  <td className="px-4 py-3 text-xs">{item.has_password ? 'O‘rnatilgan' : 'Kutilmoqda'}</td>
                  <td className="px-4 py-3">
                    <TableActions
                      onView={() => setViewing(item)}
                      onEdit={() => openEdit(item)}
                      onDelete={() => setDeleteTarget(item)}
                    />
                  </td>
                </tr>
              ))
            )}
          </tbody>
        </table>
      </div>

      {viewing ? (
        <ViewModal
          title="Mahalla do‘koni"
          fields={[
            { label: 'Nomi', value: viewing.name },
            { label: 'Hudud', value: `${viewing.district_name} / ${viewing.mfy_name}` },
            { label: 'Telefon', value: displayUzPhone(viewing.phone) },
            { label: 'Parol', value: viewing.has_password ? 'O‘rnatilgan' : 'Kutilmoqda' },
            ...auditFields(viewing.audit),
          ]}
          onClose={() => setViewing(null)}
          onEdit={() => {
            openEdit(viewing)
            setViewing(null)
          }}
        />
      ) : null}

      {open ? (
        <div className="fixed inset-0 z-50 flex items-center justify-center bg-slate-950/40 px-4">
          <form onSubmit={onSubmit} className="max-h-[90vh] w-full max-w-lg overflow-y-auto rounded-2xl bg-white p-5">
            <h3 className="text-lg font-bold">{editing ? 'Tahrirlash' : 'Yangi mahalla do‘koni'}</h3>
            <div className="mt-4 grid gap-3 sm:grid-cols-2">
              {!isDistrict ? (
                <Select
                  label="Tuman"
                  className="sm:col-span-2"
                  value={form.district_id}
                  onChange={(v) => setForm((f) => ({ ...f, district_id: v, mfy_id: '' }))}
                  required
                  searchable
                  placeholder="Tanlang"
                  options={districts.map((d) => ({ value: d.id, label: d.name }))}
                />
              ) : null}
              <Select
                label="MFY"
                className="sm:col-span-2"
                value={form.mfy_id}
                onChange={(v) => setForm((f) => ({ ...f, mfy_id: v }))}
                required
                searchable
                placeholder="Tanlang"
                options={mfys.map((m) => ({ value: m.id, label: m.name }))}
              />
              <Field label="Nomi" value={form.name} onChange={(v) => setForm((f) => ({ ...f, name: v }))} />
              <PhoneInput value={form.phone} onChange={(v) => setForm((f) => ({ ...f, phone: v }))} />
              <PasswordInput
                label={editing ? 'Yangi parol (ixtiyoriy)' : 'Parol (ixtiyoriy)'}
                value={form.password}
                onChange={(e) => setForm((f) => ({ ...f, password: e.target.value }))}
              />
              <label className="block text-sm font-medium text-slate-700 sm:col-span-2">
                Rasm (ixtiyoriy)
                <input
                  type="file"
                  accept="image/png,image/jpeg,image/webp"
                  onChange={(e) => {
                    const file = e.target.files?.[0]
                    if (!file) return
                    void readImage(file)
                      .then((image) => setForm((f) => ({ ...f, image, clear_image: false })))
                      .catch((err: Error) => toast.error(err.message))
                  }}
                  className="mt-1.5 w-full text-sm"
                />
              </label>
              {preview ? <ImageThumb src={preview} className="h-16 w-16 rounded-xl object-cover" /> : null}
            </div>
            <div className="mt-5 flex justify-end gap-2">
              <button type="button" onClick={() => setOpen(false)} className="rounded-xl border px-4 py-2 text-sm">
                Bekor
              </button>
              <button type="submit" disabled={saving} className="rounded-xl px-4 py-2 text-sm font-semibold text-white" style={{ backgroundColor: APP_COLOR }}>
                {saving ? 'Saqlanmoqda...' : 'Saqlash'}
              </button>
            </div>
          </form>
        </div>
      ) : null}

      <ConfirmModal
        open={Boolean(deleteTarget)}
        title="O‘chirish"
        confirmLabel="O‘chirish"
        loading={deleting}
        onClose={() => {
          if (!deleting) setDeleteTarget(null)
        }}
        onConfirm={() => {
          void confirmDelete()
        }}
      >
        <strong>{deleteTarget?.name}</strong> o‘chirilsinmi?
      </ConfirmModal>
    </div>
  )
}

function Field({ label, value, onChange }: { label: string; value: string; onChange: (value: string) => void }) {
  return (
    <label className="block text-sm font-medium text-slate-700">
      {label}
      <input value={value} onChange={(e) => onChange(e.target.value)} required className="mt-1.5 w-full rounded-xl border border-slate-200 px-3 py-2.5" />
    </label>
  )
}
