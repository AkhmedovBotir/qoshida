import { Plus } from 'lucide-react'
import { FormEvent, useEffect, useState } from 'react'
import { PasswordInput } from '../components/ui/PasswordInput'
import { PhoneInput, displayUzPhone } from '../components/ui/PhoneInput'
import { TableActions } from '../components/ui/TableActions'
import { ConfirmModal } from '../components/ui/ConfirmModal'
import { auditFields } from '../lib/audit'
import { ViewModal } from '../components/ui/ViewModal'
import { APP_COLOR } from '../constants/config'
import { api, ApiError } from '../lib/api'
import type { DeliveryItem, DeliveryListResult } from '../types/delivery'
import { toast } from '../lib/snack'

type FormState = {
  first_name: string
  last_name: string
  phone: string
  status: 'active' | 'inactive'
  password: string
}

const empty: FormState = {
  first_name: '',
  last_name: '',
  phone: '+998',
  status: 'active',
  password: '',
}

export function DeliveriesPage() {
  const [items, setItems] = useState<DeliveryItem[]>([])
  const [open, setOpen] = useState(false)
  const [editing, setEditing] = useState<DeliveryItem | null>(null)
  const [viewing, setViewing] = useState<DeliveryItem | null>(null)
  const [form, setForm] = useState<FormState>(empty)
  const [saving, setSaving] = useState(false)

  const [deleteTarget, setDeleteTarget] = useState<DeliveryItem | null>(null)
  const [deleting, setDeleting] = useState(false)

  async function load() {
    try {
      const list = await api<DeliveryListResult>('/api/v1/local-shop/deliveries')
      setItems(list.items ?? [])
    } catch (err) {
      if (!(err instanceof ApiError)) toast.error('Yuklab bo‘lmadi')
    }
  }

  useEffect(() => {
    void load()
  }, [])

  function openCreate() {
    setEditing(null)
    setForm(empty)
    setOpen(true)
  }

  function openEdit(item: DeliveryItem) {
    setEditing(item)
    setForm({
      first_name: item.first_name,
      last_name: item.last_name,
      phone: item.phone,
      status: item.status,
      password: '',
    })
    setOpen(true)
  }

  async function onSubmit(event: FormEvent) {
    event.preventDefault()
    setSaving(true)
    try {
      if (editing) {
        await api(`/api/v1/local-shop/deliveries/${editing.id}`, { method: 'PUT', body: form })
      } else {
        await api('/api/v1/local-shop/deliveries', { method: 'POST', body: form })
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
      await api(`/api/v1/local-shop/deliveries/${deleteTarget.id}`, { method: 'DELETE' })
      setDeleteTarget(null)
      await load()
    } catch (err) {
      if (!(err instanceof ApiError)) toast.error('O‘chirib bo‘lmadi')
    } finally {
      setDeleting(false)
    }
  }

  return (
    <div className="mx-auto max-w-6xl space-y-5">
      <div className="flex items-center justify-between gap-3">
        <div>
          <h2 className="text-2xl font-bold text-slate-900">Yetkazuvchilar</h2>
          <p className="mt-1 text-sm text-slate-500">Yetkazuvchi avtomatik ravishda shu do‘konga birikadi</p>
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
              <th className="px-4 py-3 font-medium">Yetkazuvchi</th>
              <th className="px-4 py-3 font-medium">Telefon</th>
              <th className="px-4 py-3 font-medium">Parol</th>
              <th className="px-4 py-3 text-right font-medium">Amallar</th>
            </tr>
          </thead>
          <tbody>
            {items.length === 0 ? (
              <tr>
                <td colSpan={4} className="px-4 py-10 text-center text-slate-500">
                  Hozircha yetkazuvchilar yo‘q
                </td>
              </tr>
            ) : (
              items.map((item) => (
                <tr key={item.id} className="border-t border-slate-100">
                  <td className="px-4 py-3 font-medium text-slate-900">
                    {item.first_name} {item.last_name}
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
          title="Yetkazuvchi"
          fields={[
            { label: 'Ism', value: `${viewing.first_name} ${viewing.last_name}` },
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
            <h3 className="text-lg font-bold">{editing ? 'Tahrirlash' : 'Yangi yetkazuvchi'}</h3>
            <p className="mt-1 text-sm text-slate-500">Ism, familiya va telefon kiriting. Parol ixtiyoriy — bo‘sh qoldirilsa, yetkazuvchi SMS orqali o‘rnatadi.</p>
            <div className="mt-4 grid gap-3 sm:grid-cols-2">
              <Field label="Ism" value={form.first_name} onChange={(v) => setForm((f) => ({ ...f, first_name: v }))} />
              <Field label="Familiya" value={form.last_name} onChange={(v) => setForm((f) => ({ ...f, last_name: v }))} />
              <PhoneInput value={form.phone} onChange={(v) => setForm((f) => ({ ...f, phone: v }))} />
              <PasswordInput
                label={editing ? 'Yangi parol (ixtiyoriy)' : 'Parol (ixtiyoriy)'}
                value={form.password}
                onChange={(e) => setForm((f) => ({ ...f, password: e.target.value }))}
              />
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
        {deleteTarget?.first_name} {deleteTarget?.last_name} o‘chirilsinmi?
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
