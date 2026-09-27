import { ArrowDownToLine, LayoutDashboard, PackagePlus, Truck, UserRound } from 'lucide-react'
import { useNavigate } from 'react-router-dom'
import { PanelShell, type PanelLink } from '../components/ui/PanelShell'
import { useAuth } from '../context/AuthContext'

const links: PanelLink[] = [
  { to: '/', label: 'Dashboard', icon: LayoutDashboard, end: true },
  { to: '/sellers', label: 'Sotuvchilar', icon: UserRound },
  { to: '/deliveries', label: 'Yetkazuvchilar', icon: Truck },
  { to: '/shop-products', label: 'Mahsulotlar', icon: PackagePlus },
  { to: '/shop-incomings', label: 'Kirim', icon: ArrowDownToLine },
]

export function ShopLayout() {
  const { user, logout } = useAuth()
  const navigate = useNavigate()

  return (
    <PanelShell
      mark="D"
      accent="#115E59"
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
