import { Loader2, Plus, Search, Upload } from 'lucide-react'
import { toast } from '../lib/snack'
import { motion } from 'motion/react'
import { useEffect, useMemo, useState } from 'react'
import { RegionFormModal } from '../components/regions/RegionFormModal'
import { RegionTree } from '../components/regions/RegionTree'
import { ConfirmModal } from '../components/ui/ConfirmModal'
import { APP_COLOR } from '../constants/config'
import { ApiError } from '../lib/api'
import { createRegion, deleteRegion, importRegions, listRegions, updateRegion } from '../lib/regions'
import {
  API_TYPE,
  getRegionId,
  REGION_TYPES,
  type RegionItem,
  type RegionStatus,
  type UIRegionType,
} from '../types/region'

export function RegionsPage() {
  const [viloyatlar, setViloyatlar] = useState<RegionItem[]>([])
  const [tumanlar, setTumanlar] = useState<Record<string, RegionItem[]>>({})
  const [mfyList, setMfyList] = useState<Record<string, RegionItem[]>>({})
  const [allTumanlar, setAllTumanlar] = useState<RegionItem[]>([])
  const [loading, setLoading] = useState(true)
  const [searchTerm, setSearchTerm] = useState('')
  const [expandedViloyatlar, setExpandedViloyatlar] = useState<Set<string>>(new Set())
  const [expandedTumanlar, setExpandedTumanlar] = useState<Set<string>>(new Set())
  const [loadingViloyatlar, setLoadingViloyatlar] = useState<Set<string>>(new Set())
  const [loadingTumanlar, setLoadingTumanlar] = useState<Set<string>>(new Set())

  const [formOpen, setFormOpen] = useState(false)
  const [formType, setFormType] = useState<UIRegionType>(REGION_TYPES.viloyat)
  const [editingItem, setEditingItem] = useState<RegionItem | null>(null)
  const [saving, setSaving] = useState(false)
  const [importing, setImporting] = useState(false)
  const [importOpen, setImportOpen] = useState(false)

  const [deleteTarget, setDeleteTarget] = useState<{ item: RegionItem; type: UIRegionType } | null>(null)
  const [deleting, setDeleting] = useState(false)

  async function reloadExpandedChildren(viloyatIds: string[], tumanIds: string[]) {
    const tumanEntries = await Promise.all(
      viloyatIds.map(async (viloyatId) => [viloyatId, await listRegions({ type: 'district', parent_id: viloyatId })] as const),
    )

    if (tumanEntries.length) {
      setTumanlar((prev) => ({ ...prev, ...Object.fromEntries(tumanEntries) }))
      const allTumanFlat = tumanEntries.flatMap(([, list]) => list)
      setAllTumanlar((prev) => {
        const map = new Map(prev.map((t) => [getRegionId(t), t]))
        allTumanFlat.forEach((t) => map.set(getRegionId(t), t))
        return [...map.values()]
      })
    }

    const mfyEntries = await Promise.all(
      tumanIds.map(async (tumanId) => [tumanId, await listRegions({ type: 'mfy', parent_id: tumanId })] as const),
    )
    if (mfyEntries.length) {
      setMfyList((prev) => ({ ...prev, ...Object.fromEntries(mfyEntries) }))
    }
  }

  async function loadData() {
    try {
      const viloyatData = await listRegions({ type: 'region' })
      setViloyatlar(viloyatData)
      setTumanlar({})
      setMfyList({})
      await reloadExpandedChildren([...expandedViloyatlar], [...expandedTumanlar])
    } catch (err) {
      if (!(err instanceof ApiError)) toast.error('Ma’lumotlarni yuklab bo‘lmadi')
    } finally {
      setLoading(false)
    }
  }

  useEffect(() => {
    let active = true
    listRegions({ type: 'region' })
      .then((data) => {
        if (active) setViloyatlar(data)
      })
      .catch((err) => {
        if (active) if (!(err instanceof ApiError)) toast.error('Ma’lumotlarni yuklab bo‘lmadi')
      })
      .finally(() => {
        if (active) setLoading(false)
      })
    return () => {
      active = false
    }
  }, [])

  const filteredViloyatlar = useMemo(() => {
    const q = searchTerm.trim().toLowerCase()
    if (!q) return viloyatlar
    return viloyatlar.filter((v) => v.name.toLowerCase().includes(q) || v.code.toLowerCase().includes(q))
  }, [viloyatlar, searchTerm])

  function toggleSet(setter: (fn: (prev: Set<string>) => Set<string>) => void, id: string) {
    setter((prev) => {
      const next = new Set(prev)
      if (next.has(id)) next.delete(id)
      else next.add(id)
      return next
    })
  }

  function setLoadingFlag(setter: (fn: (prev: Set<string>) => Set<string>) => void, id: string, isLoading: boolean) {
    setter((prev) => {
      const next = new Set(prev)
      if (isLoading) next.add(id)
      else next.delete(id)
      return next
    })
  }

  async function fetchTumanlarForViloyat(viloyatId: string) {
    if (tumanlar[viloyatId] !== undefined) return
    setLoadingFlag(setLoadingViloyatlar, viloyatId, true)
    try {
      const data = await listRegions({ type: 'district', parent_id: viloyatId })
      setTumanlar((prev) => ({ ...prev, [viloyatId]: data }))
      setAllTumanlar((prev) => {
        const map = new Map(prev.map((t) => [getRegionId(t), t]))
        data.forEach((t) => map.set(getRegionId(t), t))
        return [...map.values()]
      })
    } catch (err) {
      if (!(err instanceof ApiError)) toast.error('Tumanlarni yuklab bo‘lmadi')
    } finally {
      setLoadingFlag(setLoadingViloyatlar, viloyatId, false)
    }
  }

  async function fetchMfyForTuman(tumanId: string) {
    if (mfyList[tumanId] !== undefined) return
    setLoadingFlag(setLoadingTumanlar, tumanId, true)
    try {
      const data = await listRegions({ type: 'mfy', parent_id: tumanId })
      setMfyList((prev) => ({ ...prev, [tumanId]: data }))
    } catch (err) {
      if (!(err instanceof ApiError)) toast.error('MFYlarni yuklab bo‘lmadi')
    } finally {
      setLoadingFlag(setLoadingTumanlar, tumanId, false)
    }
  }

  async function handleToggleViloyat(viloyatId: string) {
    const willExpand = !expandedViloyatlar.has(viloyatId)
    toggleSet(setExpandedViloyatlar, viloyatId)
    if (willExpand) await fetchTumanlarForViloyat(viloyatId)
  }

  async function handleToggleTuman(tumanId: string) {
    const willExpand = !expandedTumanlar.has(tumanId)
    toggleSet(setExpandedTumanlar, tumanId)
    if (willExpand) await fetchMfyForTuman(tumanId)
  }

  async function ensureTumanlar() {
    if (allTumanlar.length) return
    const data = await listRegions({ type: 'district' })
    setAllTumanlar(data)
  }

  async function openCreate(type: UIRegionType) {
    try {
      if (type === REGION_TYPES.mfy) await ensureTumanlar()
      setFormType(type)
      setEditingItem(null)
      setFormOpen(true)
    } catch (err) {
      if (!(err instanceof ApiError)) toast.error('Tumanlarni yuklab bo‘lmadi')
    }
  }

  async function openEdit(item: RegionItem, type: UIRegionType) {
    try {
      if (type === REGION_TYPES.mfy) await ensureTumanlar()
      setFormType(type)
      setEditingItem(item)
      setFormOpen(true)
    } catch (err) {
      if (!(err instanceof ApiError)) toast.error('Tumanlarni yuklab bo‘lmadi')
    }
  }

  function closeForm() {
    if (saving) return
    setFormOpen(false)
    setEditingItem(null)
  }

  async function handleFormSubmit(payload: {
    name: string
    code: string
    status: RegionStatus
    parent_id: string | null
  }) {
    setSaving(true)
    try {
      const id = getRegionId(editingItem)
      const body = {
        name: payload.name,
        code: payload.code,
        status: payload.status,
        type: API_TYPE[formType],
        parent_id: payload.parent_id,
      }
      if (id) await updateRegion(id, body)
      else await createRegion(body)
      await loadData()
    } finally {
      setSaving(false)
    }
  }

  async function handleStatusToggle(item: RegionItem, type: UIRegionType) {
    try {
      await updateRegion(getRegionId(item), {
        name: item.name,
        code: item.code,
        type: API_TYPE[type],
        parent_id: item.parent_id,
        status: item.status === 'active' ? 'inactive' : 'active',
      })
      await loadData()
    } catch (err) {
      if (!(err instanceof ApiError)) toast.error('Holatni o‘zgartirib bo‘lmadi')
    }
  }

  async function confirmDelete() {
    if (!deleteTarget) return
    setDeleting(true)
    try {
      await deleteRegion(getRegionId(deleteTarget.item))
      setDeleteTarget(null)
      await loadData()
    } catch (err) {
      if (!(err instanceof ApiError)) toast.error('O‘chirib bo‘lmadi')
      setDeleteTarget(null)
    } finally {
      setDeleting(false)
    }
  }

  async function onImport() {
    setImporting(true)
    try {
      const result = await importRegions()
      toast.success(`Import: qo'shildi ${result.inserted}, yangilandi ${result.updated}`)
      await loadData()
      setImportOpen(false)
    } catch (err) {
      if (!(err instanceof ApiError)) toast.error('Import amalga oshmadi')
    } finally {
      setImporting(false)
    }
  }

  const deleteLabel =
    deleteTarget?.type === REGION_TYPES.viloyat ? 'viloyat' : deleteTarget?.type === REGION_TYPES.tuman ? 'tuman' : 'MFY'

  return (
    <motion.div initial={{ opacity: 0 }} animate={{ opacity: 1 }} className="mx-auto max-w-6xl space-y-6">
      <div className="flex flex-col gap-4 sm:flex-row sm:items-start sm:justify-between">
        <div>
          <h2 className="text-2xl font-bold text-slate-900">Hududlar</h2>
          <p className="mt-1 text-sm text-slate-500">Viloyat, tuman va MFYlarni boshqaring</p>
        </div>
        <div className="flex flex-wrap gap-2">
          <button
            type="button"
            onClick={() => setImportOpen(true)}
            disabled={importing}
            className="inline-flex items-center gap-2 rounded-xl border border-slate-200 bg-white px-3.5 py-2.5 text-sm font-semibold text-slate-700 disabled:opacity-60"
          >
            <Upload size={16} />
            {importing ? 'Import...' : 'JSON import'}
          </button>
          <motion.button
            type="button"
            whileHover={{ scale: 1.03 }}
            whileTap={{ scale: 0.97 }}
            onClick={() => {
              void openCreate(REGION_TYPES.viloyat)
            }}
            className="inline-flex items-center justify-center gap-2 self-start rounded-xl px-3.5 py-2.5 text-sm font-semibold text-white"
            style={{ backgroundColor: APP_COLOR }}
          >
            <Plus size={16} />
            Viloyat
          </motion.button>
          <motion.button
            type="button"
            whileHover={{ scale: 1.03 }}
            whileTap={{ scale: 0.97 }}
            onClick={() => {
              void openCreate(REGION_TYPES.tuman)
            }}
            className="inline-flex items-center justify-center gap-2 self-start rounded-xl px-3.5 py-2.5 text-sm font-semibold text-white"
            style={{ backgroundColor: APP_COLOR }}
          >
            <Plus size={16} />
            Tuman
          </motion.button>
          <motion.button
            type="button"
            whileHover={{ scale: 1.03 }}
            whileTap={{ scale: 0.97 }}
            onClick={() => {
              void openCreate(REGION_TYPES.mfy)
            }}
            className="inline-flex items-center justify-center gap-2 self-start rounded-xl px-3.5 py-2.5 text-sm font-semibold text-white"
            style={{ backgroundColor: APP_COLOR }}
          >
            <Plus size={16} />
            MFY
          </motion.button>
        </div>
      </div>

      <div className="rounded-2xl border border-slate-200 bg-white p-4">
        <div className="relative">
          <span className="pointer-events-none absolute inset-y-0 left-0 flex w-10 items-center justify-center text-slate-400">
            <Search size={18} />
          </span>
          <input
            type="search"
            value={searchTerm}
            onChange={(e) => setSearchTerm(e.target.value)}
            placeholder="Qidirish..."
            className="w-full rounded-xl border border-slate-200 py-2.5 pr-4 pl-10 text-sm outline-none focus:border-slate-400"
          />
        </div>
      </div>

      {loading ? (
        <div className="flex h-64 items-center justify-center rounded-2xl border border-slate-200 bg-white">
          <Loader2 className="animate-spin" size={28} style={{ color: APP_COLOR }} />
        </div>
      ) : (
        <div className="overflow-hidden rounded-2xl border border-slate-200 bg-white">
          {filteredViloyatlar.length === 0 ? (
            <p className="py-12 text-center text-slate-500">
              {searchTerm ? 'Qidiruv natijasi topilmadi' : 'Hozircha hududlar yo‘q'}
            </p>
          ) : (
            <RegionTree
              viloyatlar={filteredViloyatlar}
              tumanlar={tumanlar}
              mfyList={mfyList}
              expandedViloyatlar={expandedViloyatlar}
              expandedTumanlar={expandedTumanlar}
              loadingViloyatlar={loadingViloyatlar}
              loadingTumanlar={loadingTumanlar}
              onToggleViloyat={(id) => {
                void handleToggleViloyat(id)
              }}
              onToggleTuman={(id) => {
                void handleToggleTuman(id)
              }}
              onEdit={(item, type) => {
                void openEdit(item, type)
              }}
              onDelete={(item, type) => setDeleteTarget({ item, type })}
              onStatusToggle={(item, type) => {
                void handleStatusToggle(item, type)
              }}
            />
          )}
        </div>
      )}

      {formOpen ? (
        <RegionFormModal
          key={`${formType}-${getRegionId(editingItem) || 'new'}`}
          type={formType}
          item={editingItem}
          viloyatlar={viloyatlar}
          tumanlar={allTumanlar}
          saving={saving}
          onClose={closeForm}
          onSubmit={handleFormSubmit}
        />
      ) : null}

      <ConfirmModal
        open={Boolean(deleteTarget)}
        title={`${deleteLabel}ni o‘chirish`}
        onClose={() => {
          if (!deleting) setDeleteTarget(null)
        }}
        onConfirm={() => {
          void confirmDelete()
        }}
        confirmLabel="O‘chirish"
        loading={deleting}
      >
        <strong>{deleteTarget?.item?.name}</strong> ni o‘chirmoqchimisiz? Bu amalni qaytarib bo‘lmaydi.
      </ConfirmModal>

      <ConfirmModal
        open={importOpen}
        title="JSON import"
        confirmLabel="Import"
        tone="primary"
        loading={importing}
        onClose={() => {
          if (!importing) setImportOpen(false)
        }}
        onConfirm={() => {
          void onImport()
        }}
      >
        ttsa.regions.json dagi hududlar import qilinsinmi? Mavjud yozuvlar o'chirilmaydi.
      </ConfirmModal>
    </motion.div>
  )
}
