import { ChevronDown, ChevronRight, Pencil, Trash2 } from 'lucide-react'
import { AnimatePresence, motion } from 'motion/react'
import { categoryImageSrc } from '../../lib/categories'
import { getCategoryId, type CategoryItem } from '../../types/category'
import { ImageThumb } from '../ui/ImageLightbox'
import { StatusToggle } from '../ui/StatusToggle'

function CategoryThumb({ item, size, fallback }: { item: CategoryItem; size: string; fallback: string }) {
  const src = categoryImageSrc(item.image_url)
  if (src) {
    return <ImageThumb src={src} className={`${size} shrink-0 rounded-xl object-cover ring-1 ring-slate-200`} />
  }
  return (
    <div
      className={`${size} flex shrink-0 items-center justify-center rounded-xl text-sm font-bold text-white`}
      style={{ backgroundColor: fallback === 'A' ? '#1E3A5F' : '#2A4A73' }}
    >
      {fallback}
    </div>
  )
}

type CategoryTreeProps = {
  roots: CategoryItem[]
  childrenByParent: Record<string, CategoryItem[]>
  expanded: Set<string>
  onToggle: (id: string) => void
  onEdit: (item: CategoryItem, kind: 'root' | 'child') => void
  onDelete: (item: CategoryItem) => void
  onStatusToggle: (item: CategoryItem) => void
}

function RowActions({
  active,
  onToggleStatus,
  onEdit,
  onDelete,
}: {
  active: boolean
  onToggleStatus: () => void
  onEdit: () => void
  onDelete: () => void
}) {
  return (
    <div
      className="flex shrink-0 items-center gap-2 sm:gap-3"
      onClick={(e) => e.stopPropagation()}
      onKeyDown={(e) => e.stopPropagation()}
    >
      <div className="hidden items-center gap-2 sm:flex">
        <span className={`text-xs font-medium ${active ? 'text-slate-700' : 'text-slate-500'}`}>
          {active ? 'Faol' : 'Nofaol'}
        </span>
        <StatusToggle active={active} onChange={onToggleStatus} size="sm" />
      </div>
      <button
        type="button"
        aria-label="Tahrirlash"
        title="Tahrirlash"
        onClick={onEdit}
        className="rounded-lg border border-slate-200 p-2 text-slate-600"
      >
        <Pencil size={14} />
      </button>
      <button
        type="button"
        aria-label="O‘chirish"
        title="O‘chirish"
        onClick={onDelete}
        className="rounded-lg border border-red-100 p-2 text-red-600"
      >
        <Trash2 size={14} />
      </button>
    </div>
  )
}

export function CategoryTree({
  roots,
  childrenByParent,
  expanded,
  onToggle,
  onEdit,
  onDelete,
  onStatusToggle,
}: CategoryTreeProps) {
  if (!roots.length) {
    return <div className="py-12 text-center text-slate-500">Hozircha kategoriyalar yo‘q</div>
  }

  return (
    <div className="divide-y divide-slate-200">
      {roots.map((root) => {
        const rootId = getCategoryId(root)
        const isExpanded = expanded.has(rootId)
        const children = childrenByParent[rootId] || []
        const isActive = root.status === 'active'

        return (
          <div key={rootId} className={!isActive ? 'opacity-60' : ''}>
            <div
              role="button"
              tabIndex={0}
              aria-expanded={isExpanded}
              onClick={() => onToggle(rootId)}
              onKeyDown={(e) => {
                if (e.key === 'Enter' || e.key === ' ') {
                  e.preventDefault()
                  onToggle(rootId)
                }
              }}
              className="flex w-full cursor-pointer items-center gap-3 px-4 py-4 text-left transition hover:bg-slate-50 sm:px-6"
            >
              <span className="flex h-5 w-5 shrink-0 items-center justify-center text-slate-600">
                {isExpanded ? <ChevronDown size={18} /> : <ChevronRight size={18} />}
              </span>
              <CategoryThumb item={root} size="h-11 w-11" fallback="A" />
              <div className="min-w-0 flex-1">
                <div className="flex flex-wrap items-center gap-2">
                  <span className="rounded-full bg-slate-100 px-2.5 py-0.5 text-xs font-semibold text-slate-700">
                    Asosiy
                  </span>
                  <span className="font-bold text-slate-900">{root.name}</span>
                  {root.slug ? <span className="text-sm text-slate-500">({root.slug})</span> : null}
                  {root.censored ? (
                    <span className="rounded-full bg-amber-50 px-2 py-0.5 text-xs font-semibold text-amber-700">
                      Cheklangan
                    </span>
                  ) : null}
                </div>
                <p className="mt-1 text-xs text-slate-500">{children.length} ta ichki kategoriya</p>
              </div>
              <RowActions
                active={isActive}
                onToggleStatus={() => onStatusToggle(root)}
                onEdit={() => onEdit(root, 'root')}
                onDelete={() => onDelete(root)}
              />
            </div>

            <AnimatePresence initial={false}>
              {isExpanded ? (
                <motion.div
                  initial={{ height: 0, opacity: 0 }}
                  animate={{ height: 'auto', opacity: 1 }}
                  exit={{ height: 0, opacity: 0 }}
                  transition={{ duration: 0.2 }}
                  className="overflow-hidden border-t border-slate-200 bg-[#F4F6F9]"
                >
                  {children.length === 0 ? (
                    <p className="px-8 py-5 text-sm italic text-slate-500 sm:px-14">Ichki kategoriyalar yo‘q</p>
                  ) : (
                    <ul className="divide-y divide-slate-200">
                      {children.map((child) => {
                        const childId = getCategoryId(child)
                        const childActive = child.status === 'active'
                        return (
                          <li
                            key={childId}
                            className={`flex items-center justify-between gap-3 px-6 py-3.5 sm:px-10 ${
                              !childActive ? 'opacity-60' : ''
                            } hover:bg-white/80`}
                          >
                            <div className="flex min-w-0 items-center gap-3">
                              <CategoryThumb item={child} size="h-9 w-9" fallback="I" />
                              <div className="flex flex-wrap items-center gap-2">
                                <span className="rounded-full bg-slate-100 px-2 py-0.5 text-xs font-semibold text-slate-700">
                                  Ichki
                                </span>
                                <span className="font-semibold text-slate-800">{child.name}</span>
                                {child.slug ? <span className="text-sm text-slate-500">({child.slug})</span> : null}
                                {child.censored ? (
                                  <span className="rounded-full bg-amber-50 px-2 py-0.5 text-xs font-semibold text-amber-700">
                                    Cheklangan
                                  </span>
                                ) : null}
                              </div>
                            </div>
                            <RowActions
                              active={childActive}
                              onToggleStatus={() => onStatusToggle(child)}
                              onEdit={() => onEdit(child, 'child')}
                              onDelete={() => onDelete(child)}
                            />
                          </li>
                        )
                      })}
                    </ul>
                  )}
                </motion.div>
              ) : null}
            </AnimatePresence>
          </div>
        )
      })}
    </div>
  )
}
