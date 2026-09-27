import { Capacitor } from '@capacitor/core'

export type AppPlatform = 'web' | 'android' | 'ios'

export function usePlatform(): AppPlatform {
  const native = Capacitor.getPlatform()
  if (native === 'android') return 'android'
  if (native === 'ios') return 'ios'
  return 'web'
}
