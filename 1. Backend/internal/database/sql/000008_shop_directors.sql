-- Additive only. Does not drop tables or delete rows.

CREATE TABLE IF NOT EXISTS shop_directors (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    region_id UUID NOT NULL REFERENCES regions(id) ON DELETE RESTRICT,
    district_id UUID NOT NULL REFERENCES regions(id) ON DELETE RESTRICT,
    mfy_id UUID NOT NULL REFERENCES regions(id) ON DELETE RESTRICT,
    first_name TEXT NOT NULL CHECK (char_length(first_name) BETWEEN 2 AND 80),
    last_name TEXT NOT NULL CHECK (char_length(last_name) BETWEEN 2 AND 80),
    phone TEXT NOT NULL UNIQUE CHECK (phone ~ '^\+998[0-9]{9}$'),
    password_hash TEXT NOT NULL DEFAULT '',
    status TEXT NOT NULL CHECK (status IN ('active', 'inactive')) DEFAULT 'active',
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE INDEX IF NOT EXISTS idx_shop_directors_region ON shop_directors(region_id);
CREATE INDEX IF NOT EXISTS idx_shop_directors_district ON shop_directors(district_id);
CREATE INDEX IF NOT EXISTS idx_shop_directors_mfy ON shop_directors(mfy_id);
CREATE INDEX IF NOT EXISTS idx_shop_directors_status ON shop_directors(status);

CREATE TABLE IF NOT EXISTS shop_director_verification_codes (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    shop_director_id UUID NOT NULL REFERENCES shop_directors(id) ON DELETE CASCADE,
    phone TEXT NOT NULL,
    code_hash TEXT NOT NULL,
    purpose TEXT NOT NULL DEFAULT 'password_setup',
    expires_at TIMESTAMPTZ NOT NULL,
    attempts INTEGER NOT NULL DEFAULT 0,
    verified_at TIMESTAMPTZ,
    used_at TIMESTAMPTZ,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE INDEX IF NOT EXISTS idx_shop_director_codes_phone ON shop_director_verification_codes(phone, purpose);
CREATE INDEX IF NOT EXISTS idx_shop_director_codes_director ON shop_director_verification_codes(shop_director_id);

CREATE TABLE IF NOT EXISTS shop_director_refresh_tokens (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    shop_director_id UUID NOT NULL REFERENCES shop_directors(id) ON DELETE CASCADE,
    token_hash TEXT NOT NULL UNIQUE,
    expires_at TIMESTAMPTZ NOT NULL,
    revoked_at TIMESTAMPTZ,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE INDEX IF NOT EXISTS idx_shop_director_refresh_director ON shop_director_refresh_tokens(shop_director_id);
