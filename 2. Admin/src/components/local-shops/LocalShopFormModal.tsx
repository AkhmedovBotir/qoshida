import { FormEvent, useEffect, useState } from 'react'
import { APP_COLOR, API_URL } from '../../constants/config'
import { listRegions } from '../../lib/regions'
import { ApiError } from '../../lib/api'
import type { LocalShopForm, LocalShopItem, LocalShopStatus } from '../../types/localShop'
import type { RegionItem } from '../../types/region'
import { PasswordInput } from '../ui/PasswordInput'
import { PhoneInput } from '../ui/PhoneInput'
import { ImageThumb } from '../ui/ImageLightbox'
import { Select } from '../ui/Select'
import { toast } from '../../lib/snack'

type Props = {
  item: LocalShopItem | null
  saving: boolean
  onClose: () => void
  onSubmit: (payload: LocalShopForm) => Promise<void>
}

const empty: LocalShopForm = {
  name: '',
  phone: '+998',
  image: '',
  clear_image: false,
  region_id: '',
  district_id: '',
  mfy_id: '',
  status: 'active',
  password: '',
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

export function LocalShopFormModal({ item, saving, onClose, onSubmit }: Props) {
  const isEdit = Boolean(item)
  const [form, setForm] = useState<LocalShopForm>(() =>
    item
      ? {
          name: item.name,
          phone: item.phone,
          image: '',
          clear_image: false,
          region_id: item.region_id,
          district_id: item.district_id,
          mfy_id: item.mfy_id,
          status: item.status,
          password: '',
        }
      : empty,
  )
  const [regions, setRegions] = useState<RegionItem[]>([])
  const [districts, setDistricts] = useState<RegionItem[]>([])
  const [mfys, setMfys] = useState<RegionItem[]>([])

  useEffect(() => {
    void listRegions({ type: 'region' }).then(setRegions).catch(() => setRegions([]))
  }, [])

  useEffect(() => {
    if (!form.region_id) {
      setDistricts([])
      return
    }
    void listRegions({ type: 'district', parent_id: form.region_id })
      .then(setDistricts)
      .catch(() => setDistricts([]))
  }, [form.region_id])

  useEffect(() => {
    if (!form.district_id) {
      setMfys([])
      return
    }
    void listRegions({ type: 'mfy', parent_id: form.district_id })
      .then(setMfys)
      .catch(() => setMfys([]))
  }, [form.district_id])

  async function handleSubmit(event: FormEvent) {
    event.preventDefault()
    if (!form.region_id || !form.district_id || !form.mfy_id) {
      toast.error('Hududni tanlang')
      return
    }
    try {
      await onSubmit(form)
      onClose()
    } catch (err) {
      if (!(err instanceof ApiError)) toast.error('Saqlab bo‘lmadi')
    }
  }

  const preview = form.image || (!form.clear_image && item?.image ? `${API_URL}${item.image}` : '')

  return (
    <div className="fixed inset-0 z-50 flex items-center justify-center bg-slate-950/40 px-4">
      <form onSubmit={handleSubmit} className="max-h-[90vh] w-full max-w-xl overflow-y-auto rounded-2xl bg-white p-5 shadow-xl">
        <h3 className="text-lg font-bold text-slate-900">{isEdit ? 'Do‘konni tahrirlash' : 'Yangi mahalla do‘koni'}</h3>
        <p className="mt-1 text-sm text-slate-500">Parol ixtiyoriy — bo‘sh qoldirilsa, do‘kon o‘zi SMS orqali o‘rnatadi.</p>

        <div className="mt-4 grid gap-3 sm:grid-cols-2">
          <Field label="Nomi" value={form.name} onChange={(v) => setForm((f) => ({ ...f, name: v }))} />
          <Select
            label="Holat"
            value={form.status}
            onChange={(v) => setForm((f) => ({ ...f, status: v as LocalShopStatus }))}
            options={[
              { value: 'active', label: 'Faol' },
              { value: 'inactive', label: 'Nofaol' },
            ]}
          />
          <Select
            label="Viloyat"
            value={form.region_id}
            onChange={(v) => setForm((f) => ({ ...f, region_id: v, district_id: '', mfy_id: '' }))}
            required
            searchable
            placeholder="Viloyatni tanlang"
            options={regions.map((region) => ({ value: region.id, label: region.name }))}
          />
          <Select
            label="Tuman"
            value={form.district_id}
            onChange={(v) => setForm((f) => ({ ...f, district_id: v, mfy_id: '' }))}
            required
            searchable
            placeholder="Tumanni tanlang"
            options={districts.map((district) => ({ value: district.id, label: district.name }))}
          />
          <Select
            label="MFY"
            className="sm:col-span-2"
            value={form.mfy_id}
            onChange={(v) => setForm((f) => ({ ...f, mfy_id: v }))}
            required
            searchable
            placeholder="MFY ni tanlang"
            options={mfys.map((mfy) => ({ value: mfy.id, label: mfy.name }))}
          />
          <PhoneInput value={form.phone} onChange={(v) => setForm((f) => ({ ...f, phone: v }))} />
          <PasswordInput
            label={isEdit ? 'Yangi parol (ixtiyoriy)' : 'Parol (ixtiyoriy)'}
            value={form.password}
            onChange={(e) => setForm((f) => ({ ...f, password: e.target.value }))}
            autoComplete="new-password"
          />
          <label className="block text-sm font-medium text-slate-700 sm:col-span-2">
            Rasm (ixtiyoriy)
            <input
              type="file"
              accept="image/png,image/jpeg,image/webp"
              onChange={(e) => {
                const file = e.target.files?.[0]
                if (!file) return
                void readImage(file)
                  .then((image) => setForm((f) => ({ ...f, image, clear_image: false })))
                  .catch((err: Error) => toast.error(err.message))
              }}
              className="mt-1.5 w-full text-sm"
            />
          </label>
          {preview ? (
            <div className="sm:col-span-2">
              <ImageThumb src={preview} className="h-16 w-16 rounded-xl object-cover" />
              {item?.image && !form.image ? (
                <label className="mt-2 flex items-center gap-2 text-sm font-normal text-slate-600">
                  <input
                    type="checkbox"
                    checked={form.clear_image}
                    onChange={(e) => setForm((f) => ({ ...f, clear_image: e.target.checked }))}
                  />
                  Rasmni olib tashlash
                </label>
              ) : null}
            </div>
          ) : null}
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

function Field({ label, value, onChange }: { label: string; value: string; onChange: (value: string) => void }) {
  return (
    <label className="block text-sm font-medium text-slate-700">
      {label}
      <input
        value={value}
        onChange={(e) => onChange(e.target.value)}
        required
        className="mt-1.5 w-full rounded-xl border border-slate-200 px-3 py-2.5 outline-none focus:border-slate-400"
      />
    </label>
  )
}
