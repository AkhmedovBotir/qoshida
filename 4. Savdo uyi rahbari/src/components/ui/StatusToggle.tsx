import { APP_COLOR } from '../../constants/config'

type StatusToggleProps = {
  active: boolean
  onChange: () => void
  size?: 'sm' | 'md'
}

const sizes = {
  md: { track: 'h-6 w-11', thumb: 'h-4 w-4', on: 'translate-x-6', off: 'translate-x-1' },
  sm: { track: 'h-5 w-9', thumb: 'h-3.5 w-3.5', on: 'translate-x-5', off: 'translate-x-0.5' },
}

export function StatusToggle({ active, onChange, size = 'md' }: StatusToggleProps) {
  const s = sizes[size]

  return (
    <button
      type="button"
      role="switch"
      aria-checked={active}
      onClick={onChange}
      className={`relative inline-flex items-center rounded-full transition-colors focus:outline-none focus:ring-2 focus:ring-offset-2 ${s.track} ${
        active ? '' : 'bg-slate-300'
      }`}
      style={active ? { backgroundColor: APP_COLOR } : undefined}
    >
      <span
        className={`inline-block transform rounded-full bg-white transition-transform ${s.thumb} ${
          active ? s.on : s.off
        }`}
      />
    </button>
  )
}
