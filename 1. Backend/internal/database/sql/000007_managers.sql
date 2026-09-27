-- Additive only. Does not drop tables or delete rows.

CREATE TABLE IF NOT EXISTS managers (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    type TEXT NOT NULL CHECK (type IN ('region', 'district')),
    region_id UUID NOT NULL REFERENCES regions(id) ON DELETE RESTRICT,
    district_id UUID REFERENCES regions(id) ON DELETE RESTRICT,
    first_name TEXT NOT NULL CHECK (char_length(first_name) BETWEEN 2 AND 80),
    last_name TEXT NOT NULL CHECK (char_length(last_name) BETWEEN 2 AND 80),
    phone TEXT NOT NULL UNIQUE CHECK (phone ~ '^\+998[0-9]{9}$'),
    username TEXT NOT NULL UNIQUE CHECK (username ~ '^[a-zA-Z0-9._]{3,32}$'),
    password_hash TEXT NOT NULL DEFAULT '',
    status TEXT NOT NULL CHECK (status IN ('active', 'inactive')) DEFAULT 'active',
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    CONSTRAINT managers_district_required CHECK (
        (type = 'region' AND district_id IS NULL) OR
        (type = 'district' AND district_id IS NOT NULL)
    )
);

CREATE INDEX IF NOT EXISTS idx_managers_type ON managers(type);
CREATE INDEX IF NOT EXISTS idx_managers_region ON managers(region_id);
CREATE INDEX IF NOT EXISTS idx_managers_district ON managers(district_id);
CREATE INDEX IF NOT EXISTS idx_managers_status ON managers(status);

CREATE TABLE IF NOT EXISTS manager_verification_codes (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    manager_id UUID NOT NULL REFERENCES managers(id) ON DELETE CASCADE,
    phone TEXT NOT NULL,
    code_hash TEXT NOT NULL,
    purpose TEXT NOT NULL DEFAULT 'password_setup',
    expires_at TIMESTAMPTZ NOT NULL,
    attempts INTEGER NOT NULL DEFAULT 0,
    verified_at TIMESTAMPTZ,
    used_at TIMESTAMPTZ,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE INDEX IF NOT EXISTS idx_manager_codes_phone ON manager_verification_codes(phone, purpose);
CREATE INDEX IF NOT EXISTS idx_manager_codes_manager ON manager_verification_codes(manager_id);

CREATE TABLE IF NOT EXISTS manager_refresh_tokens (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    manager_id UUID NOT NULL REFERENCES managers(id) ON DELETE CASCADE,
    token_hash TEXT NOT NULL UNIQUE,
    expires_at TIMESTAMPTZ NOT NULL,
    revoked_at TIMESTAMPTZ,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE INDEX IF NOT EXISTS idx_manager_refresh_manager ON manager_refresh_tokens(manager_id);
