-- Additive only. Does not drop tables or delete rows.

ALTER TABLE customers
    ADD COLUMN IF NOT EXISTS avatar TEXT NOT NULL DEFAULT '';
