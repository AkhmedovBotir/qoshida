-- Additive only. Does not drop tables or delete rows.

ALTER TABLE customers
    ADD COLUMN IF NOT EXISTS first_name TEXT NOT NULL DEFAULT '',
    ADD COLUMN IF NOT EXISTS last_name TEXT NOT NULL DEFAULT '',
    ADD COLUMN IF NOT EXISTS birth_date DATE;

UPDATE customers
SET
    first_name = CASE
        WHEN btrim(first_name) <> '' THEN first_name
        ELSE split_part(btrim(name), ' ', 1)
    END,
    last_name = CASE
        WHEN btrim(last_name) <> '' THEN last_name
        WHEN position(' ' IN btrim(name)) > 0 THEN btrim(substring(btrim(name) FROM position(' ' IN btrim(name)) + 1))
        ELSE ''
    END
WHERE btrim(first_name) = '' OR btrim(last_name) = '';

ALTER TABLE customer_verification_codes
    DROP CONSTRAINT IF EXISTS customer_verification_codes_purpose_check;

ALTER TABLE customer_verification_codes
    ADD CONSTRAINT customer_verification_codes_purpose_check
    CHECK (purpose IN ('register', 'password_setup', 'password_reset', 'login'));
