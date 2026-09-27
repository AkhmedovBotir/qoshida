export type LocalShopItem = {
  id: string
  name: string
}

export function listLocalShops() {
  return Promise.resolve({ items: [] as LocalShopItem[], total: 0 })
}
