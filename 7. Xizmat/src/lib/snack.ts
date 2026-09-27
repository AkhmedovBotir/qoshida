export type SnackKind = 'success' | 'error' | 'info'

export type SnackItem = {
  id: number
  kind: SnackKind
  message: string
}

type Listener = (items: SnackItem[]) => void

let nextId = 1
const items: SnackItem[] = []
const listeners = new Set<Listener>()

function emit() {
  const snapshot = items.slice()
  listeners.forEach((fn) => fn(snapshot))
}

export function subscribeSnacks(fn: Listener) {
  listeners.add(fn)
  fn(items.slice())
  return () => {
    listeners.delete(fn)
  }
}

export function pushSnack(kind: SnackKind, message: string) {
  const text = message.trim()
  if (!text) return
  const item: SnackItem = { id: nextId++, kind, message: text }
  items.push(item)
  if (items.length > 4) items.shift()
  emit()
  window.setTimeout(() => dismissSnack(item.id), kind === 'error' ? 5200 : 3600)
}

export function dismissSnack(id: number) {
  const index = items.findIndex((row) => row.id === id)
  if (index < 0) return
  items.splice(index, 1)
  emit()
}

export const toast = {
  success: (message: string) => pushSnack('success', message),
  error: (message: string) => pushSnack('error', message),
  info: (message: string) => pushSnack('info', message),
}

export function notifyApiError(path: string, err: { message: string }, silent?: boolean) {
  if (silent) return
  if (/\/(me|refresh)(\?|$)/.test(path)) return
  toast.error(err.message || 'So‘rov bajarilmadi')
}

export function notifyApiSuccess(path: string, method = 'GET', silent?: boolean) {
  if (silent) return
  const verb = method.toUpperCase()
  if (verb === 'GET') return
  if (/\/(me|refresh|login|logout|check-phone|send-code|resend-code|verify-code|register|set-password|cart)(\?|$)|\/import(\?|$)/.test(path)) return
  if (verb === 'DELETE') {
    toast.success('O‘chirildi')
    return
  }
  if (/\/approve(\?|$)/.test(path)) {
    toast.success('Tasdiqlandi')
    return
  }
  if (/\/reject(\?|$)/.test(path)) {
    toast.success('Bekor qilindi')
    return
  }
  if (/\/checkout(\?|$)/.test(path)) {
    toast.success('Buyurtma qabul qilindi')
    return
  }
  toast.success('Saqlandi')
}
