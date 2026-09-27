-- Additive only. Does not drop tables or delete rows.

CREATE TABLE IF NOT EXISTS service_providers (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    activity_type_id UUID NOT NULL REFERENCES activity_types(id) ON DELETE RESTRICT,
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

CREATE INDEX IF NOT EXISTS idx_service_providers_activity_type ON service_providers(activity_type_id);
CREATE INDEX IF NOT EXISTS idx_service_providers_region ON service_providers(region_id);
CREATE INDEX IF NOT EXISTS idx_service_providers_district ON service_providers(district_id);
CREATE INDEX IF NOT EXISTS idx_service_providers_mfy ON service_providers(mfy_id);
CREATE INDEX IF NOT EXISTS idx_service_providers_status ON service_providers(status);

CREATE TABLE IF NOT EXISTS service_provider_verification_codes (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    provider_id UUID NOT NULL REFERENCES service_providers(id) ON DELETE CASCADE,
    phone TEXT NOT NULL,
    code_hash TEXT NOT NULL,
    purpose TEXT NOT NULL DEFAULT 'password_setup',
    expires_at TIMESTAMPTZ NOT NULL,
    attempts INTEGER NOT NULL DEFAULT 0,
    verified_at TIMESTAMPTZ,
    used_at TIMESTAMPTZ,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE INDEX IF NOT EXISTS idx_service_provider_codes_phone ON service_provider_verification_codes(phone, purpose);
CREATE INDEX IF NOT EXISTS idx_service_provider_codes_provider ON service_provider_verification_codes(provider_id);

CREATE TABLE IF NOT EXISTS service_provider_refresh_tokens (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    provider_id UUID NOT NULL REFERENCES service_providers(id) ON DELETE CASCADE,
    token_hash TEXT NOT NULL UNIQUE,
    expires_at TIMESTAMPTZ NOT NULL,
    revoked_at TIMESTAMPTZ,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE INDEX IF NOT EXISTS idx_service_provider_refresh_provider ON service_provider_refresh_tokens(provider_id);
