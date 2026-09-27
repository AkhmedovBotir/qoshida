import { motion } from 'motion/react'
import { FormEvent, useState } from 'react'
import { Navigate, useNavigate } from 'react-router-dom'
import { PasswordInput } from '../components/ui/PasswordInput'
import { APP_COLOR, APP_NAME } from '../constants/config'
import { useAuth } from '../context/AuthContext'
import { ApiError } from '../lib/api'
import { toast } from '../lib/snack'

export function LoginPage() {
  const { user, loading, login } = useAuth()
  const navigate = useNavigate()
  const [username, setUsername] = useState('')
  const [password, setPassword] = useState('')
  const [submitting, setSubmitting] = useState(false)

  if (!loading && user) {
    return <Navigate to="/" replace />
  }

  async function onSubmit(event: FormEvent) {
    event.preventDefault()
    setSubmitting(true)
    try {
      await login(username.trim(), password)
      navigate('/', { replace: true })
    } catch (err) {
      if (!(err instanceof ApiError)) toast.error('Kirish amalga oshmadi')
    } finally {
      setSubmitting(false)
    }
  }

  return (
    <main className="flex min-h-svh items-center justify-center bg-[#F4F6F9] px-4">
      <motion.form
        onSubmit={onSubmit}
        initial={{ opacity: 0, y: 16 }}
        animate={{ opacity: 1, y: 0 }}
        className="w-full max-w-md rounded-2xl border border-slate-200 bg-white p-6 shadow-sm"
      >
        <p className="text-xs font-bold uppercase tracking-[1.2px]" style={{ color: APP_COLOR }}>
          Qoshida
        </p>
        <h1 className="mt-2 text-2xl font-bold text-slate-900">{APP_NAME} kirish</h1>
        <p className="mt-1 text-sm text-slate-500">Faqat mavjud admin hisobi bilan kiring.</p>

        <label className="mt-6 block text-sm font-medium text-slate-700">
          Username
          <input
            value={username}
            onChange={(e) => setUsername(e.target.value)}
            autoComplete="username"
            required
            className="mt-1.5 w-full rounded-xl border border-slate-200 px-3 py-2.5 text-slate-900 outline-none focus:border-slate-400"
          />
        </label>
        <PasswordInput
          label="Parol"
          className="mt-4"
          value={password}
          onChange={(e) => setPassword(e.target.value)}
          autoComplete="current-password"
          required
        />

        <button
          type="submit"
          disabled={submitting}
          className="mt-6 w-full rounded-xl py-2.5 text-sm font-semibold text-white disabled:opacity-60"
          style={{ backgroundColor: APP_COLOR }}
        >
          {submitting ? 'Tekshirilmoqda...' : 'Kirish'}
        </button>
      </motion.form>
    </main>
  )
}
