import { motion } from 'motion/react'
import { useAuth } from '../context/AuthContext'

export function HomePage() {
  const { user } = useAuth()
  return (
    <motion.div initial={{ opacity: 0 }} animate={{ opacity: 1 }} className="mx-auto max-w-4xl">
      <h2 className="text-2xl font-bold text-slate-900">Savdo uyi paneli</h2>
      <p className="mt-1 text-sm text-slate-500">
        Xush kelibsiz, {user?.first_name} {user?.last_name}.
      </p>
      <article className="mt-6 rounded-2xl border border-slate-200 bg-white p-5">
        <p className="text-sm text-slate-700">
          Viloyat: <span className="font-semibold">{user?.region_name}</span>
        </p>
        <p className="mt-1 text-sm text-slate-700">
          Tuman: <span className="font-semibold">{user?.district_name}</span>
        </p>
        <p className="mt-1 text-sm text-slate-700">
          MFY: <span className="font-semibold">{user?.mfy_name}</span>
        </p>
        <p className="mt-4 text-sm text-slate-500">Rahbar faqat o‘z MFYsi bo‘yicha ishlaydi.</p>
      </article>
    </motion.div>
  )
}
