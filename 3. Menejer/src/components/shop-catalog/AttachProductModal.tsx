import { FormEvent, useEffect, useState } from 'react'
import { APP_COLOR } from '../../constants/config'
import { ApiError } from '../../lib/api'
import { listLocalShops, type LocalShopItem } from '../../lib/localShops'
import { catalogImageSrc, listShopTemplates } from '../../lib/shopCatalog'
import type { ShopProductForm, ShopProductItem, ShopTemplateItem } from '../../types/shopCatalog'
import { Select } from '../ui/Select'
import { toast } from '../../lib/snack'

type Props = {
  item: ShopProductItem | null
  saving: boolean
  lockShop?: boolean
  shopId?: string
  shopsPath?: string
  templatesPath?: string
  onClose: () => void
  onSubmit: (payload: ShopProductForm) => Promise<void>
}

export function AttachProductModal({ item, saving, lockShop, shopId, onClose, onSubmit }: Props) {
  const isEdit = Boolean(item)
  const [form, setForm] = useState<ShopProductForm>(() =>
    item
      ? {
          shop_id: item.shop_id,
          template_id: item.template_id,
          quantity: String(item.quantity),
          sale_price: String(item.sale_price),
          cost_price: String(item.cost_price),
        }
      : {
          shop_id: shopId ?? '',
          template_id: '',
          quantity: '0',
          sale_price: '',
          cost_price: '',
        },
  )
  const [shops, setShops] = useState<LocalShopItem[]>([])
  const [templates, setTemplates] = useState<ShopTemplateItem[]>([])

  useEffect(() => {
    if (!lockShop && !isEdit) {
      void listLocalShops({ page: 1, limit: 100 })
        .then((data) => setShops(data.items ?? []))
        .catch(() => setShops([]))
    }
    if (!isEdit) {
      void listShopTemplates({ page: 1, limit: 100 })
        .then((data) => setTemplates(data.items ?? []))
        .catch(() => setTemplates([]))
    }
  }, [lockShop, isEdit])

  async function handleSubmit(event: FormEvent) {
    event.preventDefault()
    if (!isEdit && !form.shop_id && !lockShop) {
      toast.error('Do‘kon tanlang')
      return
    }
    if (!isEdit && !form.template_id) {
      toast.error('Shablon tanlang')
      return
    }
    try {
      await onSubmit({ ...form, shop_id: form.shop_id || shopId || '' })
      onClose()
    } catch (err) {
      if (!(err instanceof ApiError)) toast.error('Saqlab bo‘lmadi')
    }
  }

  return (
    <div className="fixed inset-0 z-50 flex items-center justify-center bg-slate-950/40 px-4">
      <form onSubmit={handleSubmit} className="w-full max-w-lg rounded-2xl bg-white p-5 shadow-xl">
        <h3 className="text-lg font-bold text-slate-900">{isEdit ? 'Narxlarni tahrirlash' : 'Shablondan mahsulot biriktirish'}</h3>
        <p className="mt-1 text-sm text-slate-500">
          {isEdit ? 'Miqdor faqat kirim orqali o‘zgaradi.' : 'Miqdor, sotuv va asl narxini kiriting.'}
        </p>

        <div className="mt-4 grid gap-3">
          {!lockShop && !isEdit ? (
            <Select
              label="Do‘kon"
              value={form.shop_id}
              onChange={(v) => setForm((f) => ({ ...f, shop_id: v }))}
              required
              searchable
              placeholder="Do‘konni tanlang"
              options={shops.map((shop) => ({ value: shop.id, label: shop.name }))}
            />
          ) : null}
          {!isEdit ? (
            <Select
              label="Shablon"
              value={form.template_id}
              onChange={(v) => setForm((f) => ({ ...f, template_id: v }))}
              required
              searchable
              placeholder="Shablonni tanlang"
              options={templates.map((tpl) => ({
                value: tpl.id,
                label: tpl.name,
                hint: `${tpl.unit} × ${tpl.unit_size}`,
                image: tpl.images?.[0] ? catalogImageSrc(tpl.images[0]) : undefined,
              }))}
            />
          ) : (
            <p className="rounded-xl bg-slate-50 px-3 py-2 text-sm text-slate-700">{item?.name}</p>
          )}
          {!isEdit ? (
            <label className="block text-sm font-medium text-slate-700">
              Nechta bor
              <input
                type="number"
                min={0}
                step="0.001"
                value={form.quantity}
                onChange={(e) => setForm((f) => ({ ...f, quantity: e.target.value }))}
                required
                className="mt-1.5 w-full rounded-xl border border-slate-200 px-3 py-2.5 outline-none focus:border-slate-400"
              />
            </label>
          ) : null}
          <label className="block text-sm font-medium text-slate-700">
            Sotuv narxi
            <input
              type="number"
              min={0}
              step="0.01"
              value={form.sale_price}
              onChange={(e) => setForm((f) => ({ ...f, sale_price: e.target.value }))}
              required
              className="mt-1.5 w-full rounded-xl border border-slate-200 px-3 py-2.5 outline-none focus:border-slate-400"
            />
          </label>
          <label className="block text-sm font-medium text-slate-700">
            Asl narxi
            <input
              type="number"
              min={0}
              step="0.01"
              value={form.cost_price}
              onChange={(e) => setForm((f) => ({ ...f, cost_price: e.target.value }))}
              required
              className="mt-1.5 w-full rounded-xl border border-slate-200 px-3 py-2.5 outline-none focus:border-slate-400"
            />
          </label>
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
