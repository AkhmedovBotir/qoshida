function localDigits(value: string) {
  const digits = value.replace(/\D/g, '')
  return (digits.startsWith('998') ? digits.slice(3) : digits).slice(0, 9)
}

export function formatUzPhoneLocal(value: string) {
  const d = localDigits(value)
  const parts = [d.slice(0, 2), d.slice(2, 5), d.slice(5, 7), d.slice(7, 9)].filter(Boolean)
  return parts.join(' ')
}

export function toUzPhoneE164(value: string) {
  return `+998${localDigits(value)}`
}

export function displayUzPhone(value: string) {
  const local = formatUzPhoneLocal(value)
  return local ? `+998 ${local}` : '+998'
}

type PhoneInputProps = {
  label?: string
  value: string
  onChange: (value: string) => void
  required?: boolean
  disabled?: boolean
  className?: string
}

export function PhoneInput({
  label = 'Telefon',
  value,
  onChange,
  required = true,
  disabled = false,
  className = '',
}: PhoneInputProps) {
  return (
    <label className={`block text-sm font-medium text-slate-700 ${className}`.trim()}>
      {label}
      <div className="mt-1.5 flex overflow-hidden rounded-xl border border-slate-200 focus-within:border-slate-400">
        <span className="flex items-center bg-slate-50 px-3 text-sm font-semibold text-slate-600 select-none">
          +998
        </span>
        <input
          type="tel"
          inputMode="numeric"
          autoComplete="tel-national"
          placeholder="90 123 45 67"
          value={formatUzPhoneLocal(value)}
          disabled={disabled}
          required={required}
          pattern="\d{2} \d{3} \d{2} \d{2}"
          title="90 123 45 67"
          onChange={(e) => onChange(toUzPhoneE164(e.target.value))}
          className="w-full px-3 py-2.5 text-slate-900 outline-none disabled:bg-slate-50"
        />
      </div>
    </label>
  )
}
