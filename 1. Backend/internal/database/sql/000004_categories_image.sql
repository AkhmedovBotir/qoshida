-- Additive only. Does not drop tables or delete rows.

ALTER TABLE categories
    ADD COLUMN IF NOT EXISTS image_url TEXT;
