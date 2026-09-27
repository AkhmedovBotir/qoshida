import { useState } from 'react'
import { mediaSrc } from '../../lib/media'
import { FastImage } from '../ui/FastImage'
import { cn } from '../../lib/cn'

export function ItemGallery({ images, alt }: { images: string[]; alt: string }) {
  const list = images.filter(Boolean)
  const [active, setActive] = useState(0)
  const current = list[Math.min(active, Math.max(0, list.length - 1))]
  const src = mediaSrc(current)

  if (!src) {
    return (
      <div className="flex aspect-square items-center justify-center rounded-[1.5rem] bg-slate-100 text-sm text-slate-400 lg:aspect-[4/3]">
        Rasm yo‘q
      </div>
    )
  }

  return (
    <div className="space-y-3">
      <div className="overflow-hidden rounded-[1.5rem] bg-white">
        <FastImage uri={src} alt={alt} height="100%" radius={0} className="aspect-square w-full lg:aspect-[4/3]" />
      </div>
      {list.length > 1 ? (
        <div className="no-scrollbar flex gap-2 overflow-x-auto">
          {list.map((url, index) => (
            <button
              key={`${url}-${index}`}
              type="button"
              onClick={() => setActive(index)}
              className={cn(
                'h-16 w-16 shrink-0 overflow-hidden rounded-2xl border-2',
                index === active ? 'border-[#2E5D90]' : 'border-transparent',
              )}
            >
              <FastImage uri={mediaSrc(url)} alt="" height={64} radius={0} />
            </button>
          ))}
        </div>
      ) : null}
    </div>
  )
}
