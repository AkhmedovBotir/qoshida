import { Monitor, Smartphone, Tablet } from 'lucide-react'
import { motion } from 'motion/react'
import { useEffect, useState } from 'react'
import { FastImage } from '../components/ui/FastImage'
import { Screen } from '../components/ui/Screen'
import { Skeleton, SkeletonCard } from '../components/ui/Skeleton'
import { APP_COLOR, APP_NAME, APP_ROLE, API_URL } from '../constants/config'
import { usePlatform } from '../hooks/usePlatform'
import { api } from '../lib/api'

const SAMPLE_IMAGES = [
  'https://images.unsplash.com/photo-1556742049-0cfed4f6a45d?auto=format&fit=crop&w=640&q=70',
  'https://images.unsplash.com/photo-1542838132-92c53300491e?auto=format&fit=crop&w=640&q=70',
  'https://images.unsplash.com/photo-1504674900247-0877df9cc836?auto=format&fit=crop&w=640&q=70',
]

const STACK = ['Tailwind CSS', 'Framer Motion', 'React Native', 'Capacitor'] as const

type Health = {
  status: string
  service: string
}

export function HomeScreen() {
  const platform = usePlatform()
  const [ready, setReady] = useState(false)
  const [health, setHealth] = useState('tekshirilmoqda...')

  useEffect(() => {
    const timer = window.setTimeout(() => setReady(true), 700)
    return () => window.clearTimeout(timer)
  }, [])

  async function checkBackend() {
    try {
      const data = await api<Health>('/api/v1/health')
      setHealth(`${data.status} · ${data.service}`)
    } catch {
      setHealth('backend hali ulanmagan')
    }
  }

  const PlatformIcon = platform === 'web' ? Monitor : Smartphone

  return (
    <Screen>
      <motion.div
        initial={{ opacity: 0, y: 18 }}
        animate={{ opacity: 1, y: 0 }}
        transition={{ duration: 0.45, ease: 'easeOut' }}
        className="flex flex-col gap-6"
      >
        <header>
          <p className="text-[13px] font-bold uppercase tracking-[1.2px]" style={{ color: APP_COLOR }}>
            Qoshida
          </p>
          <h1 className="mt-1.5 text-3xl font-bold text-slate-900">{APP_NAME}</h1>
          <p className="mt-2 text-[15px] leading-6 text-slate-500">
            Rol: {APP_ROLE} · Platforma: {platform} · API: {API_URL}
          </p>
        </header>

        <div className="flex flex-wrap gap-2">
          {STACK.map((item, index) => (
            <motion.span
              key={item}
              initial={{ opacity: 0, scale: 0.92 }}
              animate={{ opacity: 1, scale: 1 }}
              transition={{ delay: 0.08 * index, duration: 0.28 }}
              className="rounded-full border border-slate-200 bg-white px-3 py-1 text-xs font-semibold text-slate-600"
            >
              {item}
            </motion.span>
          ))}
        </div>

        <section className="flex flex-col gap-2.5 rounded-2xl border border-slate-200 bg-white p-4">
          <div className="flex items-center gap-2">
            <PlatformIcon size={18} color={APP_COLOR} />
            <h2 className="text-base font-bold text-slate-900">Holat</h2>
          </div>
          {ready ? (
            <p className="text-sm leading-5 text-slate-500">
              Bitta kod: web + iOS + APK. Backend: {health}
            </p>
          ) : (
            <div className="flex flex-col gap-2">
              <Skeleton width="88%" height={14} />
              <Skeleton width="60%" height={14} />
            </div>
          )}
          <button
            type="button"
            onClick={() => {
              void checkBackend()
            }}
            className="self-start rounded-[10px] px-3.5 py-2.5 text-sm font-semibold text-white"
            style={{ backgroundColor: APP_COLOR }}
          >
            Backendni tekshirish
          </button>
        </section>

        <h2 className="text-lg font-bold text-slate-900">Tezkor rasmlar</h2>
        <div className="grid gap-4">
          {(ready ? SAMPLE_IMAGES : ['', '', '']).map((uri, index) =>
            uri ? (
              <motion.div
                key={uri}
                initial={{ opacity: 0, y: 12 }}
                animate={{ opacity: 1, y: 0 }}
                transition={{ delay: 0.06 * index }}
              >
                <FastImage uri={uri} alt={`Namuna rasm ${index + 1}`} height={150} />
              </motion.div>
            ) : (
              <SkeletonCard key={`skeleton-${index}`} />
            ),
          )}
        </div>

        <div className="flex items-center gap-2 rounded-2xl border border-dashed border-slate-300 bg-white/70 p-4">
          <Tablet size={18} color="#5B6B7C" />
          <p className="flex-1 text-sm leading-5 text-slate-500">
            Mobile: npm run mobile:android yoki npm run mobile:ios
          </p>
        </div>
      </motion.div>
    </Screen>
  )
}
