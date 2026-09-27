import { Navigate, Outlet, useLocation } from 'react-router-dom'
import { useAuth } from '../../context/AuthContext'
import { Skeleton } from '../ui/Skeleton'

export function RequireAuth() {
  const { user, loading } = useAuth()
  const location = useLocation()

  if (loading) {
    return (
      <div className="space-y-3">
        <Skeleton height={28} width="40%" />
        <Skeleton height={120} />
      </div>
    )
  }

  if (!user) {
    return <Navigate to={`/login?from=${encodeURIComponent(location.pathname + location.search)}`} replace />
  }

  return <Outlet />
}
