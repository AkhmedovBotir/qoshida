import { ChevronDown, ChevronRight, Loader2, Pencil, Trash2 } from 'lucide-react'
import { AnimatePresence, motion } from 'motion/react'
import { APP_COLOR } from '../../constants/config'
import { getRegionId, type RegionItem, type UIRegionType } from '../../types/region'
import { StatusToggle } from './StatusToggle'

const LEVEL = {
  viloyat: { mark: '#1E3A5F', badge: 'bg-slate-100 text-slate-700' },
  tuman: { mark: '#2A4A73', badge: 'bg-slate-100 text-slate-700' },
  mfy: { mark: '#5B6B7C', badge: 'bg-slate-100 text-slate-700' },
} as const

type RegionTreeProps = {
  viloyatlar: RegionItem[]
  tumanlar: Record<string, RegionItem[]>
  mfyList: Record<string, RegionItem[]>
  expandedViloyatlar: Set<string>
  expandedTumanlar: Set<string>
  loadingViloyatlar: Set<string>
  loadingTumanlar: Set<string>
  onToggleViloyat: (id: string) => void
  onToggleTuman: (id: string) => void
  onEdit: (item: RegionItem, type: UIRegionType) => void
  onDelete: (item: RegionItem, type: UIRegionType) => void
  onStatusToggle: (item: RegionItem, type: UIRegionType) => void
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

function AccordionChevron({ expanded }: { expanded: boolean }) {
  return (
    <span className="flex h-5 w-5 shrink-0 items-center justify-center text-slate-600">
      {expanded ? <ChevronDown size={18} /> : <ChevronRight size={18} />}
    </span>
  )
}

function LevelMark({ letter, color, size }: { letter: string; color: string; size: string }) {
  return (
    <div
      className={`flex shrink-0 items-center justify-center rounded-xl text-white ${size}`}
      style={{ backgroundColor: color }}
    >
      {letter}
    </div>
  )
}

export function RegionTree({
  viloyatlar,
  tumanlar,
  mfyList,
  expandedViloyatlar,
  expandedTumanlar,
  loadingViloyatlar,
  loadingTumanlar,
  onToggleViloyat,
  onToggleTuman,
  onEdit,
  onDelete,
  onStatusToggle,
}: RegionTreeProps) {
  if (!viloyatlar.length) {
    return <div className="py-12 text-center text-slate-500">Hozircha hududlar yo‘q</div>
  }

  return (
    <div className="divide-y divide-slate-200">
      {viloyatlar.map((viloyat) => {
        const viloyatId = getRegionId(viloyat)
        const isExpanded = expandedViloyatlar.has(viloyatId)
        const viloyatTumanlar = tumanlar[viloyatId] || []
        const isLoadingTumanlar = loadingViloyatlar.has(viloyatId)
        const isActive = viloyat.status === 'active'

        return (
          <div key={viloyatId} className={!isActive ? 'opacity-60' : ''}>
            <div
              role="button"
              tabIndex={0}
              aria-expanded={isExpanded}
              onClick={() => onToggleViloyat(viloyatId)}
              onKeyDown={(e) => {
                if (e.key === 'Enter' || e.key === ' ') {
                  e.preventDefault()
                  onToggleViloyat(viloyatId)
                }
              }}
              className="flex w-full cursor-pointer items-center gap-3 px-4 py-4 text-left transition hover:bg-slate-50 sm:px-6"
            >
              <AccordionChevron expanded={isExpanded} />
              <LevelMark letter="V" color={LEVEL.viloyat.mark} size="h-11 w-11 text-base font-bold" />
              <div className="min-w-0 flex-1">
                <div className="flex flex-wrap items-center gap-2">
                  <span className={`rounded-full px-2.5 py-0.5 text-xs font-semibold ${LEVEL.viloyat.badge}`}>
                    Viloyat
                  </span>
                  <span className="font-bold text-slate-900">{viloyat.name}</span>
                  {viloyat.code ? <span className="text-sm text-slate-500">({viloyat.code})</span> : null}
                </div>
              </div>
              <RowActions
                active={isActive}
                onToggleStatus={() => onStatusToggle(viloyat, 'viloyat')}
                onEdit={() => onEdit(viloyat, 'viloyat')}
                onDelete={() => onDelete(viloyat, 'viloyat')}
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
                  {isLoadingTumanlar ? (
                    <div className="flex justify-center py-8">
                      <Loader2 className="animate-spin" size={22} style={{ color: APP_COLOR }} />
                    </div>
                  ) : viloyatTumanlar.length === 0 ? (
                    <p className="px-8 py-5 text-sm italic text-slate-500 sm:px-14">Tumanlar yo‘q</p>
                  ) : (
                    <div className="space-y-0 py-1">
                      {viloyatTumanlar.map((tuman) => {
                        const tumanId = getRegionId(tuman)
                        const isTumanExpanded = expandedTumanlar.has(tumanId)
                        const tumanMfy = mfyList[tumanId] || []
                        const isLoadingMfy = loadingTumanlar.has(tumanId)
                        const tumanActive = tuman.status === 'active'

                        return (
                          <div
                            key={tumanId}
                            className={`border-t border-slate-200 first:border-t-0 ${!tumanActive ? 'opacity-60' : ''}`}
                          >
                            <div
                              role="button"
                              tabIndex={0}
                              aria-expanded={isTumanExpanded}
                              onClick={() => onToggleTuman(tumanId)}
                              onKeyDown={(e) => {
                                if (e.key === 'Enter' || e.key === ' ') {
                                  e.preventDefault()
                                  onToggleTuman(tumanId)
                                }
                              }}
                              className="flex w-full cursor-pointer items-center gap-3 px-6 py-3.5 text-left transition hover:bg-white/80 sm:px-10"
                            >
                              <AccordionChevron expanded={isTumanExpanded} />
                              <LevelMark letter="T" color={LEVEL.tuman.mark} size="h-9 w-9 text-sm font-bold" />
                              <div className="min-w-0 flex-1">
                                <div className="flex flex-wrap items-center gap-2">
                                  <span className={`rounded-full px-2 py-0.5 text-xs font-semibold ${LEVEL.tuman.badge}`}>
                                    Tuman
                                  </span>
                                  <span className="font-semibold text-slate-800">{tuman.name}</span>
                                  {tuman.code ? <span className="text-sm text-slate-500">({tuman.code})</span> : null}
                                </div>
                              </div>
                              <RowActions
                                active={tumanActive}
                                onToggleStatus={() => onStatusToggle(tuman, 'tuman')}
                                onEdit={() => onEdit(tuman, 'tuman')}
                                onDelete={() => onDelete(tuman, 'tuman')}
                              />
                            </div>

                            <AnimatePresence initial={false}>
                              {isTumanExpanded ? (
                                <motion.div
                                  initial={{ height: 0, opacity: 0 }}
                                  animate={{ height: 'auto', opacity: 1 }}
                                  exit={{ height: 0, opacity: 0 }}
                                  transition={{ duration: 0.2 }}
                                  className="overflow-hidden border-t border-slate-200 bg-white"
                                >
                                  {isLoadingMfy ? (
                                    <div className="flex justify-center py-6">
                                      <Loader2 className="animate-spin" size={20} style={{ color: APP_COLOR }} />
                                    </div>
                                  ) : tumanMfy.length === 0 ? (
                                    <p className="px-12 py-4 text-sm italic text-slate-500 sm:px-16">MFYlar yo‘q</p>
                                  ) : (
                                    <ul className="divide-y divide-slate-100">
                                      {tumanMfy.map((mfy) => {
                                        const mfyId = getRegionId(mfy)
                                        const mfyActive = mfy.status === 'active'

                                        return (
                                          <li
                                            key={mfyId}
                                            className={`flex items-center justify-between gap-3 px-8 py-2.5 sm:px-14 ${
                                              !mfyActive ? 'opacity-60' : ''
                                            } hover:bg-slate-50`}
                                          >
                                            <div className="flex min-w-0 items-center gap-3">
                                              <LevelMark letter="M" color={LEVEL.mfy.mark} size="h-8 w-8 text-xs font-bold" />
                                              <div className="flex flex-wrap items-center gap-2">
                                                <span className={`rounded-full px-2 py-0.5 text-xs font-semibold ${LEVEL.mfy.badge}`}>
                                                  MFY
                                                </span>
                                                <span className="font-medium text-slate-700">{mfy.name}</span>
                                                {mfy.code ? (
                                                  <span className="text-sm text-slate-500">({mfy.code})</span>
                                                ) : null}
                                              </div>
                                            </div>
                                            <RowActions
                                              active={mfyActive}
                                              onToggleStatus={() => onStatusToggle(mfy, 'mfy')}
                                              onEdit={() => onEdit(mfy, 'mfy')}
                                              onDelete={() => onDelete(mfy, 'mfy')}
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
