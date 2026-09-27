import { ArrowDown, ArrowUp, Loader2, Plus, Search } from 'lucide-react'
import { motion } from 'motion/react'
import { useCallback, useEffect, useState } from 'react'
import { StatusToggle } from '../regions/StatusToggle'
import { ConfirmModal } from '../ui/ConfirmModal'
import { Select } from '../ui/Select'
import { TableActions } from '../ui/TableActions'
import { ViewModal } from '../ui/ViewModal'
import { APP_COLOR } from '../../constants/config'
import {
  createCommentTemplate,
  deleteCommentTemplate,
  listCommentTemplates,
  reorderCommentTemplates,
  updateCommentTemplate,
} from '../../lib/commentTemplates'
import { ApiError } from '../../lib/api'
import { getCommentTemplateId, type CommentStatus, type CommentTemplateItem } from '../../types/commentTemplate'
import { CommentTemplateFormModal } from './CommentTemplateFormModal'
import { toast } from '../../lib/snack'

export function CommentTemplatesTab() {
  const [items, setItems] = useState<CommentTemplateItem[]>([])
  const [loading, setLoading] = useState(true)
  const [searchTerm, setSearchTerm] = useState('')
  const [appliedSearch, setAppliedSearch] = useState('')
  const [page, setPage] = useState(1)
  const [limit, setLimit] = useState(10)
  const [total, setTotal] = useState(0)
  const [totalPages, setTotalPages] = useState(1)
  const [formOpen, setFormOpen] = useState(false)
  const [viewing, setViewing] = useState<CommentTemplateItem | null>(null)
  const [editingItem, setEditingItem] = useState<CommentTemplateItem | null>(null)
  const [saving, setSaving] = useState(false)
  const [busyId, setBusyId] = useState('')
  const [deleteTarget, setDeleteTarget] = useState<CommentTemplateItem | null>(null)
  const [deleting, setDeleting] = useState(false)

  const loadData = useCallback(async () => {
    try {
      const data = await listCommentTemplates({ page, limit, q: appliedSearch })
      setItems(data.items ?? [])
      setTotal(data.total)
      setTotalPages(Math.max(1, data.total_pages))
    } catch (err) {
      if (!(err instanceof ApiError)) toast.error('Ma’lumotlarni yuklab bo‘lmadi')
    } finally {
      setLoading(false)
    }
  }, [page, limit, appliedSearch])

  useEffect(() => {
    setLoading(true)
    void loadData()
  }, [loadData])

  function openCreate() {
    setEditingItem(null)
    setFormOpen(true)
  }

  function openEdit(item: CommentTemplateItem) {
    setEditingItem(item)
    setFormOpen(true)
  }

  function closeForm() {
    if (saving) return
    setFormOpen(false)
    setEditingItem(null)
  }

  async function handleFormSubmit(payload: { comment: string; status: CommentStatus }) {
    setSaving(true)
    try {
      const id = getCommentTemplateId(editingItem)
      if (id) await updateCommentTemplate(id, payload)
      else await createCommentTemplate(payload)
      await loadData()
    } finally {
      setSaving(false)
    }
  }

  async function handleStatusToggle(item: CommentTemplateItem) {
    try {
      await updateCommentTemplate(item.id, {
        comment: item.comment,
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
      await deleteCommentTemplate(deleteTarget.id)
      setDeleteTarget(null)
      await loadData()
    } catch (err) {
      if (!(err instanceof ApiError)) toast.error('O‘chirib bo‘lmadi')
      setDeleteTarget(null)
    } finally {
      setDeleting(false)
    }
  }

  async function moveRow(index: number, direction: 'up' | 'down') {
    const other = direction === 'up' ? index - 1 : index + 1
    if (other < 0 || other >= items.length) return
    const from = items[index]
    const to = items[other]
    setBusyId(from.id)
    try {
      await reorderCommentTemplates(from.id, to.id)
      await loadData()
    } catch (err) {
      if (!(err instanceof ApiError)) toast.error('Tartibni o‘zgartirib bo‘lmadi')
    } finally {
      setBusyId('')
    }
  }

  function applySearch() {
    setPage(1)
    setAppliedSearch(searchTerm.trim())
  }

  return (
    <div className="space-y-4">
      <div className="flex flex-col gap-4 sm:flex-row sm:items-start sm:justify-between">
        <div>
          <h3 className="text-lg font-bold text-slate-900">Kommentariya shablonlari</h3>
          <p className="mt-1 text-sm text-slate-500">Mijozlar uchun tayyor izoh matnlarini boshqaring</p>
        </div>
        <motion.button
          type="button"
          whileHover={{ scale: 1.03 }}
          whileTap={{ scale: 0.97 }}
          onClick={openCreate}
          className="inline-flex items-center justify-center gap-2 self-start rounded-xl px-3.5 py-2.5 text-sm font-semibold text-white"
          style={{ backgroundColor: APP_COLOR }}
        >
          <Plus size={16} />
          Yangi shablon
        </motion.button>
      </div>

      <div className="rounded-2xl border border-slate-200 bg-white p-4">
        <div className="flex flex-col gap-2 sm:flex-row">
          <div className="relative flex-1">
            <span className="pointer-events-none absolute inset-y-0 left-0 flex w-10 items-center justify-center text-slate-400">
              <Search size={18} />
            </span>
            <input
              type="search"
              value={searchTerm}
              onChange={(e) => setSearchTerm(e.target.value)}
              onKeyDown={(e) => {
                if (e.key === 'Enter') applySearch()
              }}
              placeholder="Shablon matnini qidirish..."
              className="w-full rounded-xl border border-slate-200 py-2.5 pr-4 pl-10 text-sm outline-none focus:border-slate-400"
            />
          </div>
          <button
            type="button"
            onClick={applySearch}
            className="rounded-xl px-4 py-2.5 text-sm font-semibold text-white"
            style={{ backgroundColor: APP_COLOR }}
          >
            Qidirish
          </button>
        </div>
      </div>

      {loading ? (
        <div className="flex h-64 items-center justify-center rounded-2xl border border-slate-200 bg-white">
          <Loader2 className="animate-spin" size={28} style={{ color: APP_COLOR }} />
        </div>
      ) : items.length === 0 ? (
        <p className="rounded-2xl border border-slate-200 bg-white py-12 text-center text-slate-500">
          {appliedSearch ? 'Qidiruv natijasi topilmadi' : 'Hozircha shablonlar yo‘q'}
        </p>
      ) : (
        <div className="overflow-x-auto rounded-2xl border border-slate-200 bg-white">
          <table className="min-w-full text-left text-sm">
            <thead className="bg-slate-50 text-slate-500">
              <tr>
                <th className="w-24 px-4 py-3 font-medium">Tartib</th>
                <th className="min-w-[280px] px-4 py-3 font-medium">Kommentariya</th>
                <th className="px-4 py-3 font-medium">Holat</th>
                <th className="px-4 py-3 text-right font-medium">Amallar</th>
              </tr>
            </thead>
            <tbody>
              {items.map((item, index) => {
                const active = item.status === 'active'
                return (
                  <tr key={item.id} className="border-t border-slate-100">
                    <td className="px-4 py-3 text-slate-600">{item.sort_order}</td>
                    <td className="px-4 py-3 whitespace-pre-wrap text-slate-900">{item.comment}</td>
                    <td className="px-4 py-3">
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
                    </td>
                    <td className="px-4 py-3">
                      <div className="flex justify-end gap-1">
                        <button
                          type="button"
                          disabled={busyId === item.id || index === 0}
                          onClick={() => void moveRow(index, 'up')}
                          className="rounded-lg border border-slate-200 p-2 text-slate-600 disabled:opacity-40"
                          title="Yuqoriga"
                        >
                          <ArrowUp size={14} />
                        </button>
                        <button
                          type="button"
                          disabled={busyId === item.id || index === items.length - 1}
                          onClick={() => void moveRow(index, 'down')}
                          className="rounded-lg border border-slate-200 p-2 text-slate-600 disabled:opacity-40"
                          title="Pastga"
                        >
                          <ArrowDown size={14} />
                        </button>
                        <TableActions
                          onView={() => setViewing(item)}
                          onEdit={() => openEdit(item)}
                          onDelete={() => setDeleteTarget(item)}
                        />
                      </div>
                    </td>
                  </tr>
                )
              })}
            </tbody>
          </table>
        </div>
      )}

      <div className="flex flex-col gap-3 rounded-2xl border border-slate-200 bg-white px-4 py-3 sm:flex-row sm:items-center sm:justify-between">
        <p className="text-sm text-slate-600">
          Jami: <span className="font-semibold text-slate-900">{total}</span>
        </p>
        <div className="flex flex-wrap items-center gap-2">
          <Select
            value={String(limit)}
            onChange={(v) => {
              setLimit(Number(v))
              setPage(1)
            }}
            options={[
              { value: '10', label: '10 ta' },
              { value: '20', label: '20 ta' },
              { value: '50', label: '50 ta' },
            ]}
          />
          <button
            type="button"
            disabled={page <= 1}
            onClick={() => setPage((value) => value - 1)}
            className="rounded-xl border border-slate-200 px-3 py-2 text-sm disabled:opacity-50"
          >
            Oldingi
          </button>
          <span className="text-sm text-slate-600">
            {page} / {totalPages}
          </span>
          <button
            type="button"
            disabled={page >= totalPages}
            onClick={() => setPage((value) => value + 1)}
            className="rounded-xl border border-slate-200 px-3 py-2 text-sm disabled:opacity-50"
          >
            Keyingi
          </button>
        </div>
      </div>

      {viewing ? (
        <ViewModal
          title="Kommentariya shabloni"
          fields={[
            { label: 'Tartib', value: viewing.sort_order },
            { label: 'Kommentariya', value: viewing.comment },
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
        <CommentTemplateFormModal
          key={getCommentTemplateId(editingItem) || 'new'}
          item={editingItem}
          saving={saving}
          onClose={closeForm}
          onSubmit={handleFormSubmit}
        />
      ) : null}

      <ConfirmModal
        open={Boolean(deleteTarget)}
        title="Shablonni o‘chirish"
        onClose={() => {
          if (!deleting) setDeleteTarget(null)
        }}
        onConfirm={() => {
          void confirmDelete()
        }}
        confirmLabel="O‘chirish"
        loading={deleting}
      >
        «{deleteTarget?.comment || 'Ushbu shablon'}» butunlay o‘chiriladi. Davom etasizmi?
      </ConfirmModal>
    </div>
  )
}
