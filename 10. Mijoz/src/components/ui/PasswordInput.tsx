import { Eye, EyeOff } from 'lucide-react'
import { useState, type InputHTMLAttributes } from 'react'
import { cn } from '../../lib/cn'

type PasswordInputProps = Omit<InputHTMLAttributes<HTMLInputElement>, 'type'> & {
  label?: string
}

export function PasswordInput({ label, className, id, ...props }: PasswordInputProps) {
  const [visible, setVisible] = useState(false)

  return (
    <label className={cn('block text-sm font-medium text-slate-700', className)}>
      {label}
      <span className="relative mt-1.5 block">
        <input
          {...props}
          id={id}
          type={visible ? 'text' : 'password'}
          className="w-full rounded-full border-0 bg-white px-4 py-3.5 pr-11 text-slate-900 outline-none ring-1 ring-[#101826]/10 focus:ring-[#101826]"
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
    </label>
  )
}
