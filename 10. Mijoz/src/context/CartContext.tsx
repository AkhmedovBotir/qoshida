import { createContext, useCallback, useContext, useEffect, useMemo, useState, type ReactNode } from 'react'
import { getCart, upsertCart } from '../lib/market'
import type { CartLine, CartPublic, CatalogItem, CatalogKind } from '../types/market'
import { useAuth } from './AuthContext'

const LOCAL_KEY = 'qoshida_mijoz_cart'

type CartContextValue = {
  cart: CartPublic
  ready: boolean
  refresh: () => Promise<void>
  setQuantity: (kind: CatalogKind, itemId: string, quantity: number, item?: CatalogItem) => Promise<void>
  add: (item: CatalogItem, quantity?: number) => Promise<void>
}

const CartContext = createContext<CartContextValue | null>(null)

function emptyCart(): CartPublic {
  return { items: [], total: 0, count: 0 }
}

function roundMoney(value: number) {
  return Math.round(value * 100) / 100
}

function rebuild(items: CartLine[]): CartPublic {
  const next = items
    .filter((row) => row.quantity > 0)
    .map((row) => ({
      ...row,
      line_total: roundMoney((row.item?.price ?? 0) * row.quantity),
    }))
  return {
    items: next,
    total: roundMoney(next.reduce((sum, row) => sum + row.line_total, 0)),
    count: next.length,
  }
}

function readLocal(): CartPublic {
  try {
    const raw = localStorage.getItem(LOCAL_KEY)
    if (!raw) return emptyCart()
    const parsed = JSON.parse(raw) as CartPublic
    return rebuild(parsed.items ?? [])
  } catch {
    return emptyCart()
  }
}

function writeLocal(cart: CartPublic) {
  localStorage.setItem(LOCAL_KEY, JSON.stringify(rebuild(cart.items)))
}

export function CartProvider({ children }: { children: ReactNode }) {
  const { user, loading } = useAuth()
  const [cart, setCart] = useState<CartPublic>(emptyCart)
  const [ready, setReady] = useState(false)

  const refreshServer = useCallback(async () => {
    const data = await getCart()
    setCart(data)
    return data
  }, [])

  useEffect(() => {
    if (loading) return
    let cancelled = false
    ;(async () => {
      try {
        if (user) {
          const local = readLocal()
          for (const line of local.items) {
            try {
              await upsertCart({ kind: line.kind, item_id: line.item_id, quantity: line.quantity }, { silent: true })
            } catch {
              /* mahsulot o‘chgan bo‘lishi mumkin */
            }
          }
          writeLocal(emptyCart())
          const data = await getCart()
          if (!cancelled) setCart(data)
        } else if (!cancelled) {
          setCart(readLocal())
        }
      } catch {
        if (!cancelled) setCart(user ? emptyCart() : readLocal())
      } finally {
        if (!cancelled) setReady(true)
      }
    })()
    return () => {
      cancelled = true
    }
  }, [user, loading])

  const refresh = useCallback(async () => {
    if (user) {
      await refreshServer()
      return
    }
    setCart(readLocal())
  }, [user, refreshServer])

  const setQuantity = useCallback(
    async (kind: CatalogKind, itemId: string, quantity: number, item?: CatalogItem) => {
      if (user) {
        const data = await upsertCart({ kind, item_id: itemId, quantity })
        setCart(data)
        return
      }
      const current = readLocal()
      const nextItems = current.items.filter((row) => !(row.kind === kind && row.item_id === itemId))
      if (quantity > 0) {
        const existing = current.items.find((row) => row.kind === kind && row.item_id === itemId)
        nextItems.push({
          id: `${kind}:${itemId}`,
          kind,
          item_id: itemId,
          quantity,
          item: item ?? existing?.item,
          line_total: 0,
        })
      }
      const next = rebuild(nextItems)
      writeLocal(next)
      setCart(next)
    },
    [user],
  )

  const add = useCallback(
    async (item: CatalogItem, quantity = 1) => {
      const existing = cart.items.find((row) => row.kind === item.kind && row.item_id === item.id)
      const nextQty = (existing?.quantity ?? 0) + quantity
      await setQuantity(item.kind, item.id, nextQty, item)
    },
    [cart.items, setQuantity],
  )

  const value = useMemo<CartContextValue>(
    () => ({ cart, ready, refresh, setQuantity, add }),
    [cart, ready, refresh, setQuantity, add],
  )

  return <CartContext.Provider value={value}>{children}</CartContext.Provider>
}

export function useCart() {
  const ctx = useContext(CartContext)
  if (!ctx) throw new Error('useCart CartProvider ichida ishlatilishi kerak')
  return ctx
}
