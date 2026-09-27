-- Additive only. Does not drop tables or delete rows.

CREATE TABLE IF NOT EXISTS shop_product_templates (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    category_id UUID NOT NULL REFERENCES categories(id),
    subcategory_id UUID NOT NULL REFERENCES categories(id),
    name TEXT NOT NULL,
    description TEXT NOT NULL DEFAULT '',
    unit TEXT NOT NULL CHECK (unit IN ('dona', 'litr', 'kg')),
    unit_size NUMERIC(14,3) NOT NULL CHECK (unit_size > 0),
    images TEXT[] NOT NULL DEFAULT '{}',
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE INDEX IF NOT EXISTS idx_shop_product_templates_name ON shop_product_templates(name);
CREATE INDEX IF NOT EXISTS idx_shop_product_templates_category ON shop_product_templates(category_id, subcategory_id);

CREATE TABLE IF NOT EXISTS shop_products (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    shop_id UUID NOT NULL REFERENCES local_shops(id) ON DELETE CASCADE,
    template_id UUID NOT NULL REFERENCES shop_product_templates(id) ON DELETE RESTRICT,
    quantity NUMERIC(14,3) NOT NULL CHECK (quantity >= 0),
    sale_price NUMERIC(14,2) NOT NULL CHECK (sale_price >= 0),
    cost_price NUMERIC(14,2) NOT NULL CHECK (cost_price >= 0),
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    UNIQUE (shop_id, template_id)
);

CREATE INDEX IF NOT EXISTS idx_shop_products_shop ON shop_products(shop_id);
CREATE INDEX IF NOT EXISTS idx_shop_products_template ON shop_products(template_id);

CREATE TABLE IF NOT EXISTS shop_product_incomings (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    shop_product_id UUID NOT NULL REFERENCES shop_products(id) ON DELETE CASCADE,
    quantity NUMERIC(14,3) NOT NULL CHECK (quantity > 0),
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE INDEX IF NOT EXISTS idx_shop_product_incomings_product ON shop_product_incomings(shop_product_id);
CREATE INDEX IF NOT EXISTS idx_shop_product_incomings_created ON shop_product_incomings(created_at DESC);
