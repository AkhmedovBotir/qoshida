import { MessageSquareText } from 'lucide-react'
import { motion } from 'motion/react'
import { useMemo } from 'react'
import { useSearchParams } from 'react-router-dom'
import { CommentTemplatesTab } from '../components/settings/CommentTemplatesTab'
import { APP_COLOR } from '../constants/config'
import { cn } from '../lib/cn'

const tabs = [{ id: 'comments', label: 'Kommentariya shablonlari', icon: MessageSquareText }] as const

type SettingsTab = (typeof tabs)[number]['id']

export function SettingsPage() {
  const [searchParams, setSearchParams] = useSearchParams()
  const activeTab = useMemo<SettingsTab>(() => {
    const raw = searchParams.get('tab')
    return tabs.some((tab) => tab.id === raw) ? (raw as SettingsTab) : 'comments'
  }, [searchParams])

  function setTab(id: SettingsTab) {
    setSearchParams({ tab: id }, { replace: true })
  }

  return (
    <motion.div initial={{ opacity: 0 }} animate={{ opacity: 1 }} className="mx-auto max-w-6xl space-y-6">
      <div>
        <h2 className="text-2xl font-bold text-slate-900">Sozlamalar</h2>
        <p className="mt-1 text-sm text-slate-500">Platforma sozlamalari va tayyor shablonlar</p>
      </div>

      <div className="flex flex-wrap gap-2 rounded-2xl border border-slate-200 bg-white p-2">
        {tabs.map((tab) => {
          const active = activeTab === tab.id
          return (
            <button
              key={tab.id}
              type="button"
              onClick={() => setTab(tab.id)}
              className={cn(
                'inline-flex items-center gap-2 rounded-xl px-4 py-2 text-sm font-semibold transition-colors',
                active ? 'text-white' : 'text-slate-600 hover:bg-slate-100',
              )}
              style={active ? { backgroundColor: APP_COLOR } : undefined}
            >
              <tab.icon size={16} />
              {tab.label}
            </button>
          )
        })}
      </div>

      {activeTab === 'comments' ? <CommentTemplatesTab /> : null}
    </motion.div>
  )
}
