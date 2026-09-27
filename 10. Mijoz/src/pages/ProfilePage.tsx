import {
  Camera,
  ChevronRight,
  ClipboardList,
  LogOut,
  MapPin,
  Phone,
  ShoppingBag,
  UserRound,
} from 'lucide-react'
import { FormEvent, useEffect, useRef, useState } from 'react'
import { Link } from 'react-router-dom'
import { ProfileEditModal } from '../components/auth/ProfileEditModal'
import { type CustomerFormValue } from '../components/auth/CustomerFields'
import { useAuth } from '../context/AuthContext'
import { useCart } from '../context/CartContext'
import { ApiError } from '../lib/api'
import { toast } from '../lib/snack'
import { listOrders, updateProfile } from '../lib/market'
import { mediaSrc } from '../lib/media'
import type { CustomerUser } from '../types/customer'

function initials(first?: string, last?: string, fallback?: string) {
  const a = (first || '').trim().charAt(0)
  const b = (last || '').trim().charAt(0)
  const fromName = `${a}${b}`.toUpperCase()
  if (fromName.trim()) return fromName
  return (fallback || 'Q').slice(0, 2).toUpperCase()
}

function formFromUser(user?: CustomerUser | null): CustomerFormValue {
  return {
    first_name: user?.first_name || user?.name?.split(' ')[0] || '',
    last_name: user?.last_name || user?.name?.split(' ').slice(1).join(' ') || '',
    birth_date: user?.birth_date ?? '',
    region_id: user?.region_id ?? '',
    district_id: user?.district_id ?? '',
    mfy_id: user?.mfy_id ?? '',
  }
}

async function fileToDataUrl(file: File) {
  if (!['image/jpeg', 'image/jpg', 'image/png', 'image/webp'].includes(file.type)) {
    throw new Error('Rasm faqat PNG, JPG yoki WEBP bo‘lsin')
  }
  if (file.size > 10 * 1024 * 1024) {
    throw new Error('Rasm 10 MB dan oshmasin')
  }
  return await new Promise<string>((resolve, reject) => {
    const reader = new FileReader()
    reader.onload = () => resolve(String(reader.result || ''))
    reader.onerror = () => reject(new Error('Rasmni o‘qib bo‘lmadi'))
    reader.readAsDataURL(file)
  })
}

export function ProfilePage() {
  const { user, setUser, logout } = useAuth()
  const { cart } = useCart()
  const avatarInputRef = useRef<HTMLInputElement>(null)
  const [editing, setEditing] = useState(false)
  const [form, setForm] = useState<CustomerFormValue>(formFromUser(user))
  const [avatarDraft, setAvatarDraft] = useState(user?.avatar ?? '')
  const [clearAvatar, setClearAvatar] = useState(false)
  const [busy, setBusy] = useState(false)
  const [orderCount, setOrderCount] = useState(0)

  useEffect(() => {
    listOrders()
      .then((data) => setOrderCount(data.total ?? data.items?.length ?? 0))
      .catch(() => setOrderCount(0))
  }, [])

  const first = user?.first_name || user?.name?.split(' ')[0] || 'Mijoz'
  const last = user?.last_name || user?.name?.split(' ').slice(1).join(' ') || ''
  const fullName = [first, last].filter(Boolean).join(' ')
  const place = [user?.region_name, user?.district_name, user?.mfy_name].filter(Boolean).join(' · ')
  const avatarSrc = user?.avatar ? mediaSrc(user.avatar) : ''

  function openEdit() {
    setForm(formFromUser(user))
    setAvatarDraft(user?.avatar ?? '')
    setClearAvatar(false)
    setEditing(true)
  }

  async function save(payload: CustomerFormValue, extra?: { avatar?: string; clear_avatar?: boolean }) {
    const next = await updateProfile({
      ...payload,
      avatar: extra?.avatar,
      clear_avatar: extra?.clear_avatar,
    })
    setUser(next)
  }

  async function onSubmit(event: FormEvent) {
    event.preventDefault()
    setBusy(true)
    try {
      await save(form, {
        avatar: clearAvatar ? undefined : avatarDraft.startsWith('data:') ? avatarDraft : undefined,
        clear_avatar: clearAvatar,
      })
      setEditing(false)
    } catch (err) {
      if (!(err instanceof ApiError)) toast.error(err instanceof Error ? err.message : 'Saqlanmadi')
    } finally {
      setBusy(false)
    }
  }

  async function onHeaderAvatar(file: File) {
    setBusy(true)
    try {
      const dataUrl = await fileToDataUrl(file)
      await save(formFromUser(user), { avatar: dataUrl })
    } catch (err) {
      if (!(err instanceof ApiError)) toast.error(err instanceof Error ? err.message : 'Rasm yuklanmadi')
    } finally {
      setBusy(false)
    }
  }

  const menu = [
    { to: '/orders', icon: ClipboardList, title: 'Buyurtmalarim', hint: orderCount ? `${orderCount} ta buyurtma` : 'Hozircha yo‘q', tone: 'bg-[#E8F0F8] text-[#2E5D90]' },
    { to: '/cart', icon: ShoppingBag, title: 'Savat', hint: cart.count ? `${cart.count} ta mahsulot` : 'Bo‘sh', tone: 'bg-[#FFF4E5] text-[#C2410C]' },
  ]

  return (
    <div className="mx-auto max-w-3xl space-y-4">
      <section className="overflow-hidden rounded-[1.75rem] bg-[#1B3A5C] text-white shadow-[0_18px_40px_rgba(27,58,92,0.28)]">
        <div className="relative px-5 py-6 sm:px-7 sm:py-8">
          <div className="pointer-events-none absolute -top-10 right-6 h-36 w-36 rounded-full bg-[#F4B942]/20 blur-2xl" />
          <div className="relative flex items-center gap-4">
            <div className="relative shrink-0">
              <div
                className="rounded-full p-[5px] sm:p-[6px] shadow-[0_10px_24px_rgba(0,0,0,0.28)]"
                style={{
                  background: 'linear-gradient(145deg, #F4B942 0%, #FFE08A 26%, #7DD3FC 58%, #2E5D90 100%)',
                }}
              >
                <div className="flex h-16 w-16 items-center justify-center overflow-hidden rounded-full bg-[#F4B942] text-xl font-extrabold text-[#3A2A08] sm:h-[4.75rem] sm:w-[4.75rem] sm:text-2xl">
                  {avatarSrc ? <img src={avatarSrc} alt="" className="h-full w-full object-cover" /> : initials(first, last, user?.name)}
                </div>
              </div>
              <button
                type="button"
                disabled={busy}
                onClick={() => avatarInputRef.current?.click()}
                className="absolute right-0 bottom-0 flex h-8 w-8 items-center justify-center rounded-full bg-white text-[#1B3A5C] shadow-md ring-2 ring-[#1B3A5C]"
                aria-label="Avatar yuklash"
              >
                <Camera size={14} />
              </button>
              <input
                ref={avatarInputRef}
                type="file"
                accept="image/png,image/jpeg,image/webp"
                className="hidden"
                onChange={(event) => {
                  const file = event.target.files?.[0]
                  event.target.value = ''
                  if (file) void onHeaderAvatar(file)
                }}
              />
            </div>
            <div className="min-w-0">
              <p className="text-xs font-semibold tracking-wide text-white/60 uppercase">Mening profilim</p>
              <h1 className="truncate text-2xl font-extrabold tracking-tight sm:text-3xl">{fullName}</h1>
              <p className="mt-1 flex items-center gap-1.5 text-sm text-white/80">
                <Phone size={14} />
                {user?.phone}
              </p>
              {place ? (
                <p className="mt-1 flex items-center gap-1.5 truncate text-xs text-white/65">
                  <MapPin size={13} />
                  {place}
                </p>
              ) : null}
            </div>
          </div>
        </div>
      </section>

      <section className="overflow-hidden rounded-[1.5rem] bg-white shadow-[0_10px_28px_rgba(18,32,51,0.05)] ring-1 ring-[#E6EDF4]">
        {menu.map((item) => {
          const Icon = item.icon
          return (
            <Link
              key={item.to}
              to={item.to}
              className="flex items-center gap-3 border-b border-[#E6EDF4] px-4 py-3.5 last:border-b-0 sm:px-5"
            >
              <span className={`flex h-11 w-11 items-center justify-center rounded-2xl ${item.tone}`}>
                <Icon size={18} />
              </span>
              <span className="min-w-0 flex-1">
                <span className="block font-semibold text-slate-900">{item.title}</span>
                <span className="text-xs text-slate-500">{item.hint}</span>
              </span>
              <ChevronRight size={18} className="text-slate-300" />
            </Link>
          )
        })}
      </section>

      <section className="overflow-hidden rounded-[1.5rem] bg-white shadow-[0_10px_28px_rgba(18,32,51,0.05)] ring-1 ring-[#E6EDF4]">
        <button type="button" onClick={openEdit} className="flex w-full items-center gap-3 px-4 py-3.5 text-left sm:px-5">
          <span className="flex h-11 w-11 items-center justify-center rounded-2xl bg-[#E8F0F8] text-[#2E5D90]">
            <UserRound size={18} />
          </span>
          <span className="min-w-0 flex-1">
            <span className="block font-semibold text-slate-900">Shaxsiy ma’lumotlar</span>
            <span className="text-xs text-slate-500">Ism, familiya, tug‘ilgan sana</span>
          </span>
          <span className="text-sm font-semibold text-[#2E5D90]">O‘zgartirish</span>
        </button>
        <button
          type="button"
          onClick={openEdit}
          className="flex w-full items-center gap-3 border-t border-[#E6EDF4] px-4 py-3.5 text-left sm:px-5"
        >
          <span className="flex h-11 w-11 items-center justify-center rounded-2xl bg-[#EEF8F1] text-emerald-700">
            <MapPin size={18} />
          </span>
          <span className="min-w-0 flex-1">
            <span className="block font-semibold text-slate-900">Yetkazish manzili</span>
            <span className="truncate text-xs text-slate-500">{place || 'Viloyat, tuman, MFY'}</span>
          </span>
          <span className="text-sm font-semibold text-[#2E5D90]">O‘zgartirish</span>
        </button>
      </section>

      <button
        type="button"
        onClick={() => void logout()}
        className="flex w-full items-center justify-center gap-2 rounded-[1.5rem] bg-white py-3.5 text-sm font-semibold text-red-600 ring-1 ring-[#E6EDF4]"
      >
        <LogOut size={16} />
        Hisobdan chiqish
      </button>

      {editing ? (
        <ProfileEditModal
          phone={user?.phone ?? ''}
          avatar={clearAvatar ? '' : avatarDraft}
          form={form}
          onChange={setForm}
          onAvatar={(file) => {
            void fileToDataUrl(file)
              .then((dataUrl) => {
                setAvatarDraft(dataUrl)
                setClearAvatar(false)
              })
              .catch((err: unknown) => toast.error(err instanceof Error ? err.message : 'Rasm yuklanmadi'))
          }}
          onClearAvatar={() => {
            setAvatarDraft('')
            setClearAvatar(true)
          }}
          onClose={() => setEditing(false)}
          onSubmit={(event) => void onSubmit(event)}
          busy={busy}
        />
      ) : null}
    </div>
  )
}
