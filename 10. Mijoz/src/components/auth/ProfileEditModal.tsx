import { Camera, X } from 'lucide-react'
import { FormEvent, useEffect, useRef } from 'react'
import { createPortal } from 'react-dom'
import { CustomerFields, type CustomerFormValue } from './CustomerFields'
import { mediaSrc } from '../../lib/media'

export function ProfileEditModal({
  phone,
  avatar,
  form,
  onChange,
  onAvatar,
  onClearAvatar,
  onClose,
  onSubmit,
  busy,
}: {
  phone: string
  avatar?: string
  form: CustomerFormValue
  onChange: (value: CustomerFormValue) => void
  onAvatar: (file: File) => void
  onClearAvatar: () => void
  onClose: () => void
  onSubmit: (event: FormEvent) => void
  busy: boolean
}) {
  const inputRef = useRef<HTMLInputElement>(null)
  const preview = avatar ? mediaSrc(avatar) : ''

  useEffect(() => {
    const prev = document.body.style.overflow
    document.body.style.overflow = 'hidden'
    const onKey = (event: KeyboardEvent) => {
      if (event.key === 'Escape') onClose()
    }
    window.addEventListener('keydown', onKey)
    return () => {
      document.body.style.overflow = prev
      window.removeEventListener('keydown', onKey)
    }
  }, [onClose])

  return createPortal(
    <div className="fixed inset-0 z-[70] flex items-end justify-center sm:items-center">
      <button type="button" className="absolute inset-0 bg-[#101826]/50" aria-label="Yopish" onClick={onClose} />
      <section className="relative z-10 flex max-h-[92svh] w-full max-w-lg flex-col rounded-t-[1.75rem] bg-white shadow-2xl sm:mx-4 sm:rounded-[1.75rem]">
        <div className="flex items-center justify-between border-b border-[#E6EDF4] px-5 py-4">
          <h2 className="text-lg font-extrabold text-slate-900">Ma’lumotlarni o‘zgartirish</h2>
          <button
            type="button"
            onClick={onClose}
            className="flex h-9 w-9 items-center justify-center rounded-full bg-[#F2F5F8] text-slate-600"
            aria-label="Yopish"
          >
            <X size={18} />
          </button>
        </div>
        <form onSubmit={onSubmit} className="flex min-h-0 flex-1 flex-col">
          <div className="space-y-4 overflow-y-auto px-5 py-4">
            <div className="flex flex-col items-center gap-2">
              <button
                type="button"
                onClick={() => inputRef.current?.click()}
                className="relative h-24 w-24 overflow-hidden rounded-full bg-[#F4B942] text-2xl font-extrabold text-[#3A2A08]"
              >
                {preview ? (
                  <img src={preview} alt="" className="h-full w-full object-cover" />
                ) : (
                  <Camera size={28} className="mx-auto text-[#3A2A08]" />
                )}
                <span className="absolute right-1 bottom-1 flex h-8 w-8 items-center justify-center rounded-full bg-[#1B3A5C] text-white">
                  <Camera size={14} />
                </span>
              </button>
              <div className="flex gap-3 text-xs font-semibold">
                <button type="button" onClick={() => inputRef.current?.click()} className="text-[#2E5D90]">
                  Rasm yuklash
                </button>
                {avatar ? (
                  <button type="button" onClick={onClearAvatar} className="text-red-600">
                    O‘chirish
                  </button>
                ) : null}
              </div>
              <input
                ref={inputRef}
                type="file"
                accept="image/png,image/jpeg,image/webp"
                className="hidden"
                onChange={(event) => {
                  const file = event.target.files?.[0]
                  event.target.value = ''
                  if (file) onAvatar(file)
                }}
              />
            </div>
            <label className="block text-sm font-medium text-slate-700">
              Telefon raqam
              <input
                readOnly
                value={phone}
                className="mt-1.5 w-full rounded-2xl border border-[#E6EDF4] bg-[#F2F5F8] px-3.5 py-3 text-slate-600 outline-none"
              />
            </label>
            <CustomerFields value={form} onChange={onChange} />
          </div>
          <div className="border-t border-[#E6EDF4] p-4 pb-[max(1rem,env(safe-area-inset-bottom))]">
            <button
              type="submit"
              disabled={busy}
              className="w-full rounded-2xl bg-[#2E5D90] py-3 text-sm font-semibold text-white disabled:opacity-60"
            >
              {busy ? 'Saqlanmoqda...' : 'Saqlash'}
            </button>
          </div>
        </form>
      </section>
    </div>,
    document.body,
  )
}
