import { FormEvent, useEffect, useState } from 'react'
import { APP_COLOR } from '../../constants/config'
import { ApiError } from '../../lib/api'
import { catalogImageSrc, listShopProducts } from '../../lib/shopCatalog'
import type { ShopProductItem } from '../../types/shopCatalog'
import { Select } from '../ui/Select'
import { toast } from '../../lib/snack'

type Props = {
  saving: boolean
  onClose: () => void
  onSubmit: (payload: { shop_product_id: string; quantity: string }) => Promise<void>
}

export function IncomingFormModal({ saving, onClose, onSubmit }: Props) {
  const [shopProductId, setShopProductId] = useState('')
  const [quantity, setQuantity] = useState('')
  const [products, setProducts] = useState<ShopProductItem[]>([])

  useEffect(() => {
    void listShopProducts({ page: 1, limit: 100 })
      .then((data) => setProducts(data.items ?? []))
      .catch(() => setProducts([]))
  }, [])

  async function handleSubmit(event: FormEvent) {
    event.preventDefault()
    if (!shopProductId) {
      toast.error('Mahsulot tanlang')
      return
    }
    try {
      await onSubmit({ shop_product_id: shopProductId, quantity })
      onClose()
    } catch (err) {
      if (!(err instanceof ApiError)) toast.error('Saqlab bo‘lmadi')
    }
  }

  return (
    <div className="fixed inset-0 z-50 flex items-center justify-center bg-slate-950/40 px-4">
      <form onSubmit={handleSubmit} className="w-full max-w-lg rounded-2xl bg-white p-5 shadow-xl">
        <h3 className="text-lg font-bold text-slate-900">Kirim qilish</h3>
        <p className="mt-1 text-sm text-slate-500">Tanlangan mahsulot omboriga miqdor qo‘shiladi.</p>

        <div className="mt-4 grid gap-3">
          <Select
            label="Mahsulot"
            value={shopProductId}
            onChange={setShopProductId}
            required
            searchable
            placeholder="Mahsulotni tanlang"
            options={products.map((item) => ({
              value: item.id,
              label: item.name,
              hint: `${item.shop_name ? `${item.shop_name} · ` : ''}${item.quantity} ${item.unit}`,
              image: item.images?.[0] ? catalogImageSrc(item.images[0]) : undefined,
            }))}
          />
          <label className="block text-sm font-medium text-slate-700">
            Kirim miqdori
            <input
              type="number"
              min={0.001}
              step="0.001"
              value={quantity}
              onChange={(e) => setQuantity(e.target.value)}
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
            {saving ? 'Saqlanmoqda...' : 'Kirim qilish'}
          </button>
        </div>
      </form>
    </div>
  )
}
