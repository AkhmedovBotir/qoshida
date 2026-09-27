import type { ReactNode } from 'react'
import { cn } from '../../lib/cn'

type ScreenProps = {
  children: ReactNode
  className?: string
}

export function Screen({ children, className }: ScreenProps) {
  return (
    <main className={cn('min-h-svh bg-[#F2F5F8]', className)}>
      <div className="mx-auto w-full max-w-3xl px-6 py-8">{children}</div>
    </main>
  )
}
