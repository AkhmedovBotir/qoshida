import { GeoFields } from '../geo/GeoFields'
import { fieldClass } from '../../constants/theme'
import { cn } from '../../lib/cn'

export type CustomerFormValue = {
  first_name: string
  last_name: string
  birth_date: string
  region_id: string
  district_id: string
  mfy_id: string
}

const inputClass = cn(fieldClass, 'rounded-2xl bg-white')

function birthLimits() {
  const today = new Date()
  const max = new Date(today.getFullYear() - 14, today.getMonth(), today.getDate())
  const min = new Date(today.getFullYear() - 100, today.getMonth(), today.getDate())
  const pad = (n: number) => String(n).padStart(2, '0')
  const iso = (d: Date) => `${d.getFullYear()}-${pad(d.getMonth() + 1)}-${pad(d.getDate())}`
  return { min: iso(min), max: iso(max) }
}

export function CustomerFields({
  value,
  onChange,
  tone = 'market',
}: {
  value: CustomerFormValue
  onChange: (value: CustomerFormValue) => void
  tone?: 'market' | 'auth'
}) {
  const limits = birthLimits()
  const cls = tone === 'auth' ? 'mt-1.5 w-full rounded-2xl bg-white px-4 py-3 outline-none ring-1 ring-[#101826]/10 focus:ring-[#101826]' : inputClass

  return (
    <div className="space-y-3">
      <div className="grid gap-3 sm:grid-cols-2">
        <label className="block text-sm font-medium text-slate-700">
          Ism
          <input
            required
            minLength={2}
            maxLength={60}
            value={value.first_name}
            onChange={(e) => onChange({ ...value, first_name: e.target.value })}
            className={cls}
            autoComplete="given-name"
          />
        </label>
        <label className="block text-sm font-medium text-slate-700">
          Familiya
          <input
            required
            minLength={2}
            maxLength={60}
            value={value.last_name}
            onChange={(e) => onChange({ ...value, last_name: e.target.value })}
            className={cls}
            autoComplete="family-name"
          />
        </label>
      </div>
      <label className="block text-sm font-medium text-slate-700">
        Tug‘ilgan sana
        <input
          required
          type="date"
          min={limits.min}
          max={limits.max}
          value={value.birth_date}
          onChange={(e) => onChange({ ...value, birth_date: e.target.value })}
          className={cls}
        />
      </label>
      <GeoFields
        required
        value={{ region_id: value.region_id, district_id: value.district_id, mfy_id: value.mfy_id }}
        onChange={(geo) => onChange({ ...value, ...geo })}
      />
    </div>
  )
}
