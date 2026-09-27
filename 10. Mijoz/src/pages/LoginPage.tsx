import { FormEvent, useEffect, useState } from 'react'
import { Navigate, useNavigate, useSearchParams } from 'react-router-dom'
import { AuthShell } from '../components/auth/AuthShell'
import { CustomerFields, type CustomerFormValue } from '../components/auth/CustomerFields'
import { OtpBoxes } from '../components/auth/OtpBoxes'
import { PhoneInput } from '../components/ui/PhoneInput'
import { useAuth } from '../context/AuthContext'
import { api, ApiError } from '../lib/api'
import type { AuthResult, SendCodeResult, VerifyResult } from '../types/customer'
import { toast } from '../lib/snack'

type Step = 'phone' | 'otp' | 'register'

const copy: Record<Step, { title: string; subtitle: string }> = {
  phone: {
    title: 'Raqamingizni kiriting',
    subtitle: 'SMS kod bilan raqam tekshiriladi. Hisob bo‘lsa, darhol kirasiz.',
  },
  otp: {
    title: 'SMS kodni yozing',
    subtitle: '6 xonali kod telefoningizga yuborildi.',
  },
  register: {
    title: 'Ma’lumotlaringiz',
    subtitle: 'Ism, familiya, tug‘ilgan sana va manzilni to‘ldiring.',
  },
}

const emptyForm = (): CustomerFormValue => ({
  first_name: '',
  last_name: '',
  birth_date: '',
  region_id: '',
  district_id: '',
  mfy_id: '',
})

export function LoginPage() {
  const { user, loading, setUser } = useAuth()
  const navigate = useNavigate()
  const [params] = useSearchParams()
  const from = params.get('from') || '/'
  const [step, setStep] = useState<Step>('phone')
  const [purpose, setPurpose] = useState('register')
  const [phone, setPhone] = useState('+998')
  const [code, setCode] = useState('')
  const [form, setForm] = useState<CustomerFormValue>(emptyForm)
  const [submitting, setSubmitting] = useState(false)
  const [seconds, setSeconds] = useState(0)

  useEffect(() => {
    if (seconds <= 0) return
    const timer = window.setInterval(() => setSeconds((v) => v - 1), 1000)
    return () => window.clearInterval(timer)
  }, [seconds])

  if (!loading && user) {
    return <Navigate to={from} replace />
  }

  async function sendCode() {
    const data = await api<SendCodeResult>('/api/v1/mijoz-auth/send-code', { method: 'POST', body: { phone } })
    setPurpose(data.purpose || 'register')
    setStep('otp')
    setSeconds(300)
    setCode('')
    toast.info(data.message || 'SMS kod yuborildi. Kod 5 daqiqa amal qiladi.')
  }

  async function onCheckPhone(event: FormEvent) {
    event.preventDefault()
    setSubmitting(true)
    try {
      await sendCode()
    } catch (err) {
      if (!(err instanceof ApiError)) toast.error('Kod yuborilmadi')
    } finally {
      setSubmitting(false)
    }
  }

  async function onVerify(event: FormEvent) {
    event.preventDefault()
    setSubmitting(true)
    try {
      const data = await api<VerifyResult>('/api/v1/mijoz-auth/verify-code', {
        method: 'POST',
        body: { phone, code, purpose },
      })
      if (data.customer) {
        setUser(data.customer)
        navigate(from, { replace: true })
        return
      }
      setStep('register')
      toast.info('Kod tasdiqlandi. Ma’lumotlaringizni kiriting.')
    } catch (err) {
      if (!(err instanceof ApiError)) toast.error('Kod tasdiqlanmadi')
    } finally {
      setSubmitting(false)
    }
  }

  async function onResend() {
    setSubmitting(true)
    try {
      await sendCode()
    } catch (err) {
      if (!(err instanceof ApiError)) toast.error('Qayta yuborilmadi')
    } finally {
      setSubmitting(false)
    }
  }

  async function onRegister(event: FormEvent) {
    event.preventDefault()
    setSubmitting(true)
    try {
      const data = await api<AuthResult>('/api/v1/mijoz-auth/register', {
        method: 'POST',
        body: {
          phone,
          first_name: form.first_name,
          last_name: form.last_name,
          birth_date: form.birth_date,
          region_id: form.region_id,
          district_id: form.district_id,
          mfy_id: form.mfy_id,
        },
      })
      setUser(data.customer)
      navigate(from, { replace: true })
    } catch (err) {
      if (!(err instanceof ApiError)) toast.error('Ro‘yxatdan o‘tish amalga oshmadi')
    } finally {
      setSubmitting(false)
    }
  }

  const submitClass =
    'mt-2 w-full rounded-full bg-[#101826] py-3.5 text-sm font-bold text-white transition hover:bg-[#1a2838] disabled:opacity-55'
  const { title, subtitle } = copy[step]
  const timer =
    seconds > 0 ? `${Math.floor(seconds / 60)}:${String(seconds % 60).padStart(2, '0')}` : ''

  return (
    <AuthShell
      step={step}
      title={title}
      subtitle={subtitle}
      backLabel={step === 'phone' ? 'Katalog' : 'Orqaga'}
      onBack={() => {
        if (step === 'phone') {
          navigate('/')
          return
        }
        if (step === 'register') {
          setStep('otp')
          return
        }
        setStep('phone')
        setCode('')
        setForm(emptyForm())
      }}
    >
      {step !== 'phone' ? (
        <p className="mb-4 inline-flex rounded-full bg-[#101826]/5 px-3 py-1 text-xs font-semibold text-[#101826]">
          {phone}
        </p>
      ) : null}

      {step === 'phone' ? (
        <form onSubmit={onCheckPhone} className="space-y-4">
          <PhoneInput size="lg" label="Telefon raqam" value={phone} onChange={setPhone} />
          <button type="submit" disabled={submitting} className={submitClass}>
            {submitting ? 'Yuborilmoqda...' : 'SMS kod olish'}
          </button>
        </form>
      ) : null}

      {step === 'otp' ? (
        <form onSubmit={onVerify} className="space-y-4">
          <OtpBoxes value={code} onChange={setCode} disabled={submitting} />
          <p className="text-center text-xs text-slate-500">
            {seconds > 0 ? `Kod ${timer} daqiqa amal qiladi` : 'Kod muddati tugagan'}
          </p>
          <button type="submit" disabled={submitting || code.length !== 6} className={submitClass}>
            {submitting ? 'Tekshirilmoqda...' : 'Tasdiqlash'}
          </button>
          <button
            type="button"
            disabled={submitting || seconds > 240}
            onClick={() => void onResend()}
            className="w-full text-sm font-semibold text-[#101826] disabled:opacity-40"
          >
            Kodni qayta yuborish
          </button>
        </form>
      ) : null}

      {step === 'register' ? (
        <form onSubmit={onRegister} className="space-y-4">
          <CustomerFields tone="auth" value={form} onChange={setForm} />
          <button type="submit" disabled={submitting} className={submitClass}>
            {submitting ? 'Saqlanmoqda...' : 'Ro‘yxatdan o‘tish'}
          </button>
        </form>
      ) : null}
    </AuthShell>
  )
}
