import { FormEvent, useEffect, useState } from 'react'
import { ClipboardList, Home, LayoutGrid, Search, ShoppingCart, UserRound } from 'lucide-react'
import { Link, NavLink, Outlet, useLocation, useNavigate } from 'react-router-dom'
import { useAuth } from '../context/AuthContext'
import { useCart } from '../context/CartContext'
import { cn } from '../lib/cn'

const tabs = [
  { to: '/', label: 'Bosh sahifa', icon: Home, end: true },
  { to: '/catalog', label: 'Katalog', icon: LayoutGrid },
  { to: '/cart', label: 'Savat', icon: ShoppingCart },
  { to: '/orders', label: 'Buyurtmalar', icon: ClipboardList, auth: true },
  { to: '/profile', label: 'Profil', icon: UserRound, auth: true },
] as const

export function MarketLayout() {
  const { cart } = useCart()
  const { user } = useAuth()
  const navigate = useNavigate()
  const location = useLocation()
  const [query, setQuery] = useState('')
  const hideMobileNav = location.pathname.startsWith('/item/')

  useEffect(() => {
    const root = document.documentElement
    const apply = () => {
      const mobile = window.matchMedia('(max-width: 1023px)').matches
      root.style.setProperty('--snack-offset', !mobile ? '1rem' : hideMobileNav ? '5.25rem' : '5.5rem')
    }
    apply()
    const mq = window.matchMedia('(max-width: 1023px)')
    mq.addEventListener('change', apply)
    return () => {
      mq.removeEventListener('change', apply)
      root.style.removeProperty('--snack-offset')
    }
  }, [hideMobileNav])

  function onSearch(event: FormEvent) {
    event.preventDefault()
    const q = query.trim()
    navigate(q ? `/catalog?q=${encodeURIComponent(q)}` : '/catalog')
  }

  function goTab(to: string, needsAuth?: boolean) {
    if (needsAuth && !user) {
      navigate(`/login?from=${encodeURIComponent(to)}`)
      return
    }
    navigate(to)
  }

  return (
    <div className="min-h-svh bg-[#F2F5F8]">
      <header className="sticky top-0 z-30 border-b border-[#E6EDF4] bg-white/90 pt-[env(safe-area-inset-top)] backdrop-blur-md">
        <div className="mx-auto flex max-w-6xl items-center gap-3 px-4 py-3 lg:gap-6 lg:py-3.5">
          <button type="button" onClick={() => navigate('/')} className="shrink-0 text-left">
            <p className="text-[10px] font-bold tracking-[1.4px] text-[#2E5D90] uppercase">Marketplace</p>
            <h1 className="text-lg leading-5 font-extrabold text-slate-900">Qoshida</h1>
          </button>

          <form onSubmit={onSearch} className="relative hidden min-w-0 flex-1 lg:block">
            <Search size={18} className="pointer-events-none absolute top-1/2 left-3.5 -translate-y-1/2 text-slate-400" />
            <input
              value={query}
              onChange={(e) => setQuery(e.target.value)}
              placeholder="Mahsulot, do‘kon yoki xizmat qidiring"
              className="w-full rounded-2xl border border-[#E6EDF4] bg-[#F2F5F8] py-2.5 pr-4 pl-11 text-sm outline-none focus:border-[#2E5D90] focus:bg-white"
            />
          </form>

          <nav className="ml-auto hidden items-center gap-1 lg:flex">
            {tabs.map((tab) => {
              const Icon = tab.icon
              return (
                <NavLink
                  key={tab.to}
                  to={tab.to}
                  end={'end' in tab ? tab.end : false}
                  onClick={(event) => {
                    if ('auth' in tab && tab.auth && !user) {
                      event.preventDefault()
                      goTab(tab.to, true)
                    }
                  }}
                  className={({ isActive }) =>
                    cn(
                      'relative flex items-center gap-2 rounded-xl px-3 py-2 text-sm font-semibold',
                      isActive ? 'bg-[#E8F0F8] text-[#2E5D90]' : 'text-slate-600 hover:bg-slate-50',
                    )
                  }
                >
                  <span className="relative">
                    <Icon size={18} />
                    {tab.to === '/cart' && cart.count > 0 ? (
                      <span className="absolute -top-1.5 -right-2 min-w-4 rounded-full bg-[#C2410C] px-1 text-[10px] leading-4 text-white">
                        {cart.count}
                      </span>
                    ) : null}
                  </span>
                  {tab.label}
                </NavLink>
              )
            })}
          </nav>

          {user ? (
            <Link to="/profile" className="hidden max-w-36 truncate text-sm font-semibold text-slate-700 lg:block">
              {user.name}
            </Link>
          ) : (
            <Link
              to="/login"
              className="hidden rounded-xl bg-[#2E5D90] px-3.5 py-2 text-sm font-semibold text-white lg:inline-flex"
            >
              Kirish
            </Link>
          )}

          <div className="ml-auto flex items-center gap-3 lg:hidden">
            <Link
              to="/cart"
              className="relative flex h-10 w-10 items-center justify-center rounded-xl bg-[#F2F5F8] text-slate-700"
              aria-label="Savat"
            >
              <ShoppingCart size={18} />
              {cart.count > 0 ? (
                <span className="absolute -top-1 -right-1 min-w-4 rounded-full bg-[#C2410C] px-1 text-[10px] leading-4 text-white">
                  {cart.count}
                </span>
              ) : null}
            </Link>
            <p className="max-w-24 truncate text-right text-xs font-medium text-slate-500">{user ? user.name : 'Mehmon'}</p>
          </div>
        </div>
      </header>

      <main
        className={cn(
          'mx-auto w-full max-w-6xl px-4 pt-4',
          hideMobileNav ? 'pb-28 lg:pb-10' : 'pb-[calc(5.75rem+env(safe-area-inset-bottom))] lg:pb-10',
        )}
      >
        <Outlet />
      </main>

      <nav
        className={cn(
          'fixed right-0 bottom-0 left-0 z-40 border-t border-[#E6EDF4] bg-white/95 pb-[env(safe-area-inset-bottom)] backdrop-blur-md lg:hidden',
          hideMobileNav && 'hidden',
        )}
      >
        <div className="mx-auto grid max-w-6xl grid-cols-5">
          {tabs.map((tab) => {
            const Icon = tab.icon
            return (
              <NavLink
                key={tab.to}
                to={tab.to}
                end={'end' in tab ? tab.end : false}
                onClick={(event) => {
                  if ('auth' in tab && tab.auth && !user) {
                    event.preventDefault()
                    goTab(tab.to, true)
                  }
                }}
                className={({ isActive }) =>
                  cn(
                    'relative flex flex-col items-center gap-1 py-2.5 text-[11px] font-semibold',
                    isActive ? 'text-[#2E5D90]' : 'text-slate-500',
                  )
                }
              >
                <span className="relative">
                  <Icon size={20} />
                  {tab.to === '/cart' && cart.count > 0 ? (
                    <span className="absolute -top-1.5 -right-2 min-w-4 rounded-full bg-[#2E5D90] px-1 text-[10px] leading-4 text-white">
                      {cart.count}
                    </span>
                  ) : null}
                </span>
                {tab.label}
              </NavLink>
            )
          })}
        </div>
      </nav>
    </div>
  )
}
