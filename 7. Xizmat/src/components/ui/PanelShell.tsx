import { LogOut, Menu, X, type LucideIcon } from 'lucide-react'
import { useEffect, useState } from 'react'
import { NavLink, Outlet, useLocation } from 'react-router-dom'
import { APP_COLOR, APP_NAME } from '../../constants/config'
import { cn } from '../../lib/cn'

export type PanelLink = {
  to: string
  label: string
  icon: LucideIcon
  end?: boolean
}

type PanelShellProps = {
  brand?: string
  mark: string
  accent: string
  links: PanelLink[]
  userName?: string
  userMeta?: string
  onLogout: () => void | Promise<void>
}

export function PanelShell({ brand = APP_NAME, mark, accent, links, userName, userMeta, onLogout }: PanelShellProps) {
  const location = useLocation()
  const [expanded, setExpanded] = useState(true)
  const [canvasOpen, setCanvasOpen] = useState(false)

  useEffect(() => {
    setCanvasOpen(false)
  }, [location.pathname])

  useEffect(() => {
    if (!canvasOpen) return
    const prev = document.body.style.overflow
    document.body.style.overflow = 'hidden'
    return () => {
      document.body.style.overflow = prev
    }
  }, [canvasOpen])

  function toggleMenu() {
    if (window.matchMedia('(min-width: 1024px)').matches) {
      setExpanded((v) => !v)
      return
    }
    setCanvasOpen((v) => !v)
  }

  return (
    <div className="flex min-h-svh bg-[#F4F6F9]">
      {canvasOpen ? (
        <button
          type="button"
          aria-label="Menyuni yopish"
          className="fixed inset-0 z-30 bg-slate-950/45 lg:hidden"
          onClick={() => setCanvasOpen(false)}
        />
      ) : null}

      <aside
        className={cn(
          'flex h-svh flex-col text-white transition-[transform,width] duration-200',
          'fixed inset-y-0 left-0 z-40 w-72',
          canvasOpen ? 'translate-x-0' : '-translate-x-full',
          'lg:sticky lg:z-auto lg:translate-x-0',
          expanded ? 'lg:w-72' : 'lg:w-[72px]',
        )}
        style={{ backgroundColor: accent }}
      >
        <div className={cn('flex items-center py-5', expanded ? 'justify-between px-5 lg:justify-between' : 'px-5 lg:justify-center lg:px-2')}>
          <div className={cn('min-w-0', !expanded && 'lg:hidden')}>
            <p className="text-[11px] font-bold uppercase tracking-[1.4px] text-white/60">Qoshida</p>
            <p className="truncate text-lg font-semibold">{brand}</p>
          </div>
          <span className={cn('hidden h-10 w-10 items-center justify-center rounded-xl bg-white/10 text-sm font-bold lg:flex', expanded && 'lg:hidden')}>
            {mark}
          </span>
          <button type="button" className="rounded-lg p-1 text-white/80 lg:hidden" onClick={() => setCanvasOpen(false)} aria-label="Yopish">
            <X size={18} />
          </button>
        </div>
        <nav className={cn('flex flex-1 flex-col gap-1 overflow-y-auto pb-4', expanded ? 'px-3' : 'px-3 lg:px-2')}>
          {links.map((link) => (
            <NavLink
              key={link.to}
              to={link.to}
              end={link.end}
              title={link.label}
              className={({ isActive }) =>
                cn(
                  'flex items-center rounded-xl py-2.5 text-sm font-medium text-white/75 hover:bg-white/10',
                  expanded ? 'gap-3 px-3' : 'gap-3 px-3 lg:justify-center lg:px-0',
                  isActive && 'bg-white/15 text-white',
                )
              }
            >
              <link.icon size={18} className="shrink-0" />
              <span className={cn(!expanded && 'lg:sr-only')}>{link.label}</span>
            </NavLink>
          ))}
        </nav>
      </aside>

      <div className="flex min-w-0 flex-1 flex-col">
        <header className="sticky top-0 z-20 flex items-center justify-between gap-2 border-b border-slate-200 bg-white px-3 py-3 sm:px-4">
          <div className="flex min-w-0 items-center gap-2 sm:gap-3">
            <button
              type="button"
              className="rounded-lg border border-slate-200 p-2"
              onClick={toggleMenu}
              aria-label={canvasOpen || expanded ? 'Menyuni yigish' : 'Menyuni ochish'}
            >
              <Menu size={18} />
            </button>
            <h1 className="truncate text-sm font-semibold text-slate-900 sm:text-base">{APP_NAME} paneli</h1>
          </div>
          <div className="flex shrink-0 items-center gap-2 sm:gap-3">
            <div className="hidden text-right sm:block">
              <p className="text-sm font-semibold text-slate-900">{userName}</p>
              {userMeta ? <p className="text-xs text-slate-500">{userMeta}</p> : null}
            </div>
            <button
              type="button"
              onClick={() => {
                void onLogout()
              }}
              className="inline-flex items-center gap-2 rounded-xl px-3 py-2 text-sm font-semibold text-white"
              style={{ backgroundColor: APP_COLOR }}
            >
              <LogOut size={16} />
              <span className="hidden sm:inline">Chiqish</span>
            </button>
          </div>
        </header>
        <div className="min-w-0 flex-1 overflow-x-hidden p-3 sm:p-4 md:p-6">
          <Outlet />
        </div>
      </div>
    </div>
  )
}
