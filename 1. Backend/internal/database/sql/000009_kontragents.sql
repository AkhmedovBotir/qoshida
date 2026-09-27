-- Additive only. Does not drop tables or delete rows.

CREATE TABLE IF NOT EXISTS kontragents (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    activity_type_id UUID NOT NULL REFERENCES activity_types(id) ON DELETE RESTRICT,
    name TEXT NOT NULL CHECK (char_length(name) BETWEEN 2 AND 160),
    inn TEXT NOT NULL UNIQUE CHECK (inn ~ '^[0-9]{9}$'),
    region_id UUID NOT NULL REFERENCES regions(id) ON DELETE RESTRICT,
    district_id UUID NOT NULL REFERENCES regions(id) ON DELETE RESTRICT,
    mfy_id UUID NOT NULL REFERENCES regions(id) ON DELETE RESTRICT,
    phone TEXT NOT NULL UNIQUE CHECK (phone ~ '^\+998[0-9]{9}$'),
    logo TEXT NOT NULL DEFAULT '',
    password_hash TEXT NOT NULL DEFAULT '',
    status TEXT NOT NULL CHECK (status IN ('active', 'inactive')) DEFAULT 'active',
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE INDEX IF NOT EXISTS idx_kontragents_activity_type ON kontragents(activity_type_id);
CREATE INDEX IF NOT EXISTS idx_kontragents_region ON kontragents(region_id);
CREATE INDEX IF NOT EXISTS idx_kontragents_district ON kontragents(district_id);
CREATE INDEX IF NOT EXISTS idx_kontragents_mfy ON kontragents(mfy_id);
CREATE INDEX IF NOT EXISTS idx_kontragents_status ON kontragents(status);

CREATE TABLE IF NOT EXISTS kontragent_verification_codes (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    kontragent_id UUID NOT NULL REFERENCES kontragents(id) ON DELETE CASCADE,
    phone TEXT NOT NULL,
    code_hash TEXT NOT NULL,
    purpose TEXT NOT NULL DEFAULT 'password_setup',
    expires_at TIMESTAMPTZ NOT NULL,
    attempts INTEGER NOT NULL DEFAULT 0,
    verified_at TIMESTAMPTZ,
    used_at TIMESTAMPTZ,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE INDEX IF NOT EXISTS idx_kontragent_codes_phone ON kontragent_verification_codes(phone, purpose);
CREATE INDEX IF NOT EXISTS idx_kontragent_codes_kontragent ON kontragent_verification_codes(kontragent_id);

CREATE TABLE IF NOT EXISTS kontragent_refresh_tokens (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    kontragent_id UUID NOT NULL REFERENCES kontragents(id) ON DELETE CASCADE,
    token_hash TEXT NOT NULL UNIQUE,
    expires_at TIMESTAMPTZ NOT NULL,
    revoked_at TIMESTAMPTZ,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE INDEX IF NOT EXISTS idx_kontragent_refresh_kontragent ON kontragent_refresh_tokens(kontragent_id);
