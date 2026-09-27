import { createContext, useContext, useEffect, useMemo, useState, type ReactNode } from 'react'
import { ApiError, api } from '../lib/api'
import type { AdminUser } from '../types/admin'

type AuthContextValue = {
  user: AdminUser | null
  loading: boolean
  login: (username: string, password: string) => Promise<void>
  logout: () => Promise<void>
}

const AuthContext = createContext<AuthContextValue | null>(null)

export function AuthProvider({ children }: { children: ReactNode }) {
  const [user, setUser] = useState<AdminUser | null>(null)
  const [loading, setLoading] = useState(true)

  useEffect(() => {
    let cancelled = false
    api<AdminUser>('/api/v1/auth/me')
      .then((data) => {
        if (!cancelled) setUser(data)
      })
      .catch((err: unknown) => {
        if (!cancelled && err instanceof ApiError && err.status === 401) {
          setUser(null)
        }
      })
      .finally(() => {
        if (!cancelled) setLoading(false)
      })
    return () => {
      cancelled = true
    }
  }, [])

  const value = useMemo<AuthContextValue>(
    () => ({
      user,
      loading,
      async login(username, password) {
        const data = await api<AdminUser>('/api/v1/auth/login', {
          method: 'POST',
          body: { username, password },
        })
        setUser(data)
      },
      async logout() {
        try {
          await api('/api/v1/auth/logout', { method: 'POST', body: {}, silent: true })
        } finally {
          setUser(null)
        }
      },
    }),
    [user, loading],
  )

  return <AuthContext.Provider value={value}>{children}</AuthContext.Provider>
}

export function useAuth() {
  const ctx = useContext(AuthContext)
  if (!ctx) {
    throw new Error('useAuth AuthProvider ichida ishlatilishi kerak')
  }
  return ctx
}
