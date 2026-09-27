-- Additive only. Does not drop tables or delete rows.

CREATE TABLE IF NOT EXISTS sellers (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    shop_id UUID NOT NULL REFERENCES local_shops(id) ON DELETE CASCADE,
    first_name TEXT NOT NULL CHECK (char_length(first_name) BETWEEN 2 AND 80),
    last_name TEXT NOT NULL CHECK (char_length(last_name) BETWEEN 2 AND 80),
    phone TEXT NOT NULL UNIQUE CHECK (phone ~ '^\+998[0-9]{9}$'),
    password_hash TEXT NOT NULL DEFAULT '',
    status TEXT NOT NULL CHECK (status IN ('active', 'inactive')) DEFAULT 'active',
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE INDEX IF NOT EXISTS idx_sellers_shop ON sellers(shop_id);
CREATE INDEX IF NOT EXISTS idx_sellers_status ON sellers(status);

CREATE TABLE IF NOT EXISTS seller_verification_codes (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    seller_id UUID NOT NULL REFERENCES sellers(id) ON DELETE CASCADE,
    phone TEXT NOT NULL,
    code_hash TEXT NOT NULL,
    purpose TEXT NOT NULL DEFAULT 'password_setup',
    expires_at TIMESTAMPTZ NOT NULL,
    attempts INTEGER NOT NULL DEFAULT 0,
    verified_at TIMESTAMPTZ,
    used_at TIMESTAMPTZ,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE INDEX IF NOT EXISTS idx_seller_codes_phone ON seller_verification_codes(phone, purpose);
CREATE INDEX IF NOT EXISTS idx_seller_codes_seller ON seller_verification_codes(seller_id);

CREATE TABLE IF NOT EXISTS seller_refresh_tokens (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    seller_id UUID NOT NULL REFERENCES sellers(id) ON DELETE CASCADE,
    token_hash TEXT NOT NULL UNIQUE,
    expires_at TIMESTAMPTZ NOT NULL,
    revoked_at TIMESTAMPTZ,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE INDEX IF NOT EXISTS idx_seller_refresh_seller ON seller_refresh_tokens(seller_id);
