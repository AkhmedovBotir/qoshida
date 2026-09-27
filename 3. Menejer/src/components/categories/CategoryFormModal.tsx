import { FormEvent, useState } from 'react'
import { APP_COLOR } from '../../constants/config'
import { ApiError } from '../../lib/api'
import { auditFields } from '../../lib/audit'
import { categoryImageSrc, type CategoryWrite } from '../../lib/categories'
import { getCategoryId, type CategoryItem, type CategoryStatus } from '../../types/category'
import { ImageThumb } from '../ui/ImageLightbox'
import { Select } from '../ui/Select'
import { toast } from '../../lib/snack'

type FormKind = 'root' | 'child'

const TITLES = {
  create: { root: 'Yangi asosiy kategoriya', child: 'Yangi ichki kategoriya' },
  edit: { root: 'Asosiy kategoriyani tahrirlash', child: 'Ichki kategoriyani tahrirlash' },
}

type CategoryFormModalProps = {
  kind: FormKind
  item: CategoryItem | null
  roots: CategoryItem[]
  saving: boolean
  onClose: () => void
  onSubmit: (payload: CategoryWrite) => Promise<void>
}

function readImage(file: File): Promise<string> {
  return new Promise((resolve, reject) => {
    if (!['image/jpeg', 'image/jpg', 'image/png', 'image/webp'].includes(file.type)) {
      reject(new Error('Rasm faqat PNG, JPG yoki WEBP bo‘lsin'))
      return
    }
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

export function CategoryFormModal({ kind, item, roots, saving, onClose, onSubmit }: CategoryFormModalProps) {
  const isEdit = Boolean(item)
  const [name, setName] = useState(item?.name ?? '')
  const [slug, setSlug] = useState(item?.slug ?? '')
  const [parentId, setParentId] = useState(item?.parent_id ?? '')
  const [status, setStatus] = useState<CategoryStatus>(item?.status ?? 'active')
  const [censored, setCensored] = useState(item?.censored ?? false)
  const [image, setImage] = useState('')
  const [clearImage, setClearImage] = useState(false)
  const [errors, setErrors] = useState<Record<string, string>>({})

  const preview = image || (!clearImage && item?.image_url ? categoryImageSrc(item.image_url) : '')

  async function handleSubmit(event: FormEvent) {
    event.preventDefault()
    const next: Record<string, string> = {}
    if (!name.trim()) next.name = 'Nomi majburiy'
    if (kind === 'child' && !parentId) next.parent_id = 'Asosiy kategoriyani tanlang'
    setErrors(next)
    if (Object.keys(next).length) return

    try {
      await onSubmit({
        name: name.trim(),
        slug: slug.trim(),
        parent_id: kind === 'root' ? null : parentId,
        censored,
        status,
        image: image || undefined,
        clear_image: clearImage,
      })
      onClose()
    } catch (err) {
      if (!(err instanceof ApiError)) toast.error('Saqlab bo‘lmadi')
    }
  }

  return (
    <div className="fixed inset-0 z-50 flex items-center justify-center bg-slate-950/40 px-4">
      <form onSubmit={handleSubmit} className="max-h-[92vh] w-full max-w-lg overflow-y-auto rounded-2xl bg-white p-5 shadow-xl">
        <h3 className="text-lg font-bold text-slate-900">{isEdit ? TITLES.edit[kind] : TITLES.create[kind]}</h3>
        {isEdit && auditFields(item?.audit).length ? (
          <div className="mt-3 space-y-1.5 rounded-xl bg-slate-50 p-3">
            {auditFields(item?.audit).map((field) => (
              <p key={field.label} className="whitespace-pre-line text-xs text-slate-600">
                <span className="font-semibold text-slate-500">{field.label}: </span>
                {field.value}
              </p>
            ))}
          </div>
        ) : null}
        <div className="mt-4 space-y-3">
          <div>
            <p className="text-sm font-medium text-slate-700">Rasm (ixtiyoriy)</p>
            <input
              type="file"
              accept="image/png,image/jpeg,image/webp"
              className="mt-1.5 w-full text-sm"
              onChange={(event) => {
                const file = event.target.files?.[0]
                event.target.value = ''
                if (!file) return
                void readImage(file)
                  .then((dataUrl) => {
                    setImage(dataUrl)
                    setClearImage(false)
                  })
                  .catch((err: unknown) => toast.error(err instanceof Error ? err.message : 'Rasm yuklanmadi'))
              }}
            />
            {preview ? (
              <div className="mt-2">
                <ImageThumb src={preview} className="h-20 w-20 rounded-xl object-cover ring-1 ring-slate-200" />
              </div>
            ) : (
              <p className="mt-1.5 text-xs text-slate-500">PNG, JPG yoki WEBP. Maksimal 10 MB.</p>
            )}
            {item?.image_url && !image ? (
              <label className="mt-2 flex items-center gap-2 text-sm font-medium text-slate-700">
                <input type="checkbox" checked={clearImage} onChange={(e) => setClearImage(e.target.checked)} />
                Rasmni olib tashlash
              </label>
            ) : null}
          </div>

          <label className="block text-sm font-medium text-slate-700">
            Nomi
            <input
              value={name}
              onChange={(e) => setName(e.target.value)}
              placeholder="Nomi kiriting"
              className="mt-1.5 w-full rounded-xl border border-slate-200 px-3 py-2.5 outline-none focus:border-slate-400"
            />
            {errors.name ? <p className="mt-1 text-xs text-red-600">{errors.name}</p> : null}
          </label>

          <label className="block text-sm font-medium text-slate-700">
            Slug
            <input
              value={slug}
              onChange={(e) => setSlug(e.target.value)}
              placeholder="masalan: turizm (bo‘sh qolsa avtomatik)"
              className="mt-1.5 w-full rounded-xl border border-slate-200 px-3 py-2.5 outline-none focus:border-slate-400"
            />
          </label>

          {kind === 'child' ? (
            <div>
              <Select
                label="Asosiy kategoriya"
                value={parentId}
                onChange={setParentId}
                required
                searchable
                placeholder="Tanlang"
                options={roots.map((root) => ({ value: getCategoryId(root), label: root.name }))}
              />
              {errors.parent_id ? <p className="mt-1 text-xs text-red-600">{errors.parent_id}</p> : null}
            </div>
          ) : null}

          <Select
            label="Holat"
            value={status}
            onChange={(v) => setStatus(v as CategoryStatus)}
            options={[
              { value: 'active', label: 'Faol' },
              { value: 'inactive', label: 'Nofaol' },
            ]}
          />

          <label className="flex items-center gap-2 text-sm font-medium text-slate-700">
            <input type="checkbox" checked={censored} onChange={(e) => setCensored(e.target.checked)} />
            Cheklangan (censored)
          </label>
        </div>

        <div className="mt-5 flex justify-end gap-2">
          <button
            type="button"
            onClick={onClose}
            disabled={saving}
            className="rounded-xl border border-slate-200 px-4 py-2 text-sm font-medium disabled:opacity-60"
          >
            Bekor qilish
          </button>
          <button
            type="submit"
            disabled={saving}
            className="rounded-xl px-4 py-2 text-sm font-semibold text-white disabled:opacity-60"
            style={{ backgroundColor: APP_COLOR }}
          >
            {saving ? 'Saqlanmoqda...' : isEdit ? 'Yangilash' : 'Qo‘shish'}
          </button>
        </div>
      </form>
    </div>
  )
}
