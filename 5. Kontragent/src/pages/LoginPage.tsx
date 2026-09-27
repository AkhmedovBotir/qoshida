import { motion } from 'motion/react'
import { FormEvent, useEffect, useState } from 'react'
import { Navigate, useNavigate } from 'react-router-dom'
import { OtpBoxes } from '../components/auth/OtpBoxes'
import { PasswordInput } from '../components/ui/PasswordInput'
import { PhoneInput } from '../components/ui/PhoneInput'
import { APP_COLOR, APP_NAME } from '../constants/config'
import { useAuth } from '../context/AuthContext'
import { api, ApiError } from '../lib/api'
import type { KontragentUser, PhoneCheck } from '../types/kontragent'
import { toast } from '../lib/snack'

type Step = 'phone' | 'otp' | 'setup' | 'password'

export function LoginPage() {
  const { user, loading, setUser } = useAuth()
  const navigate = useNavigate()
  const [step, setStep] = useState<Step>('phone')
  const [phone, setPhone] = useState('+998')
  const [code, setCode] = useState('')
  const [password, setPassword] = useState('')
  const [submitting, setSubmitting] = useState(false)
  const [seconds, setSeconds] = useState(0)

  useEffect(() => {
    if (seconds <= 0) return
    const timer = window.setInterval(() => setSeconds((v) => v - 1), 1000)
    return () => window.clearInterval(timer)
  }, [seconds])

  if (!loading && user) {
    return <Navigate to="/" replace />
  }

  async function onCheckPhone(event: FormEvent) {
    event.preventDefault()
    setSubmitting(true)
    try {
      const check = await api<PhoneCheck>('/api/v1/kontragent-auth/check-phone', {
        method: 'POST',
        body: { phone },
      })
      if (!check.exists) {
        toast.error(check.message || "Bu raqam ro'yxatga olinmagan")
        return
      }
      if (!check.active) {
        toast.error(check.message || 'Hisob faol emas')
        return
      }
      if (check.needs_setup) {
        await api('/api/v1/kontragent-auth/send-code', { method: 'POST', body: { phone } })
        setStep('otp')
        setSeconds(300)
        toast.info('SMS kod yuborildi. Kod 5 daqiqa amal qiladi.')
        return
      }
      setStep('password')
    } catch (err) {
      if (!(err instanceof ApiError)) toast.error('Tekshirib bo‘lmadi')
    } finally {
      setSubmitting(false)
    }
  }

  async function onVerify(event: FormEvent) {
    event.preventDefault()
    setSubmitting(true)
    try {
      await api('/api/v1/kontragent-auth/verify-code', { method: 'POST', body: { phone, code } })
      setStep('setup')
      toast.info('Kod tasdiqlandi. Endi parol o‘rnating.')
    } catch (err) {
      if (!(err instanceof ApiError)) toast.error('Kod tasdiqlanmadi')
    } finally {
      setSubmitting(false)
    }
  }

  async function onResend() {
    setSubmitting(true)
    try {
      await api('/api/v1/kontragent-auth/resend-code', { method: 'POST', body: { phone } })
      setSeconds(300)
      setCode('')
      toast.info('Kod qayta yuborildi.')
    } catch (err) {
      if (!(err instanceof ApiError)) toast.error('Qayta yuborilmadi')
    } finally {
      setSubmitting(false)
    }
  }

  async function onSetPassword(event: FormEvent) {
    event.preventDefault()
    setSubmitting(true)
    try {
      const data = await api<{ kontragent: KontragentUser }>('/api/v1/kontragent-auth/set-password', {
        method: 'POST',
        body: { phone, password },
      })
      setUser(data.kontragent)
      navigate('/', { replace: true })
    } catch (err) {
      if (!(err instanceof ApiError)) toast.error('Parol o‘rnatilmadi')
    } finally {
      setSubmitting(false)
    }
  }

  async function onLogin(event: FormEvent) {
    event.preventDefault()
    setSubmitting(true)
    try {
      const data = await api<{ kontragent: KontragentUser }>('/api/v1/kontragent-auth/login', {
        method: 'POST',
        body: { phone, password },
      })
      setUser(data.kontragent)
      navigate('/', { replace: true })
    } catch (err) {
      if (!(err instanceof ApiError)) toast.error('Kirish amalga oshmadi')
    } finally {
      setSubmitting(false)
    }
  }

  return (
    <main className="flex min-h-svh items-center justify-center bg-[#F4F6F9] px-4">
      <motion.div
        initial={{ opacity: 0, y: 16 }}
        animate={{ opacity: 1, y: 0 }}
        className="w-full max-w-md rounded-2xl border border-slate-200 bg-white p-6 shadow-sm"
      >
        <p className="text-xs font-bold uppercase tracking-[1.2px]" style={{ color: APP_COLOR }}>
          Qoshida
        </p>
        <h1 className="mt-2 text-2xl font-bold text-slate-900">{APP_NAME} kirish</h1>
        <p className="mt-1 text-sm text-slate-500">
          Avval telefon raqamingizni kiriting. Faqat ro‘yxatga olingan kontragent kira oladi.
        </p>

        {step === 'phone' ? (
          <form onSubmit={onCheckPhone} className="mt-6 space-y-4">
            <PhoneInput value={phone} onChange={setPhone} />
            <button
              type="submit"
              disabled={submitting}
              className="w-full rounded-xl py-2.5 text-sm font-semibold text-white disabled:opacity-60"
              style={{ backgroundColor: APP_COLOR }}
            >
              {submitting ? 'Tekshirilmoqda...' : 'Davom etish'}
            </button>
          </form>
        ) : null}

        {step === 'otp' ? (
          <form onSubmit={onVerify} className="mt-6 space-y-4">
            <p className="text-sm text-slate-600">{phone} raqamiga yuborilgan 6 xonali kodni kiriting.</p>
            <OtpBoxes value={code} onChange={setCode} disabled={submitting} />
            <p className="text-center text-xs text-slate-500">
              {seconds > 0 ? `Kod ${Math.floor(seconds / 60)}:${String(seconds % 60).padStart(2, '0')} daqiqa amal qiladi` : 'Kod muddati tugagan'}
            </p>
            <button
              type="submit"
              disabled={submitting || code.length !== 6}
              className="w-full rounded-xl py-2.5 text-sm font-semibold text-white disabled:opacity-60"
              style={{ backgroundColor: APP_COLOR }}
            >
              {submitting ? 'Tekshirilmoqda...' : 'Kodni tasdiqlash'}
            </button>
            <button
              type="button"
              disabled={submitting || seconds > 240}
              onClick={() => void onResend()}
              className="w-full text-sm font-semibold disabled:opacity-40"
              style={{ color: APP_COLOR }}
            >
              Kodni qayta yuborish
            </button>
          </form>
        ) : null}

        {step === 'setup' ? (
          <form onSubmit={onSetPassword} className="mt-6 space-y-4">
            <PasswordInput
              label="Yangi parol"
              value={password}
              onChange={(e) => setPassword(e.target.value)}
              required
              autoComplete="new-password"
            />
            <button
              type="submit"
              disabled={submitting || !password.trim()}
              className="w-full rounded-xl py-2.5 text-sm font-semibold text-white disabled:opacity-60"
              style={{ backgroundColor: APP_COLOR }}
            >
              {submitting ? 'Saqlanmoqda...' : 'Parolni o‘rnatish'}
            </button>
          </form>
        ) : null}

        {step === 'password' ? (
          <form onSubmit={onLogin} className="mt-6 space-y-4">
            <PasswordInput
              label="Parol"
              value={password}
              onChange={(e) => setPassword(e.target.value)}
              required
              autoComplete="current-password"
            />
            <button
              type="submit"
              disabled={submitting}
              className="w-full rounded-xl py-2.5 text-sm font-semibold text-white disabled:opacity-60"
              style={{ backgroundColor: APP_COLOR }}
            >
              {submitting ? 'Tekshirilmoqda...' : 'Kirish'}
            </button>
          </form>
        ) : null}

        {step !== 'phone' ? (
          <button
            type="button"
            onClick={() => {
              setStep('phone')
              setCode('')
              setPassword('')
            }}
            className="mt-4 text-sm text-slate-500"
          >
            Boshqa raqam
          </button>
        ) : null}
      </motion.div>
    </main>
  )
}
