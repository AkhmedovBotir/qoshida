import { Plus } from 'lucide-react'
import { motion } from 'motion/react'
import { FormEvent, useEffect, useState } from 'react'
import { PasswordInput } from '../components/ui/PasswordInput'
import { PhoneInput, displayUzPhone } from '../components/ui/PhoneInput'
import { Skeleton } from '../components/ui/Skeleton'
import { TableActions } from '../components/ui/TableActions'
import { ConfirmModal } from '../components/ui/ConfirmModal'
import { ViewModal } from '../components/ui/ViewModal'
import { APP_COLOR } from '../constants/config'
import { ApiError, api } from '../lib/api'
import type { AdminForm, AdminUser } from '../types/admin'
import { toast } from '../lib/snack'

const emptyForm: AdminForm = {
  first_name: '',
  last_name: '',
  phone: '+998',
  username: '',
  password: '',
}

export function AdminsPage() {
  const [items, setItems] = useState<AdminUser[]>([])
  const [loading, setLoading] = useState(true)
  const [open, setOpen] = useState(false)
  const [editing, setEditing] = useState<AdminUser | null>(null)
  const [viewing, setViewing] = useState<AdminUser | null>(null)
  const [form, setForm] = useState<AdminForm>(emptyForm)
  const [saving, setSaving] = useState(false)

  const [deleteTarget, setDeleteTarget] = useState<AdminUser | null>(null)
  const [deleting, setDeleting] = useState(false)

  async function load() {
    setLoading(true)
    try {
      setItems(await api<AdminUser[]>('/api/v1/admins'))
    } finally {
      setLoading(false)
    }
  }

  useEffect(() => {
    void load()
  }, [])

  function openCreate() {
    setEditing(null)
    setForm(emptyForm)
    setOpen(true)
  }

  function openEdit(item: AdminUser) {
    if (item.role === 'general') return
    setEditing(item)
    setForm({
      first_name: item.first_name,
      last_name: item.last_name,
      phone: item.phone,
      username: item.username,
      password: '',
    })
    setOpen(true)
  }

  async function confirmDelete() {
    if (!deleteTarget) return
    setDeleting(true)
    try {
      await api(`/api/v1/admins/${deleteTarget.id}`, { method: 'DELETE' })
      setDeleteTarget(null)
      await load()
    } catch (err) {
      if (!(err instanceof ApiError)) toast.error("O'chirish amalga oshmadi")
    } finally {
      setDeleting(false)
    }
  }

  async function onSubmit(event: FormEvent) {
    event.preventDefault()
    setSaving(true)
    try {
      if (editing) {
        await api(`/api/v1/admins/${editing.id}`, { method: 'PUT', body: form })
      } else {
        await api('/api/v1/admins', { method: 'POST', body: form })
      }
      setOpen(false)
      await load()
    } catch (err) {
      if (!(err instanceof ApiError)) toast.error('Saqlash amalga oshmadi')
    } finally {
      setSaving(false)
    }
  }

  return (
    <motion.div initial={{ opacity: 0, y: 12 }} animate={{ opacity: 1, y: 0 }} className="mx-auto max-w-6xl">
      <div className="flex flex-wrap items-center justify-between gap-3">
        <div>
          <h2 className="text-2xl font-bold text-slate-900">Adminlar</h2>
          <p className="mt-1 text-sm text-slate-500">General adminni o'zgartirish yoki o'chirish mumkin emas.</p>
        </div>
        <button
          type="button"
          onClick={openCreate}
          className="inline-flex items-center justify-center gap-2 self-start rounded-xl px-3.5 py-2.5 text-sm font-semibold text-white"
          style={{ backgroundColor: APP_COLOR }}
        >
          <Plus size={16} />
          Admin qo'shish
        </button>
      </div>

      <div className="mt-5 overflow-x-auto rounded-2xl border border-slate-200 bg-white">
        <table className="min-w-full text-left text-sm">
          <thead className="bg-slate-50 text-slate-500">
            <tr>
              <th className="px-4 py-3 font-medium">Ism</th>
              <th className="px-4 py-3 font-medium">Familiya</th>
              <th className="px-4 py-3 font-medium">Telefon</th>
              <th className="px-4 py-3 font-medium">Username</th>
              <th className="px-4 py-3 font-medium">Role</th>
              <th className="px-4 py-3 font-medium">Amallar</th>
            </tr>
          </thead>
          <tbody>
            {loading
              ? Array.from({ length: 3 }).map((_, index) => (
                  <tr key={index} className="border-t border-slate-100">
                    <td colSpan={6} className="px-4 py-3">
                      <Skeleton height={16} />
                    </td>
                  </tr>
                ))
              : items.map((item) => {
                  const locked = item.role === 'general'
                  return (
                    <tr key={item.id} className="border-t border-slate-100">
                      <td className="px-4 py-3 font-medium text-slate-900">{item.first_name}</td>
                      <td className="px-4 py-3 text-slate-700">{item.last_name}</td>
                      <td className="px-4 py-3 text-slate-700">{displayUzPhone(item.phone)}</td>
                      <td className="px-4 py-3 text-slate-700">{item.username}</td>
                      <td className="px-4 py-3">
                        <span className="rounded-full bg-slate-100 px-2.5 py-1 text-xs font-semibold capitalize text-slate-700">
                          {item.role}
                        </span>
                      </td>
                      <td className="px-4 py-3">
                        <TableActions
                          onView={() => setViewing(item)}
                          onEdit={locked ? undefined : () => openEdit(item)}
                          onDelete={locked ? undefined : () => setDeleteTarget(item)}
                        />
                      </td>
                    </tr>
                  )
                })}
          </tbody>
        </table>
      </div>

      {viewing ? (
        <ViewModal
          title="Admin"
          fields={[
            { label: 'Ism', value: viewing.first_name },
            { label: 'Familiya', value: viewing.last_name },
            { label: 'Telefon', value: displayUzPhone(viewing.phone) },
            { label: 'Username', value: viewing.username },
            { label: 'Role', value: viewing.role },
          ]}
          onClose={() => setViewing(null)}
          onEdit={
            viewing.role === 'general'
              ? undefined
              : () => {
                  const item = viewing
                  setViewing(null)
                  openEdit(item)
                }
          }
        />
      ) : null}

      {open ? (
        <div className="fixed inset-0 z-50 flex items-center justify-center bg-slate-950/40 px-4">
          <form onSubmit={onSubmit} className="w-full max-w-lg rounded-2xl bg-white p-5">
            <h3 className="text-lg font-bold text-slate-900">{editing ? 'Adminni tahrirlash' : "Yangi admin"}</h3>
            <p className="mt-1 text-sm text-slate-500">Role avtomatik `admin` qilinadi.</p>
            <div className="mt-4 grid gap-3 sm:grid-cols-2">
              <Field label="Ism" value={form.first_name} onChange={(v) => setForm({ ...form, first_name: v })} />
              <Field label="Familiya" value={form.last_name} onChange={(v) => setForm({ ...form, last_name: v })} />
              <PhoneInput value={form.phone} onChange={(v) => setForm({ ...form, phone: v })} />
              <Field label="Username" value={form.username} onChange={(v) => setForm({ ...form, username: v })} />
              <PasswordInput
                label={editing ? 'Parol (ixtiyoriy)' : 'Parol'}
                className="sm:col-span-2"
                value={form.password}
                onChange={(e) => setForm({ ...form, password: e.target.value })}
                autoComplete="new-password"
                required={!editing}
                showRules
              />
            </div>
            <div className="mt-5 flex justify-end gap-2">
              <button type="button" onClick={() => setOpen(false)} className="rounded-xl border border-slate-200 px-4 py-2 text-sm">
                Bekor
              </button>
              <button
                type="submit"
                disabled={saving}
                className="rounded-xl px-4 py-2 text-sm font-semibold text-white disabled:opacity-60"
                style={{ backgroundColor: APP_COLOR }}
              >
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
        {deleteTarget?.username} o'chirilsinmi?
      </ConfirmModal>
    </motion.div>
  )
}

function Field({
  label,
  value,
  onChange,
}: {
  label: string
  value: string
  onChange: (value: string) => void
}) {
  return (
    <label className="block text-sm font-medium text-slate-700">
      {label}
      <input
        value={value}
        onChange={(e) => onChange(e.target.value)}
        required
        className="mt-1.5 w-full rounded-xl border border-slate-200 px-3 py-2.5 outline-none focus:border-slate-400"
      />
    </label>
  )
}
