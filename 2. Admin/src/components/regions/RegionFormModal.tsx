import { FormEvent, useState } from 'react'
import { APP_COLOR } from '../../constants/config'
import { ApiError } from '../../lib/api'
import { getRegionId, REGION_TYPES, type RegionItem, type RegionStatus, type UIRegionType } from '../../types/region'
import { Select } from '../ui/Select'
import { toast } from '../../lib/snack'

const TITLES = {
  create: {
    viloyat: 'Yangi viloyat',
    tuman: 'Yangi tuman',
    mfy: 'Yangi MFY',
  },
  edit: {
    viloyat: 'Viloyatni tahrirlash',
    tuman: 'Tumanni tahrirlash',
    mfy: 'MFYni tahrirlash',
  },
}

type FormState = {
  name: string
  code: string
  status: RegionStatus
  viloyat_id: string
  tuman_id: string
}

type RegionFormModalProps = {
  type: UIRegionType
  item: RegionItem | null
  viloyatlar: RegionItem[]
  tumanlar: RegionItem[]
  saving: boolean
  onClose: () => void
  onSubmit: (payload: {
    name: string
    code: string
    status: RegionStatus
    parent_id: string | null
  }) => Promise<void>
}

function emptyForm(): FormState {
  return { name: '', code: '', status: 'active', viloyat_id: '', tuman_id: '' }
}

function itemToForm(item: RegionItem, type: UIRegionType): FormState {
  return {
    name: item.name ?? '',
    code: item.code ?? '',
    status: item.status ?? 'active',
    viloyat_id: type === REGION_TYPES.tuman ? (item.parent_id ?? '') : '',
    tuman_id: type === REGION_TYPES.mfy ? (item.parent_id ?? '') : '',
  }
}

export function RegionFormModal({
  type,
  item,
  viloyatlar,
  tumanlar,
  saving,
  onClose,
  onSubmit,
}: RegionFormModalProps) {
  const isEdit = Boolean(item)
  const [form, setForm] = useState<FormState>(() => (item ? itemToForm(item, type) : emptyForm()))
  const [errors, setErrors] = useState<Record<string, string>>({})

  function handleChange(name: keyof FormState, value: string) {
    setForm((prev) => ({ ...prev, [name]: value }))
    setErrors((prev) => ({ ...prev, [name]: '' }))
  }

  function validate() {
    const next: Record<string, string> = {}
    if (!form.name.trim()) next.name = 'Nomi majburiy'
    if (!form.code.trim()) next.code = 'Kod majburiy'
    if (type === REGION_TYPES.tuman && !form.viloyat_id) next.viloyat_id = 'Viloyatni tanlang'
    if (type === REGION_TYPES.mfy && !form.tuman_id) next.tuman_id = 'Tumanni tanlang'
    setErrors(next)
    return Object.keys(next).length === 0
  }

  async function handleSubmit(event: FormEvent) {
    event.preventDefault()
    if (!validate()) return

    let parent_id: string | null = null
    if (type === REGION_TYPES.tuman) parent_id = form.viloyat_id
    if (type === REGION_TYPES.mfy) parent_id = form.tuman_id

    try {
      await onSubmit({
        name: form.name.trim(),
        code: form.code.trim(),
        status: form.status,
        parent_id,
      })
      onClose()
    } catch (err) {
      if (!(err instanceof ApiError)) toast.error('Saqlab bo‘lmadi')
    }
  }

  const title = isEdit ? TITLES.edit[type] : TITLES.create[type]

  return (
    <div className="fixed inset-0 z-50 flex items-center justify-center bg-slate-950/40 px-4">
      <form onSubmit={handleSubmit} className="w-full max-w-lg rounded-2xl bg-white p-5 shadow-xl">
        <h3 className="text-lg font-bold text-slate-900">{title}</h3>
        <div className="mt-4 space-y-3">
          <label className="block text-sm font-medium text-slate-700">
            Nomi
            <input
              value={form.name}
              onChange={(e) => handleChange('name', e.target.value)}
              placeholder="Nomi kiriting"
              className="mt-1.5 w-full rounded-xl border border-slate-200 px-3 py-2.5 outline-none focus:border-slate-400"
            />
            {errors.name ? <p className="mt-1 text-xs text-red-600">{errors.name}</p> : null}
          </label>

          <label className="block text-sm font-medium text-slate-700">
            Kod
            <input
              value={form.code}
              onChange={(e) => handleChange('code', e.target.value)}
              placeholder="masalan: andijon"
              className="mt-1.5 w-full rounded-xl border border-slate-200 px-3 py-2.5 outline-none focus:border-slate-400"
            />
            {errors.code ? <p className="mt-1 text-xs text-red-600">{errors.code}</p> : null}
          </label>

          {type === REGION_TYPES.tuman ? (
            <div>
              <Select
                label="Viloyat"
                value={form.viloyat_id}
                onChange={(v) => handleChange('viloyat_id', v)}
                required
                searchable
                placeholder="Tanlang"
                options={viloyatlar.map((item) => ({ value: getRegionId(item), label: item.name }))}
              />
              {errors.viloyat_id ? <p className="mt-1 text-xs text-red-600">{errors.viloyat_id}</p> : null}
            </div>
          ) : null}

          {type === REGION_TYPES.mfy ? (
            <div>
              <Select
                label="Tuman"
                value={form.tuman_id}
                onChange={(v) => handleChange('tuman_id', v)}
                required
                searchable
                placeholder="Tanlang"
                options={tumanlar.map((item) => ({
                  value: getRegionId(item),
                  label: item.parent_name ? `${item.name} — ${item.parent_name}` : item.name,
                  hint: item.code,
                }))}
              />
              {errors.tuman_id ? <p className="mt-1 text-xs text-red-600">{errors.tuman_id}</p> : null}
            </div>
          ) : null}

          <Select
            label="Holat"
            value={form.status}
            onChange={(v) => handleChange('status', v)}
            options={[
              { value: 'active', label: 'Faol' },
              { value: 'inactive', label: 'Nofaol' },
            ]}
          />
        </div>

        <div className="mt-5 flex justify-end gap-2">
          <button
            type="button"
            onClick={onClose}
            disabled={saving}
            className="rounded-xl border border-slate-200 px-4 py-2 text-sm font-medium disabled:opacity-60"
          >
            Bekor qilish
          </button>
          <button
            type="submit"
            disabled={saving}
            className="rounded-xl px-4 py-2 text-sm font-semibold text-white disabled:opacity-60"
            style={{ backgroundColor: APP_COLOR }}
          >
            {saving ? 'Saqlanmoqda...' : isEdit ? 'Yangilash' : 'Qo‘shish'}
          </button>
        </div>
      </form>
    </div>
  )
}
