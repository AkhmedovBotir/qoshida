-- Additive only. Does not drop tables or delete rows.

CREATE TABLE IF NOT EXISTS local_shops (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    name TEXT NOT NULL CHECK (char_length(name) BETWEEN 2 AND 160),
    phone TEXT NOT NULL UNIQUE CHECK (phone ~ '^\+998[0-9]{9}$'),
    image TEXT NOT NULL DEFAULT '',
    region_id UUID NOT NULL REFERENCES regions(id) ON DELETE RESTRICT,
    district_id UUID NOT NULL REFERENCES regions(id) ON DELETE RESTRICT,
    mfy_id UUID NOT NULL REFERENCES regions(id) ON DELETE RESTRICT,
    password_hash TEXT NOT NULL DEFAULT '',
    status TEXT NOT NULL CHECK (status IN ('active', 'inactive')) DEFAULT 'active',
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE INDEX IF NOT EXISTS idx_local_shops_region ON local_shops(region_id);
CREATE INDEX IF NOT EXISTS idx_local_shops_district ON local_shops(district_id);
CREATE INDEX IF NOT EXISTS idx_local_shops_mfy ON local_shops(mfy_id);
CREATE INDEX IF NOT EXISTS idx_local_shops_status ON local_shops(status);

CREATE TABLE IF NOT EXISTS local_shop_verification_codes (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    shop_id UUID NOT NULL REFERENCES local_shops(id) ON DELETE CASCADE,
    phone TEXT NOT NULL,
    code_hash TEXT NOT NULL,
    purpose TEXT NOT NULL DEFAULT 'password_setup',
    expires_at TIMESTAMPTZ NOT NULL,
    attempts INTEGER NOT NULL DEFAULT 0,
    verified_at TIMESTAMPTZ,
    used_at TIMESTAMPTZ,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE INDEX IF NOT EXISTS idx_local_shop_codes_phone ON local_shop_verification_codes(phone, purpose);
CREATE INDEX IF NOT EXISTS idx_local_shop_codes_shop ON local_shop_verification_codes(shop_id);

CREATE TABLE IF NOT EXISTS local_shop_refresh_tokens (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    shop_id UUID NOT NULL REFERENCES local_shops(id) ON DELETE CASCADE,
    token_hash TEXT NOT NULL UNIQUE,
    expires_at TIMESTAMPTZ NOT NULL,
    revoked_at TIMESTAMPTZ,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE INDEX IF NOT EXISTS idx_local_shop_refresh_shop ON local_shop_refresh_tokens(shop_id);
