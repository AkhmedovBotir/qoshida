import { Navigate, Outlet } from 'react-router-dom'
import { useAuth } from '../../context/AuthContext'

export function ProtectedRoute() {
  const { user, loading } = useAuth()
  if (loading) {
    return <div className="flex min-h-svh items-center justify-center text-sm text-slate-500">Yuklanmoqda...</div>
  }
  if (!user) return <Navigate to="/login" replace />
  return <Outlet />
}
