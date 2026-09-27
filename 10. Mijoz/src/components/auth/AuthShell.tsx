import { ArrowLeft, ShieldCheck } from 'lucide-react'
import type { ReactNode } from 'react'

const STEPS = [
  { id: 'phone', label: 'Telefon' },
  { id: 'otp', label: 'Kod' },
  { id: 'done', label: 'Ma’lumot' },
] as const

function stepIndex(step: string) {
  if (step === 'phone') return 0
  if (step === 'otp') return 1
  return 2
}

export function AuthShell({
  step,
  title,
  subtitle,
  onBack,
  backLabel,
  children,
}: {
  step: string
  title: string
  subtitle: string
  onBack: () => void
  backLabel: string
  children: ReactNode
}) {
  const active = stepIndex(step)

  return (
    <main className="relative min-h-svh overflow-y-auto bg-[#101826] text-white">
      <div className="pointer-events-none absolute top-10 left-1/2 h-72 w-72 -translate-x-1/2 rounded-full bg-[#2E5D90]/45 blur-3xl" />
      <div className="pointer-events-none absolute right-[-3rem] bottom-10 h-56 w-56 rounded-full bg-[#F4B942]/20 blur-3xl" />

      <div className="relative mx-auto flex min-h-svh w-full max-w-md flex-col justify-center px-4 py-[max(1.25rem,env(safe-area-inset-top))] sm:max-w-lg sm:px-6">
        <header className="mb-5 flex items-center justify-between gap-3">
          <button
            type="button"
            onClick={onBack}
            className="inline-flex items-center gap-1.5 rounded-full bg-white/10 px-3 py-2 text-sm font-medium text-white/90"
          >
            <ArrowLeft size={16} />
            {backLabel}
          </button>
          <p className="text-xs font-bold tracking-[1.5px] text-white/60 uppercase">Qoshida</p>
        </header>

        <div className="mb-5 flex items-center gap-2 px-1">
          {STEPS.map((item, index) => (
            <div key={item.id} className="flex flex-1 items-center gap-2">
              <div
                className={`flex h-7 w-7 shrink-0 items-center justify-center rounded-full text-[11px] font-bold ${
                  index <= active ? 'bg-[#F4B942] text-[#3A2A08]' : 'bg-white/10 text-white/50'
                }`}
              >
                {index + 1}
              </div>
              <span className={`hidden text-xs font-medium sm:inline ${index <= active ? 'text-white' : 'text-white/40'}`}>
                {item.label}
              </span>
              {index < STEPS.length - 1 ? (
                <span className={`h-px flex-1 ${index < active ? 'bg-[#F4B942]' : 'bg-white/15'}`} />
              ) : null}
            </div>
          ))}
        </div>

        <section className="rounded-[1.75rem] bg-[#F7F4EE] p-5 text-slate-900 shadow-[0_20px_60px_rgba(0,0,0,0.28)] sm:p-7">
          <div className="mb-5 flex items-start gap-3">
            <span className="mt-0.5 flex h-10 w-10 shrink-0 items-center justify-center rounded-2xl bg-[#101826] text-[#F4B942]">
              <ShieldCheck size={18} />
            </span>
            <div>
              <h1 className="text-2xl font-extrabold tracking-tight text-[#101826]">{title}</h1>
              <p className="mt-1 text-sm leading-6 text-slate-500">{subtitle}</p>
            </div>
          </div>
          {children}
        </section>
      </div>
    </main>
  )
}
