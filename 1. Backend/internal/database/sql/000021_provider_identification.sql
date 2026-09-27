-- Additive only. Does not drop tables or delete rows.

CREATE TABLE IF NOT EXISTS service_provider_identifications (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    provider_id UUID NOT NULL UNIQUE REFERENCES service_providers(id) ON DELETE CASCADE,
    full_name TEXT NOT NULL CHECK (char_length(full_name) BETWEEN 5 AND 160),
    pinfl TEXT NOT NULL CHECK (pinfl ~ '^[0-9]{14}$'),
    passport TEXT NOT NULL CHECK (passport ~ '^[A-Z]{2}[0-9]{7}$'),
    birth_date DATE NOT NULL,
    address TEXT NOT NULL CHECK (char_length(address) BETWEEN 10 AND 300),
    experience_years INTEGER NOT NULL DEFAULT 0 CHECK (experience_years BETWEEN 0 AND 70),
    about TEXT NOT NULL DEFAULT '',
    passport_image TEXT NOT NULL DEFAULT '',
    selfie_image TEXT NOT NULL DEFAULT '',
    documents JSONB NOT NULL DEFAULT '[]'::jsonb,
    status TEXT NOT NULL DEFAULT 'pending' CHECK (status IN ('pending', 'approved', 'rejected')),
    rejection_note TEXT NOT NULL DEFAULT '',
    reviewed_by_name TEXT NOT NULL DEFAULT '',
    reviewed_by_role TEXT NOT NULL DEFAULT '',
    reviewed_at TIMESTAMPTZ,
    submitted_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE UNIQUE INDEX IF NOT EXISTS idx_provider_ident_pinfl ON service_provider_identifications(pinfl);
CREATE INDEX IF NOT EXISTS idx_provider_ident_status ON service_provider_identifications(status);
CREATE INDEX IF NOT EXISTS idx_provider_ident_submitted ON service_provider_identifications(submitted_at DESC);
