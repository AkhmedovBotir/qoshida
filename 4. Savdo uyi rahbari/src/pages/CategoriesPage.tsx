import { Loader2, Plus, Search } from 'lucide-react'
import { motion } from 'motion/react'
import { useEffect, useMemo, useState } from 'react'
import { CategoryFormModal } from '../components/categories/CategoryFormModal'
import { CategoryTree } from '../components/categories/CategoryTree'
import { ConfirmModal } from '../components/ui/ConfirmModal'
import { APP_COLOR } from '../constants/config'
import { ApiError } from '../lib/api'
import { createCategory, deleteCategory, listCategories, updateCategory, type CategoryWrite } from '../lib/categories'
import { getCategoryId, type CategoryItem, type CategoryStatus } from '../types/category'
import { toast } from '../lib/snack'

export function CategoriesPage() {
  const [items, setItems] = useState<CategoryItem[]>([])
  const [loading, setLoading] = useState(true)
  const [searchTerm, setSearchTerm] = useState('')
  const [expanded, setExpanded] = useState<Set<string>>(new Set())
  const [formOpen, setFormOpen] = useState(false)
  const [formKind, setFormKind] = useState<'root' | 'child'>('root')
  const [editingItem, setEditingItem] = useState<CategoryItem | null>(null)
  const [saving, setSaving] = useState(false)

  const [deleteTarget, setDeleteTarget] = useState<CategoryItem | null>(null)
  const [deleting, setDeleting] = useState(false)

  async function loadData() {
    try {
      setItems(await listCategories())
    } catch (err) {
      if (!(err instanceof ApiError)) toast.error('Ma’lumotlarni yuklab bo‘lmadi')
    } finally {
      setLoading(false)
    }
  }

  useEffect(() => {
    void loadData()
  }, [])

  const roots = useMemo(() => items.filter((item) => !item.parent_id), [items])
  const childrenByParent = useMemo(() => {
    const map: Record<string, CategoryItem[]> = {}
    for (const item of items) {
      if (!item.parent_id) continue
      if (!map[item.parent_id]) map[item.parent_id] = []
      map[item.parent_id].push(item)
    }
    return map
  }, [items])

  const filteredRoots = useMemo(() => {
    const q = searchTerm.trim().toLowerCase()
    if (!q) return roots
    return roots.filter((root) => {
      if (root.name.toLowerCase().includes(q) || root.slug.toLowerCase().includes(q)) return true
      return (childrenByParent[root.id] || []).some(
        (child) => child.name.toLowerCase().includes(q) || child.slug.toLowerCase().includes(q),
      )
    })
  }, [roots, childrenByParent, searchTerm])

  function openCreate(kind: 'root' | 'child') {
    setFormKind(kind)
    setEditingItem(null)
    setFormOpen(true)
  }

  function openEdit(item: CategoryItem, kind: 'root' | 'child') {
    setFormKind(kind)
    setEditingItem(item)
    setFormOpen(true)
  }

  function closeForm() {
    if (saving) return
    setFormOpen(false)
    setEditingItem(null)
  }

  async function handleFormSubmit(payload: CategoryWrite) {
    setSaving(true)
    try {
      const id = getCategoryId(editingItem)
      if (id) await updateCategory(id, payload)
      else await createCategory(payload)
      await loadData()
    } finally {
      setSaving(false)
    }
  }

  async function handleStatusToggle(item: CategoryItem) {
    try {
      await updateCategory(item.id, {
        name: item.name,
        slug: item.slug,
        parent_id: item.parent_id,
        censored: item.censored,
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
      await deleteCategory(deleteTarget.id)
      setDeleteTarget(null)
      await loadData()
    } catch (err) {
      if (!(err instanceof ApiError)) toast.error('O‘chirib bo‘lmadi')
    } finally {
      setDeleting(false)
    }
  }

  return (
    <motion.div initial={{ opacity: 0 }} animate={{ opacity: 1 }} className="mx-auto max-w-6xl space-y-6">
      <div className="flex flex-col gap-4 sm:flex-row sm:items-start sm:justify-between">
        <div>
          <h2 className="text-2xl font-bold text-slate-900">Kategoriyalar</h2>
          <p className="mt-1 text-sm text-slate-500">Asosiy va ichki kategoriyalarni boshqaring</p>
        </div>
        <div className="flex flex-wrap gap-2">
          <motion.button
            type="button"
            whileHover={{ scale: 1.03 }}
            whileTap={{ scale: 0.97 }}
            onClick={() => openCreate('root')}
            className="inline-flex items-center justify-center gap-2 self-start rounded-xl px-3.5 py-2.5 text-sm font-semibold text-white"
            style={{ backgroundColor: APP_COLOR }}
          >
            <Plus size={16} />
            Asosiy
          </motion.button>
          <motion.button
            type="button"
            whileHover={{ scale: 1.03 }}
            whileTap={{ scale: 0.97 }}
            onClick={() => openCreate('child')}
            className="inline-flex items-center justify-center gap-2 self-start rounded-xl px-3.5 py-2.5 text-sm font-semibold text-white"
            style={{ backgroundColor: APP_COLOR }}
          >
            <Plus size={16} />
            Ichki
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
          {filteredRoots.length === 0 ? (
            <p className="py-12 text-center text-slate-500">
              {searchTerm ? 'Qidiruv natijasi topilmadi' : 'Hozircha kategoriyalar yo‘q'}
            </p>
          ) : (
            <CategoryTree
              roots={filteredRoots}
              childrenByParent={childrenByParent}
              expanded={expanded}
              onToggle={(id) => {
                setExpanded((prev) => {
                  const next = new Set(prev)
                  if (next.has(id)) next.delete(id)
                  else next.add(id)
                  return next
                })
              }}
              onEdit={openEdit}
              onDelete={setDeleteTarget}
              onStatusToggle={(item) => {
                void handleStatusToggle(item)
              }}
            />
          )}
        </div>
      )}

      {formOpen ? (
        <CategoryFormModal
          key={`${formKind}-${getCategoryId(editingItem) || 'new'}`}
          kind={formKind}
          item={editingItem}
          roots={roots}
          saving={saving}
          onClose={closeForm}
          onSubmit={handleFormSubmit}
        />
      ) : null}

      <ConfirmModal
        open={Boolean(deleteTarget)}
        title="Kategoriyani o‘chirish"
        confirmLabel="O‘chirish"
        loading={deleting}
        onClose={() => {
          if (!deleting) setDeleteTarget(null)
        }}
        onConfirm={() => {
          void confirmDelete()
        }}
      >
        <strong>{deleteTarget?.name}</strong> ni o‘chirmoqchimisiz? Bu amalni qaytarib bo‘lmaydi.
      </ConfirmModal>
    </motion.div>
  )
}
