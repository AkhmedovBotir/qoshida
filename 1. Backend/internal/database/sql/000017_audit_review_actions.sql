-- Additive only. Does not drop tables or delete rows.

ALTER TABLE record_audits DROP CONSTRAINT IF EXISTS record_audits_action_check;
ALTER TABLE record_audits ADD CONSTRAINT record_audits_action_check
    CHECK (action IN ('create', 'update', 'approve', 'reject'));
