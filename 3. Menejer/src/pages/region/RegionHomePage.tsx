import { motion } from 'motion/react'
import { Link } from 'react-router-dom'
import { APP_COLOR } from '../../constants/config'
import { useAuth } from '../../context/AuthContext'

export function RegionHomePage() {
  const { user } = useAuth()
  return (
    <motion.div initial={{ opacity: 0 }} animate={{ opacity: 1 }} className="mx-auto max-w-4xl">
      <h2 className="text-2xl font-bold text-slate-900">Viloyat paneli</h2>
      <p className="mt-1 text-sm text-slate-500">
        Xush kelibsiz, {user?.first_name} {user?.last_name}. Hudud: {user?.region_name}.
      </p>
      <article className="mt-6 rounded-2xl border border-slate-200 bg-white p-5">
        <h3 className="font-semibold text-slate-900">Tuman menejerlari</h3>
        <p className="mt-2 text-sm text-slate-500">
          O‘z viloyatingizdagi tuman menejerlarini ro‘yxatga oling va tahrirlang.
        </p>
        <Link to="/region/district-managers" className="mt-3 inline-block text-sm font-semibold" style={{ color: APP_COLOR }}>
          Ro‘yxatga o‘tish
        </Link>
      </article>
    </motion.div>
  )
}
