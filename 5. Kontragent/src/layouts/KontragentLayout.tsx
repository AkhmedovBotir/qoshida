import { LayoutDashboard, Package } from 'lucide-react'
import { useNavigate } from 'react-router-dom'
import { PanelShell, type PanelLink } from '../components/ui/PanelShell'
import { useAuth } from '../context/AuthContext'

const links: PanelLink[] = [
  { to: '/', label: 'Dashboard', icon: LayoutDashboard, end: true },
  { to: '/products', label: 'Mahsulotlar', icon: Package },
]

export function KontragentLayout() {
  const { user, logout } = useAuth()
  const navigate = useNavigate()

  return (
    <PanelShell
      mark="K"
      accent="#5B21B6"
      links={links}
      userName={user?.name}
      userMeta={user?.mfy_name}
      onLogout={async () => {
        await logout()
        navigate('/login', { replace: true })
      }}
    />
  )
}
