import { ArrowDownToLine, Building2, ClipboardList, IdCard, LayoutDashboard, Layers, ListOrdered, Package, PackagePlus, ShoppingBag, Truck, UserRound, Wrench } from 'lucide-react'
import { useNavigate } from 'react-router-dom'
import { PanelShell, type PanelLink } from '../components/ui/PanelShell'
import { useAuth } from '../context/AuthContext'

const links: PanelLink[] = [
  { to: '/', label: 'Dashboard', icon: LayoutDashboard, end: true },
  { to: '/kontragents', label: 'Kontragentlar', icon: Building2 },
  { to: '/products', label: 'Mahsulotlar', icon: Package },
  { to: '/shop-templates', label: 'Mahsulot shablonlari', icon: ClipboardList },
  { to: '/shop-products', label: 'Do‘kon mahsulotlari', icon: PackagePlus },
  { to: '/shop-incomings', label: 'Kirim', icon: ArrowDownToLine },
  { to: '/local-shops', label: 'Mahalla do‘konlari', icon: ShoppingBag },
  { to: '/sellers', label: 'Sotuvchilar', icon: UserRound },
  { to: '/deliveries', label: 'Yetkazuvchilar', icon: Truck },
  { to: '/categories', label: 'Kategoriyalar', icon: Layers },
  { to: '/service-providers', label: 'Xizmat ko‘rsatuvchilar', icon: Wrench },
  { to: '/identifications', label: 'Identifikatsiya', icon: IdCard },
  { to: '/services', label: 'Xizmatlar', icon: ListOrdered },
]

export function DirectorLayout() {
  const { user, logout } = useAuth()
  const navigate = useNavigate()

  return (
    <PanelShell
      brand="Savdo uyi rahbari"
      mark="S"
      accent="#9F1239"
      links={links}
      userName={`${user?.first_name ?? ''} ${user?.last_name ?? ''}`.trim()}
      userMeta={user?.mfy_name}
      onLogout={async () => {
        await logout()
        navigate('/login', { replace: true })
      }}
    />
  )
}
