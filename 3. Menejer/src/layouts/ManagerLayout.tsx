import { ArrowDownToLine, Building2, ClipboardList, IdCard, LayoutDashboard, Layers, ListOrdered, Package, PackagePlus, ShoppingBag, Store, Truck, UserRound, Users, Wrench } from 'lucide-react'
import { useNavigate } from 'react-router-dom'
import { PanelShell, type PanelLink } from '../components/ui/PanelShell'
import { useAuth } from '../context/AuthContext'

export function ManagerLayout() {
  const { user, logout } = useAuth()
  const navigate = useNavigate()
  const isRegion = user?.type === 'region'
  const base = isRegion ? '/region' : '/district'

  const links: PanelLink[] = [
    { to: base, label: 'Dashboard', icon: LayoutDashboard, end: true },
    ...(isRegion ? [{ to: `${base}/district-managers`, label: 'Tuman menejerlari', icon: Users }] : []),
    { to: `${base}/shop-directors`, label: 'Savdo uyi rahbarlari', icon: Store },
    { to: `${base}/kontragents`, label: 'Kontragentlar', icon: Building2 },
    { to: `${base}/products`, label: 'Mahsulotlar', icon: Package },
    { to: `${base}/shop-templates`, label: 'Mahsulot shablonlari', icon: ClipboardList },
    { to: `${base}/shop-products`, label: 'Do‘kon mahsulotlari', icon: PackagePlus },
    { to: `${base}/shop-incomings`, label: 'Kirim', icon: ArrowDownToLine },
    { to: `${base}/local-shops`, label: 'Mahalla do‘konlari', icon: ShoppingBag },
    { to: `${base}/sellers`, label: 'Sotuvchilar', icon: UserRound },
    { to: `${base}/deliveries`, label: 'Yetkazuvchilar', icon: Truck },
    { to: `${base}/categories`, label: 'Kategoriyalar', icon: Layers },
    { to: `${base}/service-providers`, label: 'Xizmat ko‘rsatuvchilar', icon: Wrench },
    { to: `${base}/identifications`, label: 'Identifikatsiya', icon: IdCard },
    { to: `${base}/services`, label: 'Xizmatlar', icon: ListOrdered },
  ]

  return (
    <PanelShell
      brand={isRegion ? 'Viloyat menejeri' : 'Tuman menejeri'}
      mark="M"
      accent="#92400E"
      links={links}
      userName={`${user?.first_name ?? ''} ${user?.last_name ?? ''}`.trim()}
      userMeta={user?.region_name}
      onLogout={async () => {
        await logout()
        navigate('/login', { replace: true })
      }}
    />
  )
}
