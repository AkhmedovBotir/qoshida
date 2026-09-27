import { LayoutDashboard } from 'lucide-react'
import { useNavigate } from 'react-router-dom'
import { PanelShell, type PanelLink } from '../components/ui/PanelShell'
import { useAuth } from '../context/AuthContext'

const links: PanelLink[] = [{ to: '/', label: 'Dashboard', icon: LayoutDashboard, end: true }]

export function DeliveryLayout() {
  const { user, logout } = useAuth()
  const navigate = useNavigate()

  return (
    <PanelShell
      mark="Y"
      accent="#9A3412"
      links={links}
      userName={`${user?.first_name ?? ''} ${user?.last_name ?? ''}`.trim()}
      userMeta={user?.shop_name}
      onLogout={async () => {
        await logout()
        navigate('/login', { replace: true })
      }}
    />
  )
}
