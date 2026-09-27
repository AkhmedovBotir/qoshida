import { IdCard, LayoutDashboard, ListOrdered } from 'lucide-react'
import { useNavigate } from 'react-router-dom'
import { PanelShell, type PanelLink } from '../components/ui/PanelShell'
import { useAuth } from '../context/AuthContext'

const links: PanelLink[] = [
  { to: '/', label: 'Dashboard', icon: LayoutDashboard, end: true },
  { to: '/identification', label: 'Identifikatsiya', icon: IdCard },
  { to: '/services', label: 'Xizmatlar', icon: ListOrdered },
]

export function ProviderLayout() {
  const { user, logout } = useAuth()
  const navigate = useNavigate()

  return (
    <PanelShell
      mark="X"
      accent="#075985"
      links={links}
      userName={user?.name}
      userMeta={user?.activity_name}
      onLogout={async () => {
        await logout()
        navigate('/login', { replace: true })
      }}
    />
  )
}
