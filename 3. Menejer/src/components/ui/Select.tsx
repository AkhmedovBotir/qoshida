import { Check, ChevronDown, Search } from 'lucide-react'
import { useEffect, useId, useMemo, useRef, useState } from 'react'
import { createPortal } from 'react-dom'
import { cn } from '../../lib/cn'

export type SelectOption = {
  value: string
  label: string
  hint?: string
  image?: string
  disabled?: boolean
}

type SelectProps = {
  label?: string
  value: string
  onChange: (value: string) => void
  options: SelectOption[]
  placeholder?: string
  searchable?: boolean
  required?: boolean
  disabled?: boolean
  className?: string
  emptyText?: string
}

export function Select({
  label,
  value,
  onChange,
  options,
  placeholder = 'Tanlang',
  searchable,
  required,
  disabled,
  className,
  emptyText = 'Topilmadi',
}: SelectProps) {
  const id = useId()
  const buttonRef = useRef<HTMLButtonElement>(null)
  const panelRef = useRef<HTMLDivElement>(null)
  const [open, setOpen] = useState(false)
  const [query, setQuery] = useState('')
  const [rect, setRect] = useState<DOMRect | null>(null)
  const enableSearch = searchable ?? options.length >= 8
  const selected = options.find((item) => item.value === value)

  const filtered = useMemo(() => {
    const q = query.trim().toLowerCase()
    if (!q) return options
    return options.filter(
      (item) => item.label.toLowerCase().includes(q) || (item.hint ?? '').toLowerCase().includes(q),
    )
  }, [options, query])

  function place() {
    if (!buttonRef.current) return
    setRect(buttonRef.current.getBoundingClientRect())
  }

  useEffect(() => {
    if (!open) return
    place()
    const onDoc = (event: MouseEvent) => {
      const target = event.target as Node
      if (buttonRef.current?.contains(target) || panelRef.current?.contains(target)) return
      setOpen(false)
    }
    const onKey = (event: KeyboardEvent) => {
      if (event.key === 'Escape') setOpen(false)
    }
    window.addEventListener('mousedown', onDoc)
    window.addEventListener('keydown', onKey)
    window.addEventListener('resize', place)
    window.addEventListener('scroll', place, true)
    return () => {
      window.removeEventListener('mousedown', onDoc)
      window.removeEventListener('keydown', onKey)
      window.removeEventListener('resize', place)
      window.removeEventListener('scroll', place, true)
    }
  }, [open])

  useEffect(() => {
    if (!open) setQuery('')
  }, [open])

  const panelStyle = (() => {
    if (!rect) return undefined
    const maxH = 280
    const spaceBelow = window.innerHeight - rect.bottom - 8
    const openUp = spaceBelow < 180 && rect.top > spaceBelow
    return {
      left: Math.max(8, Math.min(rect.left, window.innerWidth - rect.width - 8)),
      width: Math.max(rect.width, 220),
      maxHeight: maxH,
      ...(openUp ? { bottom: window.innerHeight - rect.top + 6 } : { top: rect.bottom + 6 }),
    }
  })()

  return (
    <div className={cn('block text-sm font-medium text-slate-700', className)}>
      {label ? (
        <span className="mb-1.5 block">
          {label}
          {required ? <span className="text-red-500"> *</span> : null}
        </span>
      ) : null}
      <input id={id} value={value} required={required} readOnly tabIndex={-1} className="sr-only" />
      <button
        ref={buttonRef}
        type="button"
        disabled={disabled}
        aria-haspopup="listbox"
        aria-expanded={open}
        onClick={() => {
          if (disabled) return
          setOpen((v) => !v)
        }}
        className={cn(
          'flex w-full items-center gap-2 rounded-xl border border-slate-200 bg-white px-3 py-2.5 text-left text-sm font-normal outline-none focus:border-slate-400',
          disabled && 'cursor-not-allowed bg-slate-50 text-slate-400',
        )}
      >
        {selected?.image ? (
          <img src={selected.image} alt="" className="h-7 w-7 shrink-0 rounded-lg object-cover" />
        ) : null}
        <span className={cn('min-w-0 flex-1 truncate', selected ? 'text-slate-900' : 'text-slate-400')}>
          {selected ? selected.label : placeholder}
        </span>
        {selected?.hint ? <span className="hidden shrink-0 text-xs text-slate-400 sm:inline">{selected.hint}</span> : null}
        <ChevronDown size={16} className={cn('shrink-0 text-slate-400 transition', open && 'rotate-180')} />
      </button>

      {open && rect
        ? createPortal(
            <div
              ref={panelRef}
              role="listbox"
              style={panelStyle}
              className="fixed z-[80] overflow-hidden rounded-xl border border-slate-200 bg-white shadow-xl"
            >
              {enableSearch ? (
                <div className="border-b border-slate-100 p-2">
                  <div className="relative">
                    <Search size={14} className="pointer-events-none absolute top-1/2 left-2.5 -translate-y-1/2 text-slate-400" />
                    <input
                      autoFocus
                      value={query}
                      onChange={(e) => setQuery(e.target.value)}
                      placeholder="Qidirish..."
                      className="w-full rounded-lg border border-slate-200 py-2 pr-3 pl-8 text-sm outline-none focus:border-slate-400"
                    />
                  </div>
                </div>
              ) : null}
              <div className="max-h-60 overflow-y-auto p-1">
                {filtered.length === 0 ? (
                  <p className="px-3 py-4 text-center text-sm text-slate-500">{emptyText}</p>
                ) : (
                  filtered.map((item) => {
                    const active = item.value === value
                    return (
                      <button
                        key={item.value || item.label}
                        type="button"
                        role="option"
                        aria-selected={active}
                        disabled={item.disabled}
                        onClick={() => {
                          if (item.disabled) return
                          onChange(item.value)
                          setOpen(false)
                        }}
                        className={cn(
                          'flex w-full items-center gap-2 rounded-lg px-2.5 py-2 text-left text-sm',
                          item.disabled ? 'cursor-not-allowed text-slate-300' : 'hover:bg-slate-50',
                          active && 'bg-slate-100 font-medium text-slate-900',
                        )}
                      >
                        {item.image ? <img src={item.image} alt="" className="h-8 w-8 shrink-0 rounded-lg object-cover" /> : null}
                        <span className="min-w-0 flex-1">
                          <span className="block truncate">{item.label}</span>
                          {item.hint ? <span className="block truncate text-xs text-slate-500">{item.hint}</span> : null}
                        </span>
                        {active ? <Check size={15} className="shrink-0 text-slate-700" /> : null}
                      </button>
                    )
                  })
                )}
              </div>
            </div>,
            document.body,
          )
        : null}
    </div>
  )
}
