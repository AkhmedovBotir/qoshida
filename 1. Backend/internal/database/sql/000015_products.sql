-- Additive only. Does not drop tables or delete rows.

CREATE TABLE IF NOT EXISTS products (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    kontragent_id UUID NOT NULL REFERENCES kontragents(id) ON DELETE CASCADE,
    category_id UUID NOT NULL REFERENCES categories(id),
    subcategory_id UUID NOT NULL REFERENCES categories(id),
    name TEXT NOT NULL,
    description TEXT NOT NULL DEFAULT '',
    sale_price NUMERIC(14,2) NOT NULL CHECK (sale_price >= 0),
    cost_price NUMERIC(14,2) NOT NULL CHECK (cost_price >= 0),
    quantity NUMERIC(14,3) NOT NULL CHECK (quantity >= 0),
    unit TEXT NOT NULL CHECK (unit IN ('dona', 'litr', 'kg')),
    unit_size NUMERIC(14,3) NOT NULL CHECK (unit_size > 0),
    commission_percent NUMERIC(5,2) NOT NULL CHECK (commission_percent >= 0 AND commission_percent <= 100),
    images TEXT[] NOT NULL DEFAULT '{}',
    approval_status TEXT NOT NULL DEFAULT 'pending' CHECK (approval_status IN ('pending', 'approved', 'rejected')),
    rejection_note TEXT NOT NULL DEFAULT '',
    submitted_by TEXT NOT NULL CHECK (submitted_by IN ('kontragent', 'staff')),
    status TEXT NOT NULL DEFAULT 'active' CHECK (status IN ('active', 'inactive')),
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE INDEX IF NOT EXISTS idx_products_kontragent ON products(kontragent_id);
CREATE INDEX IF NOT EXISTS idx_products_approval ON products(approval_status);
CREATE INDEX IF NOT EXISTS idx_products_category ON products(category_id, subcategory_id);
CREATE INDEX IF NOT EXISTS idx_products_name ON products(name);
