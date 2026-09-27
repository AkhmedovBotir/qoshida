import { Eye, Loader2, Pencil, Plus, Search, Trash2, Upload } from 'lucide-react'
import { motion } from 'motion/react'
import { useEffect, useMemo, useState } from 'react'
import { ActivityTypeFormModal } from '../components/activity/ActivityTypeFormModal'
import { StatusToggle } from '../components/regions/StatusToggle'
import { ConfirmModal } from '../components/ui/ConfirmModal'
import { ViewModal } from '../components/ui/ViewModal'
import { APP_COLOR } from '../constants/config'
import { getActivityIcon } from '../lib/activityIcons'
import {
  createActivityType,
  deleteActivityType,
  importActivityTypes,
  listActivityTypes,
  updateActivityType,
} from '../lib/activityTypes'
import { ApiError } from '../lib/api'
import { getActivityTypeId, type ActivityStatus, type ActivityTypeItem } from '../types/activityType'
import { toast } from '../lib/snack'

export function ActivityTypesPage() {
  const [items, setItems] = useState<ActivityTypeItem[]>([])
  const [loading, setLoading] = useState(true)
  const [searchTerm, setSearchTerm] = useState('')
  const [formOpen, setFormOpen] = useState(false)
  const [editingItem, setEditingItem] = useState<ActivityTypeItem | null>(null)
  const [viewing, setViewing] = useState<ActivityTypeItem | null>(null)
  const [saving, setSaving] = useState(false)
  const [importing, setImporting] = useState(false)
  const [importOpen, setImportOpen] = useState(false)
  const [deleteTarget, setDeleteTarget] = useState<ActivityTypeItem | null>(null)
  const [deleting, setDeleting] = useState(false)

  async function loadData() {
    try {
      setItems(await listActivityTypes())
    } catch (err) {
      if (!(err instanceof ApiError)) toast.error('Ma’lumotlarni yuklab bo‘lmadi')
    } finally {
      setLoading(false)
    }
  }

  useEffect(() => {
    void loadData()
  }, [])

  const filtered = useMemo(() => {
    const q = searchTerm.trim().toLowerCase()
    if (!q) return items
    return items.filter((item) => item.name.toLowerCase().includes(q) || item.icon.toLowerCase().includes(q))
  }, [items, searchTerm])

  function openCreate() {
    setEditingItem(null)
    setFormOpen(true)
  }

  function openEdit(item: ActivityTypeItem) {
    setEditingItem(item)
    setFormOpen(true)
  }

  function closeForm() {
    if (saving) return
    setFormOpen(false)
    setEditingItem(null)
  }

  async function handleFormSubmit(payload: { name: string; icon: string; status: ActivityStatus }) {
    setSaving(true)
    try {
      const id = getActivityTypeId(editingItem)
      if (id) await updateActivityType(id, payload)
      else await createActivityType(payload)
      await loadData()
    } finally {
      setSaving(false)
    }
  }

  async function handleStatusToggle(item: ActivityTypeItem) {
    try {
      await updateActivityType(item.id, {
        name: item.name,
        icon: item.icon,
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
      await deleteActivityType(deleteTarget.id)
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
      const result = await importActivityTypes()
      toast.success(`Import: qo'shildi ${result.inserted}, yangilandi ${result.updated}`)
      await loadData()
      setImportOpen(false)
    } catch (err) {
      if (!(err instanceof ApiError)) toast.error('Import amalga oshmadi')
    } finally {
      setImporting(false)
    }
  }

  return (
    <motion.div initial={{ opacity: 0 }} animate={{ opacity: 1 }} className="mx-auto max-w-6xl space-y-6">
      <div className="flex flex-col gap-4 sm:flex-row sm:items-start sm:justify-between">
        <div>
          <h2 className="text-2xl font-bold text-slate-900">Faoliyat turlari</h2>
          <p className="mt-1 text-sm text-slate-500">Kontragentlar uchun faoliyat turlarini boshqaring</p>
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
            onClick={openCreate}
            className="inline-flex items-center justify-center gap-2 self-start rounded-xl px-3.5 py-2.5 text-sm font-semibold text-white"
            style={{ backgroundColor: APP_COLOR }}
          >
            <Plus size={16} />
            Qo‘shish
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
      ) : filtered.length === 0 ? (
        <p className="rounded-2xl border border-slate-200 bg-white py-12 text-center text-slate-500">
          {searchTerm ? 'Qidiruv natijasi topilmadi' : 'Hozircha faoliyat turlari yo‘q'}
        </p>
      ) : (
        <div className="grid gap-4 sm:grid-cols-2 xl:grid-cols-3">
          {filtered.map((item) => {
            const Icon = getActivityIcon(item.icon)
            const active = item.status === 'active'
            return (
              <article
                key={item.id}
                className={`rounded-2xl border border-slate-200 bg-white p-4 ${active ? '' : 'opacity-60'}`}
              >
                <div className="flex items-start gap-3">
                  <span
                    className="flex h-12 w-12 shrink-0 items-center justify-center rounded-xl text-white"
                    style={{ backgroundColor: APP_COLOR }}
                  >
                    <Icon size={22} />
                  </span>
                  <div className="min-w-0 flex-1">
                    <h3 className="font-bold text-slate-900">{item.name}</h3>
                  </div>
                </div>
                <div className="mt-4 flex items-center justify-between gap-2">
                  <div className="flex items-center gap-2">
                    <span className={`text-xs font-medium ${active ? 'text-slate-700' : 'text-slate-500'}`}>
                      {active ? 'Faol' : 'Nofaol'}
                    </span>
                    <StatusToggle
                      active={active}
                      onChange={() => {
                        void handleStatusToggle(item)
                      }}
                      size="sm"
                    />
                  </div>
                  <div className="flex gap-1.5">
                    <button
                      type="button"
                      title="Ko‘rish"
                      aria-label="Ko‘rish"
                      onClick={() => setViewing(item)}
                      className="rounded-lg border border-slate-200 p-2 text-slate-600"
                    >
                      <Eye size={15} />
                    </button>
                    <button
                      type="button"
                      title="Tahrirlash"
                      aria-label="Tahrirlash"
                      onClick={() => openEdit(item)}
                      className="rounded-lg border border-slate-200 p-2 text-slate-600"
                    >
                      <Pencil size={15} />
                    </button>
                    <button
                      type="button"
                      title="O‘chirish"
                      aria-label="O‘chirish"
                      onClick={() => setDeleteTarget(item)}
                      className="rounded-lg border border-red-100 p-2 text-red-600"
                    >
                      <Trash2 size={15} />
                    </button>
                  </div>
                </div>
              </article>
            )
          })}
        </div>
      )}

      {viewing ? (
        <ViewModal
          title="Faoliyat turi"
          fields={[
            { label: 'Nomi', value: viewing.name },
            { label: 'Icon', value: viewing.icon },
            { label: 'Holat', value: viewing.status === 'active' ? 'Faol' : 'Nofaol' },
          ]}
          onClose={() => setViewing(null)}
          onEdit={() => {
            openEdit(viewing)
            setViewing(null)
          }}
        />
      ) : null}

      {formOpen ? (
        <ActivityTypeFormModal
          key={getActivityTypeId(editingItem) || 'new'}
          item={editingItem}
          saving={saving}
          onClose={closeForm}
          onSubmit={handleFormSubmit}
        />
      ) : null}

      <ConfirmModal
        open={Boolean(deleteTarget)}
        title="Faoliyat turini o‘chirish"
        onClose={() => {
          if (!deleting) setDeleteTarget(null)
        }}
        onConfirm={() => {
          void confirmDelete()
        }}
        confirmLabel="O‘chirish"
        loading={deleting}
      >
        <strong>{deleteTarget?.name}</strong> ni o‘chirmoqchimisiz? Bu amalni qaytarib bo‘lmaydi.
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
        ttsa.contragenttypes.json dagi faoliyat turlari import qilinsinmi? Mavjud yozuvlar o'chirilmaydi.
      </ConfirmModal>
    </motion.div>
  )
}
