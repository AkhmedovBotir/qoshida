import { FormEvent, useEffect, useState } from 'react'
import { APP_COLOR } from '../../constants/config'
import { listLocalShops } from '../../lib/localShops'
import { listRegions } from '../../lib/regions'
import { ApiError } from '../../lib/api'
import type { LocalShopItem } from '../../types/localShop'
import type { RegionItem } from '../../types/region'
import type { DeliveryForm, DeliveryItem, DeliveryStatus } from '../../types/delivery'
import { PasswordInput } from '../ui/PasswordInput'
import { PhoneInput } from '../ui/PhoneInput'
import { Select } from '../ui/Select'
import { toast } from '../../lib/snack'

type DeliveryFormModalProps = {
  item: DeliveryItem | null
  saving: boolean
  onClose: () => void
  onSubmit: (payload: DeliveryForm) => Promise<void>
}

const empty: DeliveryForm = {
  shop_id: '',
  region_id: '',
  district_id: '',
  mfy_id: '',
  first_name: '',
  last_name: '',
  phone: '+998',
  status: 'active',
  password: '',
}

export function DeliveryFormModal({ item, saving, onClose, onSubmit }: DeliveryFormModalProps) {
  const isEdit = Boolean(item)
  const [form, setForm] = useState<DeliveryForm>(() =>
    item
      ? {
          shop_id: item.shop_id,
          region_id: item.region_id ?? '',
          district_id: item.district_id ?? '',
          mfy_id: item.mfy_id ?? '',
          first_name: item.first_name,
          last_name: item.last_name,
          phone: item.phone,
          status: item.status,
          password: '',
        }
      : empty,
  )
  const [regions, setRegions] = useState<RegionItem[]>([])
  const [districts, setDistricts] = useState<RegionItem[]>([])
  const [mfys, setMfys] = useState<RegionItem[]>([])
  const [shops, setShops] = useState<LocalShopItem[]>([])

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

  useEffect(() => {
    if (!form.mfy_id) {
      setShops([])
      return
    }
    void listLocalShops({ page: 1, limit: 100, mfy_id: form.mfy_id })
      .then((data) => setShops(data.items ?? []))
      .catch(() => setShops([]))
  }, [form.mfy_id])

  async function handleSubmit(event: FormEvent) {
    event.preventDefault()
    if (!form.region_id || !form.district_id || !form.mfy_id || !form.shop_id) {
      toast.error('Viloyat, tuman, MFY va do‘kon tanlang')
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
      <form onSubmit={handleSubmit} className="max-h-[90vh] w-full max-w-xl overflow-y-auto rounded-2xl bg-white p-5 shadow-xl">
        <h3 className="text-lg font-bold text-slate-900">{isEdit ? 'Yetkazuvchini tahrirlash' : 'Yangi yetkazuvchi'}</h3>
        <p className="mt-1 text-sm text-slate-500">Avval hudud va do‘konni tanlang. Parol ixtiyoriy — bo‘sh qoldirilsa, yetkazuvchi o‘zi SMS orqali o‘rnatadi.</p>

        <div className="mt-4 grid gap-3 sm:grid-cols-2">
          <Select
            label="Holat"
            value={form.status}
            onChange={(v) => setForm((f) => ({ ...f, status: v as DeliveryStatus }))}
            options={[
              { value: 'active', label: 'Faol' },
              { value: 'inactive', label: 'Nofaol' },
            ]}
          />
          <Select
            label="Viloyat"
            value={form.region_id}
            onChange={(v) => setForm((f) => ({ ...f, region_id: v, district_id: '', mfy_id: '', shop_id: '' }))}
            required
            searchable
            placeholder="Viloyatni tanlang"
            options={regions.map((region) => ({ value: region.id, label: region.name }))}
          />
          <Select
            label="Tuman"
            value={form.district_id}
            onChange={(v) => setForm((f) => ({ ...f, district_id: v, mfy_id: '', shop_id: '' }))}
            required
            searchable
            placeholder="Tumanni tanlang"
            options={districts.map((district) => ({ value: district.id, label: district.name }))}
          />
          <Select
            label="MFY"
            value={form.mfy_id}
            onChange={(v) => setForm((f) => ({ ...f, mfy_id: v, shop_id: '' }))}
            required
            searchable
            placeholder="MFY ni tanlang"
            options={mfys.map((mfy) => ({ value: mfy.id, label: mfy.name }))}
          />
          <Select
            label="Do‘kon"
            className="sm:col-span-2"
            value={form.shop_id}
            onChange={(v) => setForm((f) => ({ ...f, shop_id: v }))}
            required
            searchable
            placeholder="Do‘konni tanlang"
            options={shops.map((shop) => ({ value: shop.id, label: shop.name }))}
          />
          <Field label="Ism" value={form.first_name} onChange={(v) => setForm((f) => ({ ...f, first_name: v }))} />
          <Field label="Familiya" value={form.last_name} onChange={(v) => setForm((f) => ({ ...f, last_name: v }))} />
          <PhoneInput value={form.phone} onChange={(v) => setForm((f) => ({ ...f, phone: v }))} />
          <PasswordInput
            label={isEdit ? 'Yangi parol (ixtiyoriy)' : 'Parol (ixtiyoriy)'}
            value={form.password}
            onChange={(e) => setForm((f) => ({ ...f, password: e.target.value }))}
            autoComplete="new-password"
          />
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
