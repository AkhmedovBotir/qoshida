import { FormEvent, useEffect, useMemo, useState } from 'react'
import { APP_COLOR } from '../../constants/config'
import { listCategories } from '../../lib/categories'
import { ApiError } from '../../lib/api'
import { catalogImageSrc } from '../../lib/shopCatalog'
import type { CategoryItem } from '../../types/category'
import type { ProductUnit, ShopTemplateForm, ShopTemplateItem } from '../../types/shopCatalog'
import { ImageThumb } from '../ui/ImageLightbox'
import { Select } from '../ui/Select'
import { toast } from '../../lib/snack'

type Props = {
  item: ShopTemplateItem | null
  saving: boolean
  onClose: () => void
  onSubmit: (payload: ShopTemplateForm) => Promise<void>
}

const empty: ShopTemplateForm = {
  category_id: '',
  subcategory_id: '',
  name: '',
  description: '',
  unit: 'dona',
  unit_size: '1',
  images: [],
}

function readImage(file: File): Promise<string> {
  return new Promise((resolve, reject) => {
    if (file.size > 10 * 1024 * 1024) {
      reject(new Error('Rasm 10 MB dan oshmasin'))
      return
    }
    const reader = new FileReader()
    reader.onload = () => resolve(String(reader.result ?? ''))
    reader.onerror = () => reject(new Error('Rasmni o‘qib bo‘lmadi'))
    reader.readAsDataURL(file)
  })
}

export function TemplateFormModal({ item, saving, onClose, onSubmit }: Props) {
  const isEdit = Boolean(item)
  const [form, setForm] = useState<ShopTemplateForm>(() =>
    item
      ? {
          category_id: item.category_id,
          subcategory_id: item.subcategory_id,
          name: item.name,
          description: item.description,
          unit: item.unit,
          unit_size: String(item.unit_size),
          images: item.images ?? [],
        }
      : empty,
  )
  const [categories, setCategories] = useState<CategoryItem[]>([])

  useEffect(() => {
    void listCategories().then(setCategories).catch(() => setCategories([]))
  }, [])

  const roots = useMemo(() => categories.filter((c) => !c.parent_id), [categories])
  const children = useMemo(
    () => categories.filter((c) => c.parent_id === form.category_id),
    [categories, form.category_id],
  )

  async function handleImages(files: FileList | null) {
    if (!files?.length) return
    try {
      const next = [...form.images]
      for (const file of Array.from(files)) {
        if (next.length >= 5) break
        next.push(await readImage(file))
      }
      setForm((f) => ({ ...f, images: next.slice(0, 5) }))
    } catch (err) {
      toast.error(err instanceof Error ? err.message : 'Rasm qo‘shib bo‘lmadi')
    }
  }

  async function handleSubmit(event: FormEvent) {
    event.preventDefault()
    if (!form.category_id || !form.subcategory_id) {
      toast.error('Kategoriya va subkategoriyani tanlang')
      return
    }
    try {
      await onSubmit(form)
      onClose()
    } catch (err) {
      if (!(err instanceof ApiError)) toast.error('Saqlab bo‘lmadi')
    }
  }

  return (
    <div className="fixed inset-0 z-50 flex items-center justify-center bg-slate-950/40 px-4">
      <form onSubmit={handleSubmit} className="max-h-[92vh] w-full max-w-3xl overflow-y-auto rounded-2xl bg-white p-5 shadow-xl">
        <h3 className="text-lg font-bold text-slate-900">{isEdit ? 'Shablonni tahrirlash' : 'Yangi mahsulot shabloni'}</h3>
        <p className="mt-1 text-sm text-slate-500">Mahalla do‘konlari shu shablondan mahsulot biriktiradi. Rasmlar 5 tagacha.</p>

        <div className="mt-4 grid gap-3 sm:grid-cols-2">
          <Select
            label="Kategoriya"
            value={form.category_id}
            onChange={(v) => setForm((f) => ({ ...f, category_id: v, subcategory_id: '' }))}
            required
            searchable
            placeholder="Kategoriyani tanlang"
            options={roots.map((cat) => ({ value: cat.id, label: cat.name }))}
          />
          <Select
            label="Subkategoriya"
            value={form.subcategory_id}
            onChange={(v) => setForm((f) => ({ ...f, subcategory_id: v }))}
            required
            searchable
            placeholder="Subkategoriyani tanlang"
            emptyText={form.category_id ? 'Subkategoriya yo‘q' : 'Avval kategoriya tanlang'}
            options={children.map((cat) => ({ value: cat.id, label: cat.name }))}
          />
          <label className="block text-sm font-medium text-slate-700 sm:col-span-2">
            Nomi
            <input
              value={form.name}
              onChange={(e) => setForm((f) => ({ ...f, name: e.target.value }))}
              required
              className="mt-1.5 w-full rounded-xl border border-slate-200 px-3 py-2.5 outline-none focus:border-slate-400"
            />
          </label>
          <label className="block text-sm font-medium text-slate-700 sm:col-span-2">
            Tavsif
            <textarea
              value={form.description}
              onChange={(e) => setForm((f) => ({ ...f, description: e.target.value }))}
              rows={3}
              className="mt-1.5 w-full rounded-xl border border-slate-200 px-3 py-2.5 outline-none focus:border-slate-400"
            />
          </label>
          <Select
            label="Birlik"
            value={form.unit}
            onChange={(v) => setForm((f) => ({ ...f, unit: v as ProductUnit }))}
            options={[
              { value: 'dona', label: 'Dona' },
              { value: 'litr', label: 'Litr' },
              { value: 'kg', label: 'Kg' },
            ]}
          />
          <label className="block text-sm font-medium text-slate-700">
            Birlik o‘lchami
            <input
              type="number"
              min={0.001}
              step="0.001"
              value={form.unit_size}
              onChange={(e) => setForm((f) => ({ ...f, unit_size: e.target.value }))}
              required
              className="mt-1.5 w-full rounded-xl border border-slate-200 px-3 py-2.5 outline-none focus:border-slate-400"
            />
          </label>
          <div className="sm:col-span-2">
            <p className="text-sm font-medium text-slate-700">Rasmlar ({form.images.length}/5)</p>
            <div className="mt-2 flex flex-wrap gap-2">
              {form.images.map((src, index) => (
                <div key={`${src}-${index}`} className="relative h-20 w-20 overflow-hidden rounded-xl border border-slate-200">
                  <ImageThumb
                    src={catalogImageSrc(src)}
                    images={form.images.map((item) => catalogImageSrc(item))}
                    className="h-20 w-20 object-cover"
                  />
                  <button
                    type="button"
                    onClick={() => setForm((f) => ({ ...f, images: f.images.filter((_, i) => i !== index) }))}
                    className="absolute right-1 top-1 z-10 rounded bg-white/90 px-1 text-xs text-red-600"
                  >
                    ×
                  </button>
                </div>
              ))}
              {form.images.length < 5 ? (
                <label className="flex h-20 w-20 cursor-pointer items-center justify-center rounded-xl border border-dashed border-slate-300 text-xs text-slate-500">
                  + Rasm
                  <input type="file" accept="image/png,image/jpeg,image/webp" hidden multiple onChange={(e) => void handleImages(e.target.files)} />
                </label>
              ) : null}
            </div>
          </div>
        </div>

        <div className="mt-5 flex justify-end gap-2">
          <button type="button" onClick={onClose} className="rounded-xl border border-slate-200 px-4 py-2 text-sm">
            Bekor qilish
          </button>
          <button
            type="submit"
            disabled={saving}
            className="rounded-xl px-4 py-2 text-sm font-semibold text-white disabled:opacity-60"
            style={{ backgroundColor: APP_COLOR }}
          >
            {saving ? 'Saqlanmoqda...' : 'Saqlash'}
          </button>
        </div>
      </form>
    </div>
  )
}
