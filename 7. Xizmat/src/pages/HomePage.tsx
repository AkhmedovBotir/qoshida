import { motion } from 'motion/react'
import { Link } from 'react-router-dom'
import { ImageThumb } from '../components/ui/ImageLightbox'
import { API_URL } from '../constants/config'
import { useAuth } from '../context/AuthContext'
import { identStatusLabel } from '../types/identification'

export function HomePage() {
  const { user } = useAuth()
  const status = user?.identification_status ?? 'none'
  return (
    <motion.div initial={{ opacity: 0 }} animate={{ opacity: 1 }} className="mx-auto max-w-4xl">
      <h2 className="text-2xl font-bold text-slate-900">Xizmat ko‘rsatuvchi paneli</h2>
      <p className="mt-1 text-sm text-slate-500">Xush kelibsiz, {user?.name}.</p>
      {status !== 'approved' ? (
        <div className="mt-4 rounded-2xl border border-amber-200 bg-amber-50 px-4 py-3 text-sm text-amber-900">
          Identifikatsiya: <strong>{identStatusLabel[status]}</strong>. Xizmat joylash uchun{' '}
          <Link to="/identification" className="font-semibold underline">
            ma’lumot va hujjatlarni yuboring
          </Link>
          .
        </div>
      ) : null}
      <article className="mt-6 rounded-2xl border border-slate-200 bg-white p-5">
        {user?.image ? <ImageThumb src={`${API_URL}${user.image}`} className="mb-4 h-16 w-16 rounded-xl object-cover" /> : null}
        <p className="text-sm text-slate-700">
          Faoliyat: <span className="font-semibold">{user?.activity_name}</span>
        </p>
        <p className="mt-1 text-sm text-slate-700">
          Telefon: <span className="font-semibold">{user?.phone}</span>
        </p>
        <p className="mt-1 text-sm text-slate-700">
          Viloyat: <span className="font-semibold">{user?.region_name}</span>
        </p>
        <p className="mt-1 text-sm text-slate-700">
          Tuman: <span className="font-semibold">{user?.district_name}</span>
        </p>
        <p className="mt-1 text-sm text-slate-700">
          MFY: <span className="font-semibold">{user?.mfy_name}</span>
        </p>
      </article>
    </motion.div>
  )
}
