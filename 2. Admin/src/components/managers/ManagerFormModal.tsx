import { FormEvent, useEffect, useState } from 'react'
import { APP_COLOR } from '../../constants/config'
import { listRegions } from '../../lib/regions'
import { ApiError } from '../../lib/api'
import type { ManagerForm, ManagerItem, ManagerStatus, ManagerType } from '../../types/manager'
import type { RegionItem } from '../../types/region'
import { PasswordInput } from '../ui/PasswordInput'
import { PhoneInput } from '../ui/PhoneInput'
import { Select } from '../ui/Select'
import { isPasswordValid } from '../ui/PasswordRules'
import { toast } from '../../lib/snack'

type ManagerFormModalProps = {
  item: ManagerItem | null
  saving: boolean
  onClose: () => void
  onSubmit: (payload: ManagerForm) => Promise<void>
}

const empty: ManagerForm = {
  type: 'region',
  region_id: '',
  district_id: '',
  first_name: '',
  last_name: '',
  phone: '+998',
  username: '',
  status: 'active',
  password: '',
}

export function ManagerFormModal({ item, saving, onClose, onSubmit }: ManagerFormModalProps) {
  const isEdit = Boolean(item)
  const [form, setForm] = useState<ManagerForm>(() =>
    item
      ? {
          type: item.type,
          region_id: item.region_id,
          district_id: item.district_id ?? '',
          first_name: item.first_name,
          last_name: item.last_name,
          phone: item.phone,
          username: item.username,
          status: item.status,
          password: '',
        }
      : empty,
  )
  const [regions, setRegions] = useState<RegionItem[]>([])
  const [districts, setDistricts] = useState<RegionItem[]>([])

  useEffect(() => {
    void listRegions({ type: 'region' }).then(setRegions).catch(() => setRegions([]))
  }, [])

  useEffect(() => {
    if (!form.region_id || form.type !== 'district') {
      setDistricts([])
      return
    }
    void listRegions({ type: 'district', parent_id: form.region_id })
      .then(setDistricts)
      .catch(() => setDistricts([]))
  }, [form.region_id, form.type])

  function setType(type: ManagerType) {
    setForm((prev) => ({ ...prev, type, district_id: type === 'region' ? '' : prev.district_id }))
  }

  async function handleSubmit(event: FormEvent) {
    event.preventDefault()
    if (!form.region_id) {
      toast.error('Viloyatni tanlang')
      return
    }
    if (form.type === 'district' && !form.district_id) {
      toast.error('Tumanni tanlang')
      return
    }
    if (form.password && !isPasswordValid(form.password)) {
      toast.error('Parol qoidalariga rioya qiling')
      return
    }
    try {
      await onSubmit(form)
      onClose()
    } catch (err) {
      if (!(err instanceof ApiError)) toast.error('Saqlab bo‘lmadi')
    }
  }

  return (
    <div className="fixed inset-0 z-50 flex items-center justify-center bg-slate-950/40 px-4">
      <form onSubmit={handleSubmit} className="max-h-[90vh] w-full max-w-xl overflow-y-auto rounded-2xl bg-white p-5 shadow-xl">
        <h3 className="text-lg font-bold text-slate-900">{isEdit ? 'Menejerni tahrirlash' : 'Yangi menejer'}</h3>
        <p className="mt-1 text-sm text-slate-500">
          Avval hudud turini tanlang. Parolni menejer o‘zi SMS orqali o‘rnatadi.
        </p>

        <div className="mt-4 grid gap-3 sm:grid-cols-2">
          <Select
            label="Tur"
            value={form.type}
            onChange={(v) => setType(v as ManagerType)}
            options={[
              { value: 'region', label: 'Viloyat menejeri' },
              { value: 'district', label: 'Tuman menejeri' },
            ]}
          />
          <Select
            label="Holat"
            value={form.status}
            onChange={(v) => setForm((f) => ({ ...f, status: v as ManagerStatus }))}
            options={[
              { value: 'active', label: 'Faol' },
              { value: 'inactive', label: 'Nofaol' },
            ]}
          />
          <Select
            label="Viloyat"
            className="sm:col-span-2"
            value={form.region_id}
            onChange={(v) => setForm((f) => ({ ...f, region_id: v, district_id: '' }))}
            required
            searchable
            placeholder="Viloyatni tanlang"
            options={regions.map((region) => ({ value: region.id, label: region.name }))}
          />
          {form.type === 'district' ? (
            <Select
              label="Tuman"
              className="sm:col-span-2"
              value={form.district_id}
              onChange={(v) => setForm((f) => ({ ...f, district_id: v }))}
              required
              searchable
              placeholder="Tumanni tanlang"
              options={districts.map((district) => ({ value: district.id, label: district.name }))}
            />
          ) : null}
          <Field label="Ism" value={form.first_name} onChange={(v) => setForm((f) => ({ ...f, first_name: v }))} />
          <Field label="Familiya" value={form.last_name} onChange={(v) => setForm((f) => ({ ...f, last_name: v }))} />
          <PhoneInput value={form.phone} onChange={(v) => setForm((f) => ({ ...f, phone: v }))} />
          <Field label="Username" value={form.username} onChange={(v) => setForm((f) => ({ ...f, username: v }))} />
          {isEdit ? (
            <PasswordInput
              label="Parol (ixtiyoriy)"
              className="sm:col-span-2"
              value={form.password}
              onChange={(e) => setForm((f) => ({ ...f, password: e.target.value }))}
              showRules
              autoComplete="new-password"
            />
          ) : null}
        </div>

        <div className="mt-5 flex justify-end gap-2">
          <button type="button" onClick={onClose} className="rounded-xl border border-slate-200 px-4 py-2 text-sm">
            Bekor qilish
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
  )
}

function Field({ label, value, onChange }: { label: string; value: string; onChange: (value: string) => void }) {
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
