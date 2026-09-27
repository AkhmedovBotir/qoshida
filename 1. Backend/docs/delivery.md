# Delivery API Documentation

Yetkazuvchi (delivery courier) module. Same scoped CRUD pattern as sellers. Cookies: `qoshida_delivery_access`, `qoshida_delivery_refresh`. JWT `use` is `delivery`.

---

## POST /api/v1/delivery-auth/check-phone

Description: Checks whether a delivery phone exists, is active, and needs first-time password setup.

Authentication: Public (rate limited)

### Headers

| Header | Type | Required | Description |
|--------|------|----------|-------------|
| Content-Type | String | Yes | application/json |

### Request Body (JSON)

{
  "phone": "+998901234567"
}

### Responses

#### 200 Success

{
  "success": true,
  "data": {
    "exists": true,
    "active": true,
    "has_password": false,
    "needs_setup": true,
    "message": ""
  }
}

#### 400 / 429 Error Examples

{
  "success": false,
  "error": "Telefon raqami noto'g'ri"
}

---

## POST /api/v1/delivery-auth/send-code

Description: Sends SMS OTP for password setup. Also used by `/resend-code`.

Authentication: Public (rate limited)

### Request Body (JSON)

{
  "phone": "+998901234567"
}

### Responses

#### 200 Success

{
  "success": true,
  "data": {
    "status": "sent",
    "message": "SMS kodi yuborildi. Kod 5 daqiqa amal qiladi"
  }
}

---

## POST /api/v1/delivery-auth/verify-code

Description: Verifies the 6-digit SMS code before password setup.

Authentication: Public (rate limited)

### Request Body (JSON)

{
  "phone": "+998901234567",
  "code": "123456"
}

### Responses

#### 200 Success

{
  "success": true,
  "data": { "status": "verified" }
}

---

## POST /api/v1/delivery-auth/set-password

Description: Sets the first password after OTP verification and issues HttpOnly cookies.

Authentication: Public (rate limited)

### Request Body (JSON)

{
  "phone": "+998901234567",
  "password": "secret12"
}

### Responses

#### 200 Success

{
  "success": true,
  "data": {
    "delivery": {
      "id": "7c8a1c2e-1111-4d33-9a11-aaaaaaaaaaaa",
      "shop_id": "2b1a0c9d-2222-4e44-8b22-bbbbbbbbbbbb",
      "first_name": "Aziz",
      "last_name": "Karimov",
      "phone": "+998901234567",
      "status": "active",
      "has_password": true
    }
  }
}

---

## POST /api/v1/delivery-auth/login

Description: Password login for an existing delivery account. Sets access and refresh cookies.

Authentication: Public (rate limited)

### Request Body (JSON)

{
  "phone": "+998901234567",
  "password": "secret12"
}

### Responses

#### 200 Success

{
  "success": true,
  "data": {
    "delivery": {
      "id": "7c8a1c2e-1111-4d33-9a11-aaaaaaaaaaaa",
      "first_name": "Aziz",
      "last_name": "Karimov",
      "phone": "+998901234567",
      "status": "active",
      "has_password": true,
      "shop_name": "Chilonzor do'koni"
    }
  }
}

#### 401 Error Examples

{
  "success": false,
  "error": "Telefon yoki parol noto'g'ri"
}

---

## POST /api/v1/delivery-auth/refresh

Description: Rotates delivery access token from the refresh cookie.

Authentication: Refresh cookie required

### Request Body (JSON)

{}

---

## POST /api/v1/delivery-auth/logout

Description: Clears delivery cookies and revokes the refresh token.

Authentication: Cookie session

### Request Body (JSON)

{}

---

## GET /api/v1/delivery-auth/me

Description: Returns the current delivery actor from cookies.

Authentication: Delivery cookies

---

## GET /api/v1/delivery/profile

Description: Returns the authenticated delivery public profile.

Authentication: Required (delivery JWT cookies)

---

## GET /api/v1/deliveries

Description: Admin list of deliveries. Supports `q`, `page`, `limit`, `shop_id`, `region_id`, `district_id`, `mfy_id`. Each item includes `audit` (creator + last distinct updaters).

Authentication: Required (admin)

### Request Params / Query Parameters

| Parameter | Type | Required | Description |
|-----------|------|----------|-------------|
| q | String | No | Search by name or phone |
| page | Number | No | Page number, default 1 |
| limit | Number | No | Page size, default 20 |
| shop_id | UUID | No | Filter by shop |

### Responses

#### 200 Success

{
  "success": true,
  "data": {
    "items": [],
    "total": 0,
    "page": 1,
    "limit": 20
  }
}

---

## POST /api/v1/deliveries

Description: Admin creates a delivery. Password is optional; if empty an SMS setup code is sent. Create is automatically signed in `audit.created`.

Authentication: Required (admin)

### Request Body (JSON)

{
  "shop_id": "2b1a0c9d-2222-4e44-8b22-bbbbbbbbbbbb",
  "first_name": "Aziz",
  "last_name": "Karimov",
  "phone": "+998901234567",
  "status": "active",
  "password": ""
}

---

## GET /api/v1/deliveries/stats

Description: Admin totals for the dashboard.

Authentication: Required (admin)

---

## GET /api/v1/deliveries/{id}

Description: Admin get one delivery, including audit trail.

Authentication: Required (admin)

---

## PUT /api/v1/deliveries/{id}

Description: Admin update. Stamps the current actor into `audit.updates` (last 3 distinct people).

Authentication: Required (admin)

---

## DELETE /api/v1/deliveries/{id}

Description: Admin delete.

Authentication: Required (admin)

---

## Scoped CRUD

Same create/update/delete body as admin. Scope is forced by the caller:

| Caller | Auth | Base path |
|--------|------|-----------|
| Manager | manager cookies | `/api/v1/manager/deliveries` |
| Shop director | director cookies | `/api/v1/shop-director/deliveries` |
| Local shop | shop cookies | `/api/v1/local-shop/deliveries` (shop_id ignored, own shop is used) |

Local shop create does not accept another shop. Manager/director/admin pick a shop in their geo scope.
