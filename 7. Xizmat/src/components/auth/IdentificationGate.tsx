import { Navigate, Outlet, useLocation } from 'react-router-dom'
import { useAuth } from '../../context/AuthContext'

export function IdentificationGate() {
  const { user } = useAuth()
  const location = useLocation()
  const status = user?.identification_status ?? 'none'
  const onIdent = location.pathname === '/identification'
  if ((status === 'none' || status === 'rejected') && !onIdent) {
    return <Navigate to="/identification" replace />
  }
  return <Outlet />
}
