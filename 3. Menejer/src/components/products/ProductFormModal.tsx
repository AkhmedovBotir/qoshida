import { FormEvent, useEffect, useMemo, useState } from 'react'
import { APP_COLOR } from '../../constants/config'
import { listCategories } from '../../lib/categories'
import { listKontragents } from '../../lib/kontragents'
import { ApiError } from '../../lib/api'
import { commissionAmount, formatMoney, productImageSrc } from '../../lib/products'
import type { CategoryItem } from '../../types/category'
import type { KontragentItem } from '../../types/kontragent'
import type { ProductForm, ProductItem, ProductStatus, ProductUnit } from '../../types/product'
import { ImageThumb } from '../ui/ImageLightbox'
import { Select } from '../ui/Select'
import { toast } from '../../lib/snack'

type ProductFormModalProps = {
  item: ProductItem | null
  saving: boolean
  lockKontragent?: boolean
  onClose: () => void
  onSubmit: (payload: ProductForm) => Promise<void>
}

const empty: ProductForm = {
  kontragent_id: '',
  category_id: '',
  subcategory_id: '',
  name: '',
  description: '',
  sale_price: '',
  cost_price: '',
  quantity: '',
  unit: 'dona',
  unit_size: '1',
  commission_percent: '0',
  images: [],
  status: 'active',
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

export function ProductFormModal({ item, saving, lockKontragent, onClose, onSubmit }: ProductFormModalProps) {
  const isEdit = Boolean(item)
  const [form, setForm] = useState<ProductForm>(() =>
    item
      ? {
          kontragent_id: item.kontragent_id,
          category_id: item.category_id,
          subcategory_id: item.subcategory_id,
          name: item.name,
          description: item.description,
          sale_price: String(item.sale_price),
          cost_price: String(item.cost_price),
          quantity: String(item.quantity),
          unit: item.unit,
          unit_size: String(item.unit_size),
          commission_percent: String(item.commission_percent),
          images: item.images ?? [],
          status: item.status,
        }
      : empty,
  )
  const [kontragents, setKontragents] = useState<KontragentItem[]>([])
  const [categories, setCategories] = useState<CategoryItem[]>([])

  useEffect(() => {
    if (!lockKontragent) {
      void listKontragents({ page: 1, limit: 100 })
        .then((data) => setKontragents(data.items ?? []))
        .catch(() => setKontragents([]))
    }
    void listCategories().then(setCategories).catch(() => setCategories([]))
  }, [lockKontragent])

  const roots = useMemo(() => categories.filter((c) => !c.parent_id), [categories])
  const children = useMemo(
    () => categories.filter((c) => c.parent_id === form.category_id),
    [categories, form.category_id],
  )

  const sale = Number(form.sale_price) || 0
  const cost = Number(form.cost_price) || 0
  const percent = Number(form.commission_percent) || 0
  const commission = commissionAmount(sale, cost, percent)
  const margin = Math.max(0, sale - cost)

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
    if (!lockKontragent && !form.kontragent_id) {
      toast.error('Kontragent tanlang')
      return
    }
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
        <h3 className="text-lg font-bold text-slate-900">{isEdit ? 'Mahsulotni tahrirlash' : 'Yangi mahsulot'}</h3>
        <p className="mt-1 text-sm text-slate-500">Komissiya sotuv va asl narx farqining foizi. Rasmlar ixtiyoriy, 5 tagacha.</p>

        <div className="mt-4 grid gap-3 sm:grid-cols-2">
          {!lockKontragent ? (
            <Select
              label="Kontragent"
              className="sm:col-span-2"
              value={form.kontragent_id}
              onChange={(v) => setForm((f) => ({ ...f, kontragent_id: v }))}
              required
              searchable
              placeholder="Kontragentni tanlang"
              options={kontragents.map((item) => ({ value: item.id, label: item.name }))}
            />
          ) : null}
          <Select
            label="Kategoriya"
            value={form.category_id}
            onChange={(v) => setForm((f) => ({ ...f, category_id: v, subcategory_id: '' }))}
            required
            searchable
            placeholder="Kategoriyani tanlang"
            options={roots.map((item) => ({ value: item.id, label: item.name }))}
          />
          <Select
            label="Subkategoriya"
            value={form.subcategory_id}
            onChange={(v) => setForm((f) => ({ ...f, subcategory_id: v }))}
            required
            searchable
            placeholder="Subkategoriyani tanlang"
            emptyText={form.category_id ? 'Subkategoriya yo‘q' : 'Avval kategoriya tanlang'}
            options={children.map((item) => ({ value: item.id, label: item.name }))}
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
          <NumberField label="Sotuv narxi" value={form.sale_price} onChange={(v) => setForm((f) => ({ ...f, sale_price: v }))} />
          <NumberField label="Asl narxi" value={form.cost_price} onChange={(v) => setForm((f) => ({ ...f, cost_price: v }))} />
          <NumberField label="Miqdori" value={form.quantity} onChange={(v) => setForm((f) => ({ ...f, quantity: v }))} />
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
          <NumberField label="Birlik o‘lchami" value={form.unit_size} onChange={(v) => setForm((f) => ({ ...f, unit_size: v }))} />
          <label className="block text-sm font-medium text-slate-700">
            Komissiya (%)
            <input
              type="number"
              min={0}
              max={100}
              step="0.01"
              value={form.commission_percent}
              onChange={(e) => setForm((f) => ({ ...f, commission_percent: e.target.value }))}
              required
              className="mt-1.5 w-full rounded-xl border border-slate-200 px-3 py-2.5 outline-none focus:border-slate-400"
            />
            <span className="mt-1 block text-xs text-slate-500">
              Farq: {formatMoney(margin)} · Komissiya: {formatMoney(commission)}
            </span>
          </label>
          <Select
            label="Holat"
            value={form.status}
            onChange={(v) => setForm((f) => ({ ...f, status: v as ProductStatus }))}
            options={[
              { value: 'active', label: 'Faol' },
              { value: 'inactive', label: 'Nofaol' },
            ]}
          />
          <div className="sm:col-span-2">
            <p className="text-sm font-medium text-slate-700">Rasmlar ({form.images.length}/5)</p>
            <div className="mt-2 flex flex-wrap gap-2">
              {form.images.map((src, index) => (
                <div key={`${src}-${index}`} className="relative h-20 w-20 overflow-hidden rounded-xl border border-slate-200">
                  <ImageThumb
                    src={productImageSrc(src)}
                    images={form.images.map((item) => productImageSrc(item))}
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

function NumberField({ label, value, onChange }: { label: string; value: string; onChange: (value: string) => void }) {
  return (
    <label className="block text-sm font-medium text-slate-700">
      {label}
      <input
        type="number"
        min={0}
        step="0.01"
        value={value}
        onChange={(e) => onChange(e.target.value)}
        required
        className="mt-1.5 w-full rounded-xl border border-slate-200 px-3 py-2.5 outline-none focus:border-slate-400"
      />
    </label>
  )
}
