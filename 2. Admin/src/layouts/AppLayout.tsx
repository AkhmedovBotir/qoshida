import { ArrowDownToLine, Briefcase, Building2, ClipboardList, IdCard, LayoutDashboard, Layers, ListOrdered, MapPin, Package, PackagePlus, Settings, Shield, ShoppingBag, Store, Truck, UserRound, Users, Wrench } from 'lucide-react'
import { useNavigate } from 'react-router-dom'
import { PanelShell, type PanelLink } from '../components/ui/PanelShell'
import { useAuth } from '../context/AuthContext'

const links: PanelLink[] = [
  { to: '/', label: 'Dashboard', icon: LayoutDashboard, end: true },
  { to: '/admins', label: 'Adminlar', icon: Shield },
  { to: '/regions', label: 'Hududlar', icon: MapPin },
  { to: '/categories', label: 'Kategoriyalar', icon: Layers },
  { to: '/activity-types', label: 'Faoliyat turlari', icon: Briefcase },
  { to: '/managers', label: 'Menejerlar', icon: Users },
  { to: '/shop-directors', label: 'Savdo uyi rahbarlari', icon: Store },
  { to: '/kontragents', label: 'Kontragentlar', icon: Building2 },
  { to: '/products', label: 'Mahsulotlar', icon: Package },
  { to: '/shop-templates', label: 'Mahsulot shablonlari', icon: ClipboardList },
  { to: '/shop-products', label: 'Do‘kon mahsulotlari', icon: PackagePlus },
  { to: '/shop-incomings', label: 'Kirim', icon: ArrowDownToLine },
  { to: '/local-shops', label: 'Mahalla do‘konlari', icon: ShoppingBag },
  { to: '/sellers', label: 'Sotuvchilar', icon: UserRound },
  { to: '/deliveries', label: 'Yetkazuvchilar', icon: Truck },
  { to: '/service-providers', label: 'Xizmat ko‘rsatuvchilar', icon: Wrench },
  { to: '/identifications', label: 'Identifikatsiya', icon: IdCard },
  { to: '/services', label: 'Xizmatlar', icon: ListOrdered },
  { to: '/settings', label: 'Sozlamalar', icon: Settings },
]

export function AppLayout() {
  const { user, logout } = useAuth()
  const navigate = useNavigate()

  return (
    <PanelShell
      mark="A"
      accent="#122033"
      links={links}
      userName={`${user?.first_name ?? ''} ${user?.last_name ?? ''}`.trim()}
      userMeta={user?.role}
      onLogout={async () => {
        await logout()
        navigate('/login', { replace: true })
      }}
    />
  )
}
