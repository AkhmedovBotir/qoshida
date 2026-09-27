-- Additive only. Does not drop tables or delete rows.

ALTER TABLE products ADD COLUMN IF NOT EXISTS reviewed_by_name TEXT NOT NULL DEFAULT '';
ALTER TABLE products ADD COLUMN IF NOT EXISTS reviewed_by_role TEXT NOT NULL DEFAULT '';
ALTER TABLE products ADD COLUMN IF NOT EXISTS reviewed_at TIMESTAMPTZ;
