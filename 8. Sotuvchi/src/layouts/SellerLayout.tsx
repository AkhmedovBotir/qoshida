import { ArrowDownToLine, LayoutDashboard, PackagePlus } from 'lucide-react'
import { useNavigate } from 'react-router-dom'
import { PanelShell, type PanelLink } from '../components/ui/PanelShell'
import { useAuth } from '../context/AuthContext'

const links: PanelLink[] = [
  { to: '/', label: 'Dashboard', icon: LayoutDashboard, end: true },
  { to: '/shop-products', label: 'Mahsulotlar', icon: PackagePlus },
  { to: '/shop-incomings', label: 'Kirim', icon: ArrowDownToLine },
]

export function SellerLayout() {
  const { user, logout } = useAuth()
  const navigate = useNavigate()

  return (
    <PanelShell
      mark="S"
      accent="#166534"
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
