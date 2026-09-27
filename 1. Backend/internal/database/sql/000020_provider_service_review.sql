-- Additive only. Does not drop tables or delete rows.

ALTER TABLE provider_services ADD COLUMN IF NOT EXISTS approval_status TEXT NOT NULL DEFAULT 'approved';
ALTER TABLE provider_services ADD COLUMN IF NOT EXISTS rejection_note TEXT NOT NULL DEFAULT '';
ALTER TABLE provider_services ADD COLUMN IF NOT EXISTS submitted_by TEXT NOT NULL DEFAULT 'staff';
ALTER TABLE provider_services ADD COLUMN IF NOT EXISTS reviewed_by_name TEXT NOT NULL DEFAULT '';
ALTER TABLE provider_services ADD COLUMN IF NOT EXISTS reviewed_by_role TEXT NOT NULL DEFAULT '';
ALTER TABLE provider_services ADD COLUMN IF NOT EXISTS reviewed_at TIMESTAMPTZ;

DO $$
BEGIN
    IF NOT EXISTS (
        SELECT 1 FROM pg_constraint WHERE conname = 'provider_services_approval_status_check'
    ) THEN
        ALTER TABLE provider_services
            ADD CONSTRAINT provider_services_approval_status_check
            CHECK (approval_status IN ('pending', 'approved', 'rejected'));
    END IF;
    IF NOT EXISTS (
        SELECT 1 FROM pg_constraint WHERE conname = 'provider_services_submitted_by_check'
    ) THEN
        ALTER TABLE provider_services
            ADD CONSTRAINT provider_services_submitted_by_check
            CHECK (submitted_by IN ('provider', 'staff'));
    END IF;
END $$;

CREATE INDEX IF NOT EXISTS idx_provider_services_approval ON provider_services(approval_status);
