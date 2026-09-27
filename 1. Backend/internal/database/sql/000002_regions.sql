-- Additive only. Does not drop tables or delete rows.

CREATE TABLE IF NOT EXISTS regions (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    source_id TEXT NOT NULL UNIQUE,
    name TEXT NOT NULL CHECK (char_length(name) BETWEEN 2 AND 160),
    type TEXT NOT NULL CHECK (type IN ('region', 'district', 'mfy')),
    parent_id UUID REFERENCES regions(id) ON DELETE RESTRICT,
    parent_source_id TEXT,
    code TEXT NOT NULL CHECK (char_length(code) BETWEEN 1 AND 80),
    status TEXT NOT NULL CHECK (status IN ('active', 'inactive')) DEFAULT 'active',
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE INDEX IF NOT EXISTS idx_regions_type ON regions(type);
CREATE INDEX IF NOT EXISTS idx_regions_parent ON regions(parent_id);
CREATE INDEX IF NOT EXISTS idx_regions_name ON regions(name);
CREATE INDEX IF NOT EXISTS idx_regions_status ON regions(status);
CREATE INDEX IF NOT EXISTS idx_regions_parent_source ON regions(parent_source_id);
