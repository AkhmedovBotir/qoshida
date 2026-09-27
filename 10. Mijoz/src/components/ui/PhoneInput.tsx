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

type PhoneInputProps = {
  label?: string
  value: string
  onChange: (value: string) => void
  required?: boolean
  disabled?: boolean
  className?: string
  size?: 'md' | 'lg'
}

export function PhoneInput({
  label = 'Telefon',
  value,
  onChange,
  required = true,
  disabled = false,
  className = '',
  size = 'md',
}: PhoneInputProps) {
  const large = size === 'lg'
  return (
    <label className={`block text-sm font-medium text-slate-700 ${className}`.trim()}>
      {label}
      <div
        className={`mt-1.5 flex overflow-hidden border focus-within:border-[#101826] ${
          large
            ? 'rounded-full border-transparent bg-white shadow-[0_8px_24px_rgba(16,24,38,0.08)]'
            : 'rounded-2xl border-[#E6EDF4] focus-within:border-[#2E5D90]'
        }`}
      >
        <span
          className={`flex items-center px-3 text-sm font-bold select-none ${
            large ? 'bg-[#101826] text-[#F4B942]' : 'bg-slate-50 text-slate-600'
          }`}
        >
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
          className={`w-full px-3 text-slate-900 outline-none disabled:bg-slate-50 ${large ? 'py-3.5 text-base' : 'py-2.5'}`}
        />
      </div>
    </label>
  )
}
