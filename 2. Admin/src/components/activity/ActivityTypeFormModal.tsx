import { FormEvent, useMemo, useState } from 'react'
import { APP_COLOR } from '../../constants/config'
import { ACTIVITY_ICON_OPTIONS, resolveActivityIconKey } from '../../lib/activityIcons'
import { ApiError } from '../../lib/api'
import { type ActivityStatus, type ActivityTypeItem } from '../../types/activityType'
import { Select } from '../ui/Select'
import { toast } from '../../lib/snack'

type ActivityTypeFormModalProps = {
  item: ActivityTypeItem | null
  saving: boolean
  onClose: () => void
  onSubmit: (payload: { name: string; icon: string; status: ActivityStatus }) => Promise<void>
}

export function ActivityTypeFormModal({ item, saving, onClose, onSubmit }: ActivityTypeFormModalProps) {
  const isEdit = Boolean(item)
  const [name, setName] = useState(item?.name ?? '')
  const [icon, setIcon] = useState(resolveActivityIconKey(item?.icon))
  const [iconQuery, setIconQuery] = useState('')
  const [status, setStatus] = useState<ActivityStatus>(item?.status ?? 'active')
  const [errors, setErrors] = useState<Record<string, string>>({})

  const icons = useMemo(() => {
    const q = iconQuery.trim().toLowerCase()
    if (!q) return ACTIVITY_ICON_OPTIONS
    return ACTIVITY_ICON_OPTIONS.filter(
      (option) => option.label.toLowerCase().includes(q) || option.key.toLowerCase().includes(q),
    )
  }, [iconQuery])

  async function handleSubmit(event: FormEvent) {
    event.preventDefault()
    const next: Record<string, string> = {}
    if (!name.trim()) next.name = 'Nomi majburiy'
    if (!icon) next.icon = 'Ikonkani tanlang'
    setErrors(next)
    if (Object.keys(next).length) return

    try {
      await onSubmit({ name: name.trim(), icon, status })
      onClose()
    } catch (err) {
      if (!(err instanceof ApiError)) toast.error('Saqlab bo‘lmadi')
    }
  }

  return (
    <div className="fixed inset-0 z-50 flex items-center justify-center bg-slate-950/40 px-4">
      <form onSubmit={handleSubmit} className="flex max-h-[90vh] w-full max-w-2xl flex-col rounded-2xl bg-white p-5 shadow-xl">
        <h3 className="text-lg font-bold text-slate-900">
          {isEdit ? 'Faoliyat turini tahrirlash' : 'Yangi faoliyat turi'}
        </h3>
        <div className="mt-4 min-h-0 flex-1 space-y-3 overflow-y-auto pr-1">
          <label className="block text-sm font-medium text-slate-700">
            Nomi
            <input
              value={name}
              onChange={(e) => setName(e.target.value)}
              placeholder="Nomi kiriting"
              className="mt-1.5 w-full rounded-xl border border-slate-200 px-3 py-2.5 outline-none focus:border-slate-400"
            />
            {errors.name ? <p className="mt-1 text-xs text-red-600">{errors.name}</p> : null}
          </label>

          <div>
            <div className="flex items-center justify-between gap-3">
              <p className="text-sm font-medium text-slate-700">Ikonka</p>
              <span className="text-xs text-slate-500">{ACTIVITY_ICON_OPTIONS.length} ta</span>
            </div>
            <input
              value={iconQuery}
              onChange={(e) => setIconQuery(e.target.value)}
              placeholder="Ikonka qidirish..."
              className="mt-1.5 w-full rounded-xl border border-slate-200 px-3 py-2 text-sm outline-none focus:border-slate-400"
            />
            <div className="mt-2 max-h-64 overflow-y-auto rounded-xl border border-slate-100 p-2">
              {icons.length === 0 ? (
                <p className="py-6 text-center text-sm text-slate-500">Ikonka topilmadi</p>
              ) : (
                <div className="grid grid-cols-5 gap-2 sm:grid-cols-6 md:grid-cols-7">
                  {icons.map((option) => {
                    const selected = icon === option.key
                    return (
                      <button
                        key={option.key}
                        type="button"
                        title={option.label}
                        onClick={() => setIcon(option.key)}
                        className="flex flex-col items-center gap-1 rounded-xl border px-1 py-2 text-[10px] font-medium"
                        style={{
                          borderColor: selected ? APP_COLOR : '#e2e8f0',
                          backgroundColor: selected ? '#1E3A5F0F' : 'white',
                          color: selected ? APP_COLOR : '#475569',
                        }}
                      >
                        <option.Icon size={18} />
                        <span className="w-full truncate">{option.label}</span>
                      </button>
                    )
                  })}
                </div>
              )}
            </div>
            {errors.icon ? <p className="mt-1 text-xs text-red-600">{errors.icon}</p> : null}
          </div>

          <Select
            label="Holat"
            value={status}
            onChange={(v) => setStatus(v as ActivityStatus)}
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
