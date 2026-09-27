import { useRef, type KeyboardEvent, type ClipboardEvent } from 'react'

type OtpBoxesProps = {
  value: string
  onChange: (value: string) => void
  disabled?: boolean
}

export function OtpBoxes({ value, onChange, disabled }: OtpBoxesProps) {
  const refs = useRef<Array<HTMLInputElement | null>>([])
  const digits = value.padEnd(6, ' ').slice(0, 6).split('')

  function setDigit(index: number, raw: string) {
    const char = raw.replace(/\D/g, '').slice(-1)
    const next = value.split('')
    while (next.length < 6) next.push('')
    next[index] = char
    const joined = next.join('').replace(/\s/g, '').slice(0, 6)
    onChange(joined)
    if (char && index < 5) refs.current[index + 1]?.focus()
  }

  function onKeyDown(index: number, event: KeyboardEvent<HTMLInputElement>) {
    if (event.key === 'Backspace' && !value[index] && index > 0) {
      refs.current[index - 1]?.focus()
      onChange(value.slice(0, index - 1))
    }
  }

  function onPaste(event: ClipboardEvent<HTMLInputElement>) {
    event.preventDefault()
    const text = event.clipboardData.getData('text').replace(/\D/g, '').slice(0, 6)
    onChange(text)
    refs.current[Math.min(text.length, 5)]?.focus()
  }

  return (
    <div className="flex justify-center gap-2">
      {digits.map((digit, index) => (
        <input
          key={index}
          ref={(el) => {
            refs.current[index] = el
          }}
          inputMode="numeric"
          maxLength={1}
          disabled={disabled}
          value={digit.trim()}
          onChange={(e) => setDigit(index, e.target.value)}
          onKeyDown={(e) => onKeyDown(index, e)}
          onPaste={onPaste}
          className="h-12 w-11 rounded-xl border border-slate-200 text-center text-lg font-bold text-slate-900 outline-none focus:border-rose-600"
        />
      ))}
    </div>
  )
}
