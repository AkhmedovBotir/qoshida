import { Plus, Trash2, X } from 'lucide-react'
import { FormEvent, useEffect, useMemo, useState } from 'react'
import { APP_COLOR } from '../../constants/config'
import { OWN_SERVICES, serviceImageSrc } from '../../lib/services'
import type { ProviderServiceForm, ProviderServiceItem, ProviderServiceRow } from '../../types/providerService'
import { ImageThumb } from '../ui/ImageLightbox'
import { Select } from '../ui/Select'
import { toast } from '../../lib/snack'

type ProviderOption = { id: string; name: string }

const emptyRow = (): ProviderServiceRow => ({ name: '', price: '', images: [] })

type ServiceFormModalProps = {
  item?: ProviderServiceItem | null
  providers: ProviderOption[]
  saving: boolean
  onClose: () => void
  onSubmit: (payload: ProviderServiceForm) => Promise<void>
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

export function ServiceFormModal({ item, providers, saving, onClose, onSubmit }: ServiceFormModalProps) {
  const [providerId, setProviderId] = useState(item?.provider_id ?? '')
  const [items, setItems] = useState<ProviderServiceRow[]>(
    item ? [{ name: item.name, price: String(item.price), images: item.images ?? [] }] : [emptyRow()],
  )

  useEffect(() => {
    setProviderId(item?.provider_id ?? '')
    setItems(item ? [{ name: item.name, price: String(item.price), images: item.images ?? [] }] : [emptyRow()])
  }, [item])

  const options = useMemo(
    () => providers.map((p) => ({ value: p.id, label: p.name })),
    [providers],
  )

  async function handleImages(index: number, files: FileList | null) {
    if (!files?.length) return
    try {
      const next = [...items]
      const images = [...(next[index].images ?? [])]
      for (const file of Array.from(files)) {
        if (images.length >= 5) break
        images.push(await readImage(file))
      }
      next[index] = { ...next[index], images: images.slice(0, 5) }
      setItems(next)
    } catch (err) {
      toast.error(err instanceof Error ? err.message : 'Rasm qo‘shib bo‘lmadi')
    }
  }

  async function handleSubmit(event: FormEvent) {
    event.preventDefault()
    const filled = items.filter((row) => row.name.trim() || row.price.trim() || (row.images?.length ?? 0) > 0)
    if (!OWN_SERVICES && !providerId) {
      toast.error('Xizmat ko‘rsatuvchini tanlang')
      return
    }
    if (!filled.length) {
      toast.error('Kamida bitta xizmat kiriting')
      return
    }
    for (const row of filled) {
      if (row.name.trim().length < 2) {
        toast.error('Xizmat nomi kamida 2 belgi bo‘lishi kerak')
        return
      }
      if (row.price.trim() === '' || Number(row.price) < 0 || Number.isNaN(Number(row.price))) {
        toast.error('Narx 0 yoki undan katta son bo‘lishi kerak')
        return
      }
      if (!row.images?.length) {
        toast.error('Har bir xizmatga kamida bitta rasm yuklang')
        return
      }
    }
    try {
      await onSubmit({ provider_id: providerId, items: filled })
    } catch (err) {
      toast.error(err instanceof Error ? err.message : 'Saqlab bo‘lmadi')
    }
  }

  return (
    <div className="fixed inset-0 z-50 flex items-center justify-center bg-slate-950/40 px-4">
      <form onSubmit={handleSubmit} className="max-h-[92vh] w-full max-w-2xl overflow-y-auto rounded-2xl bg-white p-5 shadow-xl">
        <div className="flex items-start justify-between gap-3">
          <div>
            <h3 className="text-lg font-bold text-slate-900">{item ? 'Xizmatni tahrirlash' : 'Xizmat qo‘shish'}</h3>
            <p className="mt-1 text-sm text-slate-500">Har qator: nomi, narxi va rasmi</p>
          </div>
          <button type="button" onClick={onClose} className="rounded-lg p-1 text-slate-400 hover:bg-slate-100" aria-label="Yopish">
            <X size={18} />
          </button>
        </div>

        <div className="mt-4 space-y-4">
          {!OWN_SERVICES ? (
            <Select
              label="Xizmat ko‘rsatuvchi"
              value={providerId}
              onChange={setProviderId}
              options={options}
              placeholder="Tanlang"
              searchable
              required
              disabled={Boolean(item)}
            />
          ) : null}

          {items.map((row, index) => (
            <div key={index} className="rounded-2xl border border-slate-200 p-3">
              <div className="grid gap-2 sm:grid-cols-[1fr_140px_auto] sm:items-end">
                <label className="block text-sm font-medium text-slate-700">
                  Xizmat nomi
                  <input
                    value={row.name}
                    onChange={(e) => {
                      const next = [...items]
                      next[index] = { ...next[index], name: e.target.value }
                      setItems(next)
                    }}
                    placeholder="Masalan, Soch olish"
                    className="mt-1.5 w-full rounded-xl border border-slate-200 px-3 py-2.5 text-sm outline-none focus:border-slate-400"
                  />
                </label>
                <label className="block text-sm font-medium text-slate-700">
                  Narxi
                  <input
                    value={row.price}
                    onChange={(e) => {
                      const next = [...items]
                      next[index] = { ...next[index], price: e.target.value.replace(/[^\d.]/g, '') }
                      setItems(next)
                    }}
                    inputMode="decimal"
                    placeholder="0"
                    className="mt-1.5 w-full rounded-xl border border-slate-200 px-3 py-2.5 text-sm outline-none focus:border-slate-400"
                  />
                </label>
                {!item && items.length > 1 ? (
                  <button
                    type="button"
                    onClick={() => setItems(items.filter((_, i) => i !== index))}
                    className="mb-0.5 rounded-xl border border-red-100 p-2.5 text-red-600 hover:bg-red-50"
                    aria-label="Qatorni o‘chirish"
                  >
                    <Trash2 size={16} />
                  </button>
                ) : (
                  <span className="hidden sm:block" />
                )}
              </div>
              <div className="mt-3">
                <p className="text-sm font-medium text-slate-700">Rasmlar ({(row.images ?? []).length}/5)</p>
                <div className="mt-2 flex flex-wrap gap-2">
                  {(row.images ?? []).map((src, imgIndex) => (
                    <div key={`${src}-${imgIndex}`} className="relative h-20 w-20 overflow-hidden rounded-xl border border-slate-200">
                      <ImageThumb
                        src={serviceImageSrc(src)}
                        images={(row.images ?? []).map((itemSrc) => serviceImageSrc(itemSrc))}
                        className="h-20 w-20 object-cover"
                      />
                      <button
                        type="button"
                        onClick={() => {
                          const next = [...items]
                          next[index] = {
                            ...next[index],
                            images: (next[index].images ?? []).filter((_, i) => i !== imgIndex),
                          }
                          setItems(next)
                        }}
                        className="absolute top-1 right-1 z-10 rounded bg-white/90 px-1 text-xs text-red-600"
                      >
                        ×
                      </button>
                    </div>
                  ))}
                  {(row.images ?? []).length < 5 ? (
                    <label className="flex h-20 w-20 cursor-pointer items-center justify-center rounded-xl border border-dashed border-slate-300 text-xs text-slate-500">
                      + Rasm
                      <input
                        type="file"
                        accept="image/png,image/jpeg,image/webp"
                        hidden
                        multiple
                        onChange={(e) => void handleImages(index, e.target.files)}
                      />
                    </label>
                  ) : null}
                </div>
              </div>
            </div>
          ))}

          {!item ? (
            <button
              type="button"
              onClick={() => setItems((rows) => [...rows, emptyRow()])}
              className="inline-flex items-center gap-2 text-sm font-medium text-slate-600 hover:text-slate-900"
            >
              <Plus size={16} />
              Yana qator
            </button>
          ) : null}
        </div>

        <div className="mt-5 flex justify-end gap-2">
          <button type="button" onClick={onClose} disabled={saving} className="rounded-xl border border-slate-200 px-4 py-2 text-sm">
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
