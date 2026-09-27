-- Additive only. Does not drop tables or delete rows.

CREATE TABLE IF NOT EXISTS comment_templates (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    comment TEXT NOT NULL CHECK (char_length(comment) BETWEEN 1 AND 4000),
    sort_order INTEGER NOT NULL DEFAULT 0,
    status TEXT NOT NULL CHECK (status IN ('active', 'inactive')) DEFAULT 'active',
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE INDEX IF NOT EXISTS idx_comment_templates_sort ON comment_templates(sort_order, id);
CREATE INDEX IF NOT EXISTS idx_comment_templates_status ON comment_templates(status);
