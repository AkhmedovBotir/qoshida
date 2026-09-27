import { API_URL } from '../constants/config'
import type { CatalogItem, CatalogKind, OrderStatus } from '../types/market'

export function mediaSrc(url?: string | null) {
  if (!url) return ''
  if (url.startsWith('http://') || url.startsWith('https://') || url.startsWith('data:')) return url
  return `${API_URL}${url}`
}

export function formatMoney(value: number) {
  return `${new Intl.NumberFormat('uz-UZ').format(Math.round(value))} so‘m`
}

export function formatQty(value: number) {
  if (Number.isInteger(value)) return String(value)
  return String(Math.round(value * 1000) / 1000)
}

export function qtyStep(item?: Pick<CatalogItem, 'kind' | 'unit'> | null) {
  if (!item) return 1
  if (item.kind === 'service' || item.unit === 'dona' || item.unit === 'xizmat') return 1
  return 0.1
}

export function itemPath(kind: CatalogKind, id: string) {
  return `/item/${kind}/${id}`
}

export function kindLabel(kind: CatalogKind) {
  if (kind === 'shop') return 'Mahalla do‘koni'
  if (kind === 'service') return 'Xizmat'
  return 'Mahsulot'
}

export function kindMeta(kind: CatalogKind | '') {
  if (kind === 'shop') {
    return {
      title: 'Do‘konlar',
      hint: 'Mahalla do‘konlaridagi mahsulotlar. Yetkazib berish manzilingizga.',
      cta: 'Savatga qo‘shish',
      badge: 'Do‘kon',
    }
  }
  if (kind === 'service') {
    return {
      title: 'Xizmatlar',
      hint: 'Tasdiqlangan xizmat ko‘rsatuvchilar. Buyurtmadan keyin bog‘laniladi.',
      cta: 'Xizmatni buyurtma qilish',
      badge: 'Xizmat',
    }
  }
  if (kind === 'product') {
    return {
      title: 'Mahsulotlar',
      hint: 'Kontragentlardan tasdiqlangan tovarlar. Ombordan yetkaziladi.',
      cta: 'Savatga qo‘shish',
      badge: 'Mahsulot',
    }
  }
  return {
    title: 'Katalog',
    hint: 'Mahsulot, mahalla do‘koni va xizmatlar bir joyda.',
    cta: 'Savatga',
    badge: 'Barchasi',
  }
}

export function placeLine(item: Pick<CatalogItem, 'region_name' | 'district_name' | 'mfy_name'>) {
  return [item.region_name, item.district_name, item.mfy_name].filter(Boolean).join(' · ')
}

export function statusLabel(status: OrderStatus) {
  switch (status) {
    case 'confirmed':
      return 'Tasdiqlandi'
    case 'delivering':
      return 'Yetkazilmoqda'
    case 'done':
      return 'Yetkazildi'
    case 'cancelled':
      return 'Bekor qilindi'
    default:
      return 'Kutilmoqda'
  }
}

export function statusClass(status: OrderStatus) {
  switch (status) {
    case 'confirmed':
      return 'bg-sky-50 text-sky-800'
    case 'delivering':
      return 'bg-amber-50 text-amber-800'
    case 'done':
      return 'bg-emerald-50 text-emerald-800'
    case 'cancelled':
      return 'bg-red-50 text-red-700'
    default:
      return 'bg-slate-100 text-slate-700'
  }
}
