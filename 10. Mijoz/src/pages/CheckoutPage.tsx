import { FormEvent, useState } from 'react'
import { Link, useNavigate } from 'react-router-dom'
import { GeoFields } from '../components/geo/GeoFields'
import { EmptyState } from '../components/ui/EmptyState'
import { useAuth } from '../context/AuthContext'
import { useCart } from '../context/CartContext'
import { checkout } from '../lib/market'
import { ApiError } from '../lib/api'
import { formatMoney } from '../lib/media'
import { fieldClass } from '../constants/theme'
import { toast } from '../lib/snack'

export function CheckoutPage() {
  const { user } = useAuth()
  const { cart, refresh } = useCart()
  const navigate = useNavigate()
  const [address, setAddress] = useState(
    user?.address || [user?.region_name, user?.district_name, user?.mfy_name].filter(Boolean).join(', '),
  )
  const [note, setNote] = useState('')
  const [geo, setGeo] = useState({
    region_id: user?.region_id ?? '',
    district_id: user?.district_id ?? '',
    mfy_id: user?.mfy_id ?? '',
  })
  const [busy, setBusy] = useState(false)

  async function onSubmit(event: FormEvent) {
    event.preventDefault()
    setBusy(true)
    try {
      const order = await checkout({
        address,
        note,
        region_id: geo.region_id,
        district_id: geo.district_id,
        mfy_id: geo.mfy_id,
      })
      await refresh()
      navigate(`/orders/${order.id}`, { replace: true })
    } catch (err) {
      if (!(err instanceof ApiError)) toast.error('Buyurtma yuborilmadi')
    } finally {
      setBusy(false)
    }
  }

  if (cart.items.length === 0) {
    return (
      <EmptyState
        title="Savat bo‘sh"
        text="Avval katalogdan mahsulot qo‘shing, keyin buyurtma bering."
        actionTo="/catalog"
        actionLabel="Katalog"
      />
    )
  }

  return (
    <form onSubmit={onSubmit} className="grid items-start gap-5 lg:grid-cols-[minmax(0,1fr)_320px]">
      <div className="space-y-4">
        <div>
          <p className="text-xs font-bold tracking-wide text-[#2E5D90] uppercase">2-qadam</p>
          <h2 className="text-xl font-extrabold text-slate-900">Yetkazish manzili</h2>
          <p className="mt-1 text-sm text-slate-500">Buyurtma ushbu manzilga yetkaziladi. To‘lov kuryerga naqd.</p>
        </div>
        <div className="market-card space-y-4 p-4 sm:p-5">
          <GeoFields value={geo} onChange={setGeo} />
          <label className="block text-sm font-medium text-slate-700">
            Aniq manzil
            <textarea
              required
              minLength={8}
              maxLength={300}
              value={address}
              onChange={(e) => setAddress(e.target.value)}
              rows={3}
              className={fieldClass}
              placeholder="Ko‘cha, uy, mo‘ljal"
            />
          </label>
          <label className="block text-sm font-medium text-slate-700">
            Izoh
            <textarea
              maxLength={400}
              value={note}
              onChange={(e) => setNote(e.target.value)}
              rows={2}
              className={fieldClass}
              placeholder="Ixtiyoriy: vaqt, eshik kodi..."
            />
          </label>
        </div>
      </div>

      <aside className="market-card space-y-3 p-5 lg:sticky lg:top-24">
        <h3 className="font-bold text-slate-900">Buyurtma</h3>
        <ul className="space-y-2 text-sm">
          {cart.items.slice(0, 4).map((line) => (
            <li key={line.id} className="flex justify-between gap-3 text-slate-600">
              <span className="line-clamp-1">{line.item?.name ?? 'Mahsulot'}</span>
              <span className="shrink-0 font-medium">{formatMoney(line.line_total)}</span>
            </li>
          ))}
          {cart.items.length > 4 ? <li className="text-xs text-slate-400">va yana {cart.items.length - 4} ta</li> : null}
        </ul>
        <div className="flex items-center justify-between border-t border-[#E6EDF4] pt-3">
          <span className="text-slate-500">Jami</span>
          <span className="price text-xl">{formatMoney(cart.total)}</span>
        </div>
        <p className="rounded-xl bg-[#E8F0F8] px-3 py-2 text-xs leading-5 text-[#244A73]">
          To‘lov: yetkazib berganda naqd (COD)
        </p>
        <button
          type="submit"
          disabled={busy}
          className="w-full rounded-2xl bg-[#2E5D90] py-3 text-sm font-semibold text-white disabled:opacity-60"
        >
          {busy ? 'Yuborilmoqda...' : 'Buyurtma berish'}
        </button>
        <Link to="/cart" className="block text-center text-sm font-semibold text-slate-500">
          Savatga qaytish
        </Link>
      </aside>
    </form>
  )
}
