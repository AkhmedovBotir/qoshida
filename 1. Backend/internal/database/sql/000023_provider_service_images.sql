-- Additive only. Does not drop tables or delete rows.

ALTER TABLE provider_services
    ADD COLUMN IF NOT EXISTS images TEXT[] NOT NULL DEFAULT '{}';
