import { Eye, EyeOff } from 'lucide-react'
import { useState, type InputHTMLAttributes } from 'react'
import { cn } from '../../lib/cn'
import { PasswordRules } from './PasswordRules'

type PasswordInputProps = Omit<InputHTMLAttributes<HTMLInputElement>, 'type'> & {
  label?: string
  showRules?: boolean
}

export function PasswordInput({ label, className, id, showRules, value, ...props }: PasswordInputProps) {
  const [visible, setVisible] = useState(false)
  const current = typeof value === 'string' ? value : ''

  return (
    <label className={cn('block text-sm font-medium text-slate-700', className)}>
      {label}
      <span className="relative mt-1.5 block">
        <input
          {...props}
          id={id}
          value={value}
          type={visible ? 'text' : 'password'}
          className="w-full rounded-xl border border-slate-200 px-3 py-2.5 pr-11 text-slate-900 outline-none focus:border-slate-400"
        />
        <button
          type="button"
          onClick={() => setVisible((v) => !v)}
          className="absolute inset-y-0 right-0 flex w-11 items-center justify-center text-slate-500 hover:text-slate-800"
          aria-label={visible ? 'Parolni yashirish' : "Parolni ko'rish"}
        >
          {visible ? <EyeOff size={18} /> : <Eye size={18} />}
        </button>
      </span>
      {showRules ? (
        <span className="mt-2 block font-normal">
          <PasswordRules value={current} />
        </span>
      ) : null}
    </label>
  )
}
