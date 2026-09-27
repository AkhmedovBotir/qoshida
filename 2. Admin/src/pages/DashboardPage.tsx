import { Briefcase, Building2, Layers, MapPin, MessageSquareText, Shield, ShoppingBag, Store, Truck, UserRound, Users, Wrench } from 'lucide-react'
import { motion } from 'motion/react'
import { useEffect, useState } from 'react'
import { Link } from 'react-router-dom'
import { Skeleton } from '../components/ui/Skeleton'
import { APP_COLOR } from '../constants/config'
import { useAuth } from '../context/AuthContext'
import { api } from '../lib/api'

export function DashboardPage() {
  const { user } = useAuth()
  const [count, setCount] = useState<number | null>(null)
  const [regions, setRegions] = useState<number | null>(null)
  const [categories, setCategories] = useState<number | null>(null)
  const [activityTypes, setActivityTypes] = useState<number | null>(null)
  const [commentTemplates, setCommentTemplates] = useState<number | null>(null)
  const [managers, setManagers] = useState<number | null>(null)
  const [shopDirectors, setShopDirectors] = useState<number | null>(null)
  const [kontragents, setKontragents] = useState<number | null>(null)
  const [localShops, setLocalShops] = useState<number | null>(null)
  const [sellers, setSellers] = useState<number | null>(null)
  const [deliveries, setDeliveries] = useState<number | null>(null)
  const [serviceProviders, setServiceProviders] = useState<number | null>(null)

  useEffect(() => {
    api<{ admins: number }>('/api/v1/dashboard/stats')
      .then((data) => setCount(data.admins))
      .catch(() => setCount(0))
    api<{ total: number }>('/api/v1/regions/stats')
      .then((data) => setRegions(data.total))
      .catch(() => setRegions(0))
    api<{ total: number }>('/api/v1/categories/stats')
      .then((data) => setCategories(data.total))
      .catch(() => setCategories(0))
    api<{ total: number }>('/api/v1/activity-types/stats')
      .then((data) => setActivityTypes(data.total))
      .catch(() => setActivityTypes(0))
    api<{ total: number }>('/api/v1/comment-templates/stats')
      .then((data) => setCommentTemplates(data.total))
      .catch(() => setCommentTemplates(0))
    api<{ total: number }>('/api/v1/managers/stats')
      .then((data) => setManagers(data.total))
      .catch(() => setManagers(0))
    api<{ total: number }>('/api/v1/shop-directors/stats')
      .then((data) => setShopDirectors(data.total))
      .catch(() => setShopDirectors(0))
    api<{ total: number }>('/api/v1/kontragents/stats')
      .then((data) => setKontragents(data.total))
      .catch(() => setKontragents(0))
    api<{ total: number }>('/api/v1/local-shops/stats')
      .then((data) => setLocalShops(data.total))
      .catch(() => setLocalShops(0))
    api<{ total: number }>('/api/v1/sellers/stats')
      .then((data) => setSellers(data.total))
      .catch(() => setSellers(0))
    api<{ total: number }>('/api/v1/deliveries/stats')
      .then((data) => setDeliveries(data.total))
      .catch(() => setDeliveries(0))
    api<{ total: number }>('/api/v1/service-providers/stats')
      .then((data) => setServiceProviders(data.total))
      .catch(() => setServiceProviders(0))
  }, [])

  return (
    <motion.div initial={{ opacity: 0, y: 12 }} animate={{ opacity: 1, y: 0 }} className="mx-auto max-w-5xl">
      <h2 className="text-2xl font-bold text-slate-900">Dashboard</h2>
      <p className="mt-1 text-sm text-slate-500">
        Xush kelibsiz, {user?.first_name} {user?.last_name}.
      </p>

      <div className="mt-6 grid gap-4 md:grid-cols-2 xl:grid-cols-3">
        <article className="rounded-2xl border border-slate-200 bg-white p-5">
          <div className="flex items-center gap-3">
            <span className="rounded-xl p-2 text-white" style={{ backgroundColor: APP_COLOR }}>
              <Shield size={18} />
            </span>
            <h3 className="font-semibold text-slate-900">Adminlar</h3>
          </div>
          {count === null ? (
            <div className="mt-4">
              <Skeleton width={80} height={32} />
            </div>
          ) : (
            <p className="mt-4 text-3xl font-bold text-slate-900">{count}</p>
          )}
          <Link to="/admins" className="mt-3 inline-block text-sm font-semibold" style={{ color: APP_COLOR }}>
            Jadvalga o'tish
          </Link>
        </article>

        <article className="rounded-2xl border border-slate-200 bg-white p-5">
          <div className="flex items-center gap-3">
            <span className="rounded-xl p-2 text-white" style={{ backgroundColor: APP_COLOR }}>
              <MapPin size={18} />
            </span>
            <h3 className="font-semibold text-slate-900">Hududlar</h3>
          </div>
          {regions === null ? (
            <div className="mt-4">
              <Skeleton width={80} height={32} />
            </div>
          ) : (
            <p className="mt-4 text-3xl font-bold text-slate-900">{regions}</p>
          )}
          <Link to="/regions" className="mt-3 inline-block text-sm font-semibold" style={{ color: APP_COLOR }}>
            Jadvalga o'tish
          </Link>
        </article>

        <article className="rounded-2xl border border-slate-200 bg-white p-5">
          <div className="flex items-center gap-3">
            <span className="rounded-xl p-2 text-white" style={{ backgroundColor: APP_COLOR }}>
              <Layers size={18} />
            </span>
            <h3 className="font-semibold text-slate-900">Kategoriyalar</h3>
          </div>
          {categories === null ? (
            <div className="mt-4">
              <Skeleton width={80} height={32} />
            </div>
          ) : (
            <p className="mt-4 text-3xl font-bold text-slate-900">{categories}</p>
          )}
          <Link to="/categories" className="mt-3 inline-block text-sm font-semibold" style={{ color: APP_COLOR }}>
            Jadvalga o'tish
          </Link>
        </article>

        <article className="rounded-2xl border border-slate-200 bg-white p-5">
          <div className="flex items-center gap-3">
            <span className="rounded-xl p-2 text-white" style={{ backgroundColor: APP_COLOR }}>
              <Briefcase size={18} />
            </span>
            <h3 className="font-semibold text-slate-900">Faoliyat turlari</h3>
          </div>
          {activityTypes === null ? (
            <div className="mt-4">
              <Skeleton width={80} height={32} />
            </div>
          ) : (
            <p className="mt-4 text-3xl font-bold text-slate-900">{activityTypes}</p>
          )}
          <Link to="/activity-types" className="mt-3 inline-block text-sm font-semibold" style={{ color: APP_COLOR }}>
            Jadvalga o'tish
          </Link>
        </article>

        <article className="rounded-2xl border border-slate-200 bg-white p-5">
          <div className="flex items-center gap-3">
            <span className="rounded-xl p-2 text-white" style={{ backgroundColor: APP_COLOR }}>
              <MessageSquareText size={18} />
            </span>
            <h3 className="font-semibold text-slate-900">Kommentariya shablonlari</h3>
          </div>
          {commentTemplates === null ? (
            <div className="mt-4">
              <Skeleton width={80} height={32} />
            </div>
          ) : (
            <p className="mt-4 text-3xl font-bold text-slate-900">{commentTemplates}</p>
          )}
          <Link to="/settings?tab=comments" className="mt-3 inline-block text-sm font-semibold" style={{ color: APP_COLOR }}>
            Sozlamalarga o'tish
          </Link>
        </article>

        <article className="rounded-2xl border border-slate-200 bg-white p-5">
          <div className="flex items-center gap-3">
            <span className="rounded-xl p-2 text-white" style={{ backgroundColor: APP_COLOR }}>
              <Users size={18} />
            </span>
            <h3 className="font-semibold text-slate-900">Menejerlar</h3>
          </div>
          {managers === null ? (
            <div className="mt-4">
              <Skeleton width={80} height={32} />
            </div>
          ) : (
            <p className="mt-4 text-3xl font-bold text-slate-900">{managers}</p>
          )}
          <Link to="/managers" className="mt-3 inline-block text-sm font-semibold" style={{ color: APP_COLOR }}>
            Jadvalga o'tish
          </Link>
        </article>

        <article className="rounded-2xl border border-slate-200 bg-white p-5">
          <div className="flex items-center gap-3">
            <span className="rounded-xl p-2 text-white" style={{ backgroundColor: APP_COLOR }}>
              <Store size={18} />
            </span>
            <h3 className="font-semibold text-slate-900">Savdo uyi rahbarlari</h3>
          </div>
          {shopDirectors === null ? (
            <div className="mt-4">
              <Skeleton width={80} height={32} />
            </div>
          ) : (
            <p className="mt-4 text-3xl font-bold text-slate-900">{shopDirectors}</p>
          )}
          <Link to="/shop-directors" className="mt-3 inline-block text-sm font-semibold" style={{ color: APP_COLOR }}>
            Jadvalga o'tish
          </Link>
        </article>

        <article className="rounded-2xl border border-slate-200 bg-white p-5">
          <div className="flex items-center gap-3">
            <span className="rounded-xl p-2 text-white" style={{ backgroundColor: APP_COLOR }}>
              <Building2 size={18} />
            </span>
            <h3 className="font-semibold text-slate-900">Kontragentlar</h3>
          </div>
          {kontragents === null ? (
            <div className="mt-4">
              <Skeleton width={80} height={32} />
            </div>
          ) : (
            <p className="mt-4 text-3xl font-bold text-slate-900">{kontragents}</p>
          )}
          <Link to="/kontragents" className="mt-3 inline-block text-sm font-semibold" style={{ color: APP_COLOR }}>
            Jadvalga o'tish
          </Link>
        </article>

        <article className="rounded-2xl border border-slate-200 bg-white p-5">
          <div className="flex items-center gap-3">
            <span className="rounded-xl p-2 text-white" style={{ backgroundColor: APP_COLOR }}>
              <ShoppingBag size={18} />
            </span>
            <h3 className="font-semibold text-slate-900">Mahalla do‘konlari</h3>
          </div>
          {localShops === null ? (
            <div className="mt-4">
              <Skeleton width={80} height={32} />
            </div>
          ) : (
            <p className="mt-4 text-3xl font-bold text-slate-900">{localShops}</p>
          )}
          <Link to="/local-shops" className="mt-3 inline-block text-sm font-semibold" style={{ color: APP_COLOR }}>
            Jadvalga o'tish
          </Link>
        </article>

        <article className="rounded-2xl border border-slate-200 bg-white p-5">
          <div className="flex items-center gap-3">
            <span className="rounded-xl p-2 text-white" style={{ backgroundColor: APP_COLOR }}>
              <UserRound size={18} />
            </span>
            <h3 className="font-semibold text-slate-900">Sotuvchilar</h3>
          </div>
          {sellers === null ? (
            <div className="mt-4">
              <Skeleton width={80} height={32} />
            </div>
          ) : (
            <p className="mt-4 text-3xl font-bold text-slate-900">{sellers}</p>
          )}
          <Link to="/sellers" className="mt-3 inline-block text-sm font-semibold" style={{ color: APP_COLOR }}>
            Jadvalga o'tish
          </Link>
        </article>

        <article className="rounded-2xl border border-slate-200 bg-white p-5">
          <div className="flex items-center gap-3">
            <span className="rounded-xl p-2 text-white" style={{ backgroundColor: APP_COLOR }}>
              <Truck size={18} />
            </span>
            <h3 className="font-semibold text-slate-900">Yetkazuvchilar</h3>
          </div>
          {deliveries === null ? (
            <div className="mt-4">
              <Skeleton width={80} height={32} />
            </div>
          ) : (
            <p className="mt-4 text-3xl font-bold text-slate-900">{deliveries}</p>
          )}
          <Link to="/deliveries" className="mt-3 inline-block text-sm font-semibold" style={{ color: APP_COLOR }}>
            Jadvalga o'tish
          </Link>
        </article>

        <article className="rounded-2xl border border-slate-200 bg-white p-5">
          <div className="flex items-center gap-3">
            <span className="rounded-xl p-2 text-white" style={{ backgroundColor: APP_COLOR }}>
              <Wrench size={18} />
            </span>
            <h3 className="font-semibold text-slate-900">Xizmat ko‘rsatuvchilar</h3>
          </div>
          {serviceProviders === null ? (
            <div className="mt-4">
              <Skeleton width={80} height={32} />
            </div>
          ) : (
            <p className="mt-4 text-3xl font-bold text-slate-900">{serviceProviders}</p>
          )}
          <Link to="/service-providers" className="mt-3 inline-block text-sm font-semibold" style={{ color: APP_COLOR }}>
            Jadvalga o'tish
          </Link>
        </article>

        <article className="rounded-2xl border border-slate-200 bg-white p-5">
          <div className="flex items-center gap-3">
            <span className="rounded-xl bg-slate-100 p-2 text-slate-700">
              <UserRound size={18} />
            </span>
            <h3 className="font-semibold text-slate-900">Sizning rol</h3>
          </div>
          <p className="mt-4 text-3xl font-bold capitalize text-slate-900">{user?.role}</p>
          <p className="mt-2 text-sm text-slate-500">General admin faqat `create-admin` orqali 1 marta yaratiladi.</p>
        </article>
      </div>
    </motion.div>
  )
}
