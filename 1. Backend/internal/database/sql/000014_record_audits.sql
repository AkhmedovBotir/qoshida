-- Additive only. Does not drop tables or delete rows.

CREATE TABLE IF NOT EXISTS record_audits (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    entity_type TEXT NOT NULL,
    entity_id UUID NOT NULL,
    action TEXT NOT NULL CHECK (action IN ('create', 'update')),
    actor_type TEXT NOT NULL,
    actor_id UUID NOT NULL,
    actor_name TEXT NOT NULL,
    actor_role TEXT NOT NULL,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE INDEX IF NOT EXISTS idx_record_audits_entity ON record_audits(entity_type, entity_id, created_at DESC);
