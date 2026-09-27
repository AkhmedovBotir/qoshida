import { ChevronLeft, ChevronRight, X } from 'lucide-react'
import { useEffect, useState } from 'react'
import { createPortal } from 'react-dom'

type ImageLightboxProps = {
  images: string[]
  index?: number
  onClose: () => void
}

export function ImageLightbox({ images, index = 0, onClose }: ImageLightboxProps) {
  const list = images.filter(Boolean)
  const [current, setCurrent] = useState(Math.min(Math.max(index, 0), Math.max(list.length - 1, 0)))
  const many = list.length > 1

  useEffect(() => {
    const onKey = (event: KeyboardEvent) => {
      if (event.key === 'Escape') onClose()
      if (event.key === 'ArrowRight' && many) setCurrent((v) => (v + 1) % list.length)
      if (event.key === 'ArrowLeft' && many) setCurrent((v) => (v - 1 + list.length) % list.length)
    }
    window.addEventListener('keydown', onKey)
    return () => window.removeEventListener('keydown', onKey)
  }, [list.length, many, onClose])

  if (!list.length) return null

  return createPortal(
    <div className="fixed inset-0 z-[90] flex items-center justify-center bg-slate-950/80 p-4" onClick={onClose}>
      <button
        type="button"
        onClick={onClose}
        className="absolute top-4 right-4 rounded-full bg-white/15 p-2 text-white hover:bg-white/25"
        aria-label="Yopish"
      >
        <X size={20} />
      </button>
      {many ? (
        <button
          type="button"
          onClick={(e) => {
            e.stopPropagation()
            setCurrent((v) => (v - 1 + list.length) % list.length)
          }}
          className="absolute left-3 rounded-full bg-white/15 p-2 text-white hover:bg-white/25 sm:left-6"
          aria-label="Oldingi rasm"
        >
          <ChevronLeft size={22} />
        </button>
      ) : null}
      <img
        src={list[current]}
        alt=""
        onClick={(e) => e.stopPropagation()}
        className="max-h-[86vh] max-w-[min(96vw,1100px)] rounded-xl object-contain shadow-2xl"
      />
      {many ? (
        <button
          type="button"
          onClick={(e) => {
            e.stopPropagation()
            setCurrent((v) => (v + 1) % list.length)
          }}
          className="absolute right-3 rounded-full bg-white/15 p-2 text-white hover:bg-white/25 sm:right-6"
          aria-label="Keyingi rasm"
        >
          <ChevronRight size={22} />
        </button>
      ) : null}
      {many ? (
        <p className="absolute bottom-4 left-1/2 -translate-x-1/2 rounded-full bg-black/50 px-3 py-1 text-xs text-white">
          {current + 1} / {list.length}
        </p>
      ) : null}
    </div>,
    document.body,
  )
}

type ImageThumbProps = {
  src: string
  images?: string[]
  className?: string
  alt?: string
}

export function ImageThumb({ src, images, className, alt = '' }: ImageThumbProps) {
  const [open, setOpen] = useState(false)
  if (!src) return null
  const list = images?.filter(Boolean).length ? images.filter(Boolean) : [src]
  const start = Math.max(0, list.indexOf(src))

  return (
    <>
      <button type="button" onClick={() => setOpen(true)} className="block overflow-hidden rounded-lg" aria-label={alt || 'Rasmni ochish'}>
        <img src={src} alt={alt} className={className ?? 'h-12 w-12 object-cover'} />
      </button>
      {open ? <ImageLightbox images={list} index={start} onClose={() => setOpen(false)} /> : null}
    </>
  )
}
