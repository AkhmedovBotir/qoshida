# Record audit API Documentation

Audit is not a standalone HTTP resource. Create and update handlers stamp the current actor into `record_audits`. List/get responses attach:

```json
"audit": {
  "created": { "name": "Ali Valiyev", "role": "Admin", "at": "2026-09-14T00:00:00Z" },
  "updates": [
    { "name": "Nodira", "role": "Viloyat menejeri", "at": "2026-09-14T01:00:00Z" },
    { "name": "Jasur", "role": "Savdo uyi rahbari", "at": "2026-09-14T02:00:00Z" }
  ]
}
```

`updates` keeps the last 3 distinct people. The same person updating again replaces their own row.

Covered entities: manager, shop_director, kontragent, local_shop, service_provider, seller, delivery, category.

Actor roles: `Admin`, `Viloyat menejeri`, `Tuman menejeri`, `Savdo uyi rahbari`, `Mahalla do‘koni`.
