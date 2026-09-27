# Category API Documentation

Categories are global (no geo scope). JSON import stays Admin-only. Manager and shop director can list/create/update/delete.

---

## GET /api/v1/categories

Description: Admin paginated category list. Query: `page`, `limit`, `parent_id`, `roots`, `q`. Items include `audit`.

Authentication: Required (admin)

### Responses

#### 200 Success

{
  "success": true,
  "data": {
    "items": [
      {
        "id": "11111111-1111-1111-1111-111111111111",
        "name": "Oziq-ovqat",
        "slug": "oziq-ovqat",
        "parent_id": null,
        "censored": false,
        "status": "active",
        "audit": {
          "created": { "name": "Ali Valiyev", "role": "Admin", "at": "2026-09-14T00:00:00Z" },
          "updates": []
        }
      }
    ],
    "total": 1,
    "page": 1,
    "limit": 100
  }
}

---

## POST /api/v1/categories

Description: Admin creates a root or child category. Automatically signed by the current admin.

Authentication: Required (admin)

### Request Body (JSON)

{
  "name": "Oziq-ovqat",
  "slug": "oziq-ovqat",
  "parent_id": null,
  "censored": false,
  "status": "active"
}

---

## PUT /api/v1/categories/{id}

Description: Admin updates a category and stamps `audit.updates`.

Authentication: Required (admin)

---

## DELETE /api/v1/categories/{id}

Description: Admin deletes a category that has no children.

Authentication: Required (admin)

---

## POST /api/v1/categories/import

Description: Imports `scripts/ttsa.categories.json`. Admin only. Not mounted for manager or shop director.

Authentication: Required (admin)

### Request Body (JSON)

{}

---

## GET /api/v1/manager/categories

Description: Same list/create/get/update/delete as admin, without import. Auth: viloyat or tuman menejeri.

Authentication: Required (manager cookies)

Same paths under `/api/v1/shop-director/categories` for savdo uyi rahbari.
