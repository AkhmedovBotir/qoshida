import type { ReactNode } from 'react'
import { cn } from '../../lib/cn'

type ButtonProps = {
  children: ReactNode
  type?: 'button' | 'submit'
  variant?: 'primary' | 'ghost' | 'danger' | 'soft'
  className?: string
  disabled?: boolean
  onClick?: () => void
}

export function Button({
  children,
  type = 'button',
  variant = 'primary',
  className,
  disabled,
  onClick,
}: ButtonProps) {
  return (
    <button
      type={type}
      disabled={disabled}
      onClick={onClick}
      className={cn(
        'inline-flex min-h-11 items-center justify-center rounded-2xl px-4 text-sm font-semibold transition disabled:opacity-55',
        variant === 'primary' && 'bg-[#2E5D90] text-white hover:bg-[#244A73]',
        variant === 'soft' && 'bg-[#E8F0F8] text-[#2E5D90] hover:bg-[#d7e6f3]',
        variant === 'ghost' && 'border border-[#E6EDF4] bg-white text-slate-700 hover:bg-slate-50',
        variant === 'danger' && 'bg-red-600 text-white hover:bg-red-700',
        className,
      )}
    >
      {children}
    </button>
  )
}
