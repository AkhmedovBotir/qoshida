import { createContext, useContext, useEffect, useMemo, useState, type ReactNode } from 'react'
import { api, ApiError } from '../lib/api'
import type { SellerUser } from '../types/seller'

type AuthContextValue = {
  user: SellerUser | null
  loading: boolean
  setUser: (user: SellerUser | null) => void
  logout: () => Promise<void>
}

const AuthContext = createContext<AuthContextValue | null>(null)

export function AuthProvider({ children }: { children: ReactNode }) {
  const [user, setUser] = useState<SellerUser | null>(null)
  const [loading, setLoading] = useState(true)

  useEffect(() => {
    let cancelled = false
    api<SellerUser>('/api/v1/seller-auth/me')
      .then((data) => {
        if (!cancelled) setUser(data)
      })
      .catch((err: unknown) => {
        if (!cancelled && err instanceof ApiError && err.status === 401) setUser(null)
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
      setUser,
      async logout() {
        try {
          await api('/api/v1/seller-auth/logout', { method: 'POST', body: {}, silent: true })
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
  if (!ctx) throw new Error('useAuth AuthProvider ichida ishlatilishi kerak')
  return ctx
}
