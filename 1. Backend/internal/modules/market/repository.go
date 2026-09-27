package market

import (
	"context"
	"errors"
	"fmt"
	"strings"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

var ErrNotFound = errors.New("catalog item not found")

type Repository struct {
	pool *pgxpool.Pool
}

func NewRepository(pool *pgxpool.Pool) *Repository {
	return &Repository{pool: pool}
}

const catalogSelect = `
SELECT kind, id, name, description, price, quantity, unit, unit_size, images, seller_name, seller_id,
       category_id, category_name, region_id, district_id, mfy_id, region_name, district_name, mfy_name
FROM (
	SELECT 'product'::text AS kind, p.id, p.name, p.description, p.sale_price::float8 AS price, p.quantity::float8,
		p.unit, p.unit_size::float8, p.images, k.name AS seller_name, k.id AS seller_id,
		p.category_id, COALESCE(c.name, '') AS category_name,
		k.region_id, k.district_id, k.mfy_id,
		COALESCE(reg.name, '') AS region_name, COALESCE(dis.name, '') AS district_name, COALESCE(mfy.name, '') AS mfy_name
	FROM products p
	JOIN kontragents k ON k.id = p.kontragent_id
	LEFT JOIN categories c ON c.id = p.category_id
	LEFT JOIN regions reg ON reg.id = k.region_id
	LEFT JOIN regions dis ON dis.id = k.district_id
	LEFT JOIN regions mfy ON mfy.id = k.mfy_id
	WHERE p.approval_status = 'approved' AND p.status = 'active' AND p.quantity > 0 AND k.status = 'active'
		AND COALESCE(c.censored, false) = false
	UNION ALL
	SELECT 'shop', sp.id, t.name, t.description, sp.sale_price::float8, sp.quantity::float8,
		t.unit, t.unit_size::float8,
		CASE WHEN cardinality(t.images) > 0 THEN t.images WHEN s.image <> '' THEN ARRAY[s.image] ELSE ARRAY[]::text[] END,
		s.name, s.id,
		t.category_id, COALESCE(c.name, ''),
		s.region_id, s.district_id, s.mfy_id,
		COALESCE(reg.name, ''), COALESCE(dis.name, ''), COALESCE(mfy.name, '')
	FROM shop_products sp
	JOIN shop_product_templates t ON t.id = sp.template_id
	JOIN local_shops s ON s.id = sp.shop_id
	LEFT JOIN categories c ON c.id = t.category_id
	LEFT JOIN regions reg ON reg.id = s.region_id
	LEFT JOIN regions dis ON dis.id = s.district_id
	LEFT JOIN regions mfy ON mfy.id = s.mfy_id
	WHERE sp.quantity > 0 AND s.status = 'active' AND COALESCE(c.censored, false) = false
	UNION ALL
	SELECT 'service', sv.id, sv.name, '', sv.price::float8, 1,
		'xizmat', 1,
		CASE
			WHEN cardinality(COALESCE(sv.images, '{}')) > 0 THEN sv.images
			WHEN pr.image <> '' THEN ARRAY[pr.image]
			ELSE ARRAY[]::text[]
		END,
		pr.name, pr.id,
		pr.activity_type_id, COALESCE(at.name, ''),
		pr.region_id, pr.district_id, pr.mfy_id,
		COALESCE(reg.name, ''), COALESCE(dis.name, ''), COALESCE(mfy.name, '')
	FROM provider_services sv
	JOIN service_providers pr ON pr.id = sv.provider_id
	LEFT JOIN activity_types at ON at.id = pr.activity_type_id
	LEFT JOIN service_provider_identifications ident ON ident.provider_id = pr.id
	LEFT JOIN regions reg ON reg.id = pr.region_id
	LEFT JOIN regions dis ON dis.id = pr.district_id
	LEFT JOIN regions mfy ON mfy.id = pr.mfy_id
	WHERE sv.approval_status = 'approved' AND pr.status = 'active' AND ident.status = 'approved'
) catalog
`

func scanCatalog(row pgx.Row) (CatalogItem, error) {
	var item CatalogItem
	err := row.Scan(
		&item.Kind, &item.ID, &item.Name, &item.Description, &item.Price, &item.Quantity, &item.Unit, &item.UnitSize,
		&item.Images, &item.SellerName, &item.SellerID, &item.CategoryID, &item.CategoryName,
		&item.RegionID, &item.DistrictID, &item.MFYID, &item.RegionName, &item.DistrictName, &item.MFYName,
	)
	if errors.Is(err, pgx.ErrNoRows) {
		return CatalogItem{}, ErrNotFound
	}
	if item.Images == nil {
		item.Images = []string{}
	}
	return item, err
}

func (r *Repository) ListCatalog(ctx context.Context, q ListQuery) ([]CatalogItem, int, error) {
	where := []string{"TRUE"}
	args := make([]any, 0, 8)
	if q.Kind == KindProduct || q.Kind == KindShop || q.Kind == KindService {
		args = append(args, q.Kind)
		where = append(where, fmt.Sprintf("kind = $%d", len(args)))
	}
	if q.CategoryID != nil {
		args = append(args, *q.CategoryID)
		where = append(where, fmt.Sprintf("category_id = $%d", len(args)))
	}
	if q.RegionID != nil {
		args = append(args, *q.RegionID)
		where = append(where, fmt.Sprintf("region_id = $%d", len(args)))
	}
	if q.DistrictID != nil {
		args = append(args, *q.DistrictID)
		where = append(where, fmt.Sprintf("district_id = $%d", len(args)))
	}
	if q.MFYID != nil {
		args = append(args, *q.MFYID)
		where = append(where, fmt.Sprintf("mfy_id = $%d", len(args)))
	}
	if q.Query != "" {
		args = append(args, "%"+q.Query+"%")
		where = append(where, fmt.Sprintf("(name ILIKE $%d OR seller_name ILIKE $%d OR category_name ILIKE $%d)", len(args), len(args), len(args)))
	}
	clause := strings.Join(where, " AND ")

	var total int
	if err := r.pool.QueryRow(ctx, `SELECT COUNT(*) FROM (`+catalogSelect+`) x WHERE `+clause, args...).Scan(&total); err != nil {
		return nil, 0, err
	}
	args = append(args, q.Limit, (q.Page-1)*q.Limit)
	rows, err := r.pool.Query(ctx, catalogSelect+`
		WHERE `+clause+`
		ORDER BY name
		LIMIT $`+fmt.Sprintf("%d", len(args)-1)+` OFFSET $`+fmt.Sprintf("%d", len(args))+`
	`, args...)
	if err != nil {
		return nil, 0, err
	}
	defer rows.Close()
	items := make([]CatalogItem, 0, q.Limit)
	for rows.Next() {
		item, err := scanCatalog(rows)
		if err != nil {
			return nil, 0, err
		}
		items = append(items, item)
	}
	return items, total, rows.Err()
}

func (r *Repository) GetCatalog(ctx context.Context, kind string, id uuid.UUID) (CatalogItem, error) {
	return scanCatalog(r.pool.QueryRow(ctx, catalogSelect+` WHERE kind = $1 AND id = $2`, kind, id))
}

func (r *Repository) ListCategories(ctx context.Context, parent *uuid.UUID, roots bool, limit int) ([]CategoryPublic, error) {
	where := []string{"c.status = 'active'", "c.censored = false"}
	args := make([]any, 0, 3)
	if roots {
		where = append(where, "c.parent_id IS NULL")
	}
	if parent != nil {
		args = append(args, *parent)
		where = append(where, fmt.Sprintf("c.parent_id = $%d", len(args)))
	}
	args = append(args, limit)
	rows, err := r.pool.Query(ctx, `
		SELECT c.id, c.name, c.slug, COALESCE(c.image_url, ''), COALESCE(c.parent_id::text, '')
		FROM categories c
		WHERE `+strings.Join(where, " AND ")+`
		ORDER BY c.name
		LIMIT $`+fmt.Sprintf("%d", len(args))+`
	`, args...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	items := make([]CategoryPublic, 0, limit)
	for rows.Next() {
		var item CategoryPublic
		if err = rows.Scan(&item.ID, &item.Name, &item.Slug, &item.ImageURL, &item.ParentID); err != nil {
			return nil, err
		}
		items = append(items, item)
	}
	return items, rows.Err()
}

func (r *Repository) ListCart(ctx context.Context, customerID uuid.UUID) ([]CartRow, error) {
	rows, err := r.pool.Query(ctx, `
		SELECT id, kind, item_id, quantity::float8, created_at
		FROM cart_items WHERE customer_id = $1 ORDER BY created_at
	`, customerID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	items := make([]CartRow, 0)
	for rows.Next() {
		var row CartRow
		if err = rows.Scan(&row.ID, &row.Kind, &row.ItemID, &row.Quantity, &row.CreatedAt); err != nil {
			return nil, err
		}
		items = append(items, row)
	}
	return items, rows.Err()
}

func (r *Repository) UpsertCart(ctx context.Context, customerID uuid.UUID, kind string, itemID uuid.UUID, qty float64) error {
	if qty <= 0 {
		_, err := r.pool.Exec(ctx, `DELETE FROM cart_items WHERE customer_id = $1 AND kind = $2 AND item_id = $3`, customerID, kind, itemID)
		return err
	}
	_, err := r.pool.Exec(ctx, `
		INSERT INTO cart_items (customer_id, kind, item_id, quantity)
		VALUES ($1,$2,$3,$4)
		ON CONFLICT (customer_id, kind, item_id)
		DO UPDATE SET quantity = EXCLUDED.quantity, updated_at = NOW()
	`, customerID, kind, itemID, qty)
	return err
}

func (r *Repository) ClearCart(ctx context.Context, customerID uuid.UUID) error {
	_, err := r.pool.Exec(ctx, `DELETE FROM cart_items WHERE customer_id = $1`, customerID)
	return err
}

func (r *Repository) CreateOrder(ctx context.Context, order Order) (Order, error) {
	tx, err := r.pool.Begin(ctx)
	if err != nil {
		return Order{}, err
	}
	defer tx.Rollback(ctx)

	if _, err = tx.Exec(ctx, `
		INSERT INTO orders (id, customer_id, status, total, address, note, region_id, district_id, mfy_id)
		VALUES ($1,$2,$3,$4,$5,$6,$7,$8,$9)
	`, order.ID, order.CustomerID, order.Status, order.Total, order.Address, order.Note, nilUUID(order.RegionID), nilUUID(order.DistrictID), nilUUID(order.MFYID)); err != nil {
		return Order{}, err
	}
	for _, row := range order.Items {
		if _, err = tx.Exec(ctx, `
			INSERT INTO order_items (order_id, kind, item_id, name, unit, quantity, unit_price, seller_name, image)
			VALUES ($1,$2,$3,$4,$5,$6,$7,$8,$9)
		`, order.ID, row.Kind, row.ItemID, row.Name, row.Unit, row.Quantity, row.UnitPrice, row.SellerName, row.Image); err != nil {
			return Order{}, err
		}
		switch row.Kind {
		case KindProduct:
			tag, err := tx.Exec(ctx, `UPDATE products SET quantity = quantity - $2, updated_at = NOW() WHERE id = $1 AND quantity >= $2`, row.ItemID, row.Quantity)
			if err != nil {
				return Order{}, err
			}
			if tag.RowsAffected() == 0 {
				return Order{}, errInsufficient
			}
		case KindShop:
			tag, err := tx.Exec(ctx, `UPDATE shop_products SET quantity = quantity - $2, updated_at = NOW() WHERE id = $1 AND quantity >= $2`, row.ItemID, row.Quantity)
			if err != nil {
				return Order{}, err
			}
			if tag.RowsAffected() == 0 {
				return Order{}, errInsufficient
			}
		}
	}
	if err = tx.Commit(ctx); err != nil {
		return Order{}, err
	}
	return r.FindOrder(ctx, order.ID, order.CustomerID)
}

var errInsufficient = errors.New("insufficient stock")

func (r *Repository) FindOrder(ctx context.Context, id, customerID uuid.UUID) (Order, error) {
	var item Order
	err := r.pool.QueryRow(ctx, `
		SELECT id, customer_id, status, total::float8, address, note, created_at
		FROM orders WHERE id = $1 AND customer_id = $2
	`, id, customerID).Scan(&item.ID, &item.CustomerID, &item.Status, &item.Total, &item.Address, &item.Note, &item.CreatedAt)
	if errors.Is(err, pgx.ErrNoRows) {
		return Order{}, ErrNotFound
	}
	if err != nil {
		return Order{}, err
	}
	rows, err := r.pool.Query(ctx, `
		SELECT kind, item_id, name, unit, quantity::float8, unit_price::float8, seller_name, image
		FROM order_items WHERE order_id = $1
	`, id)
	if err != nil {
		return Order{}, err
	}
	defer rows.Close()
	for rows.Next() {
		var row OrderItem
		if err = rows.Scan(&row.Kind, &row.ItemID, &row.Name, &row.Unit, &row.Quantity, &row.UnitPrice, &row.SellerName, &row.Image); err != nil {
			return Order{}, err
		}
		item.Items = append(item.Items, row)
	}
	return item, rows.Err()
}

func (r *Repository) ListOrders(ctx context.Context, customerID uuid.UUID, page, limit int) ([]Order, int, error) {
	var total int
	if err := r.pool.QueryRow(ctx, `SELECT COUNT(*) FROM orders WHERE customer_id = $1`, customerID).Scan(&total); err != nil {
		return nil, 0, err
	}
	rows, err := r.pool.Query(ctx, `
		SELECT id FROM orders WHERE customer_id = $1 ORDER BY created_at DESC LIMIT $2 OFFSET $3
	`, customerID, limit, (page-1)*limit)
	if err != nil {
		return nil, 0, err
	}
	defer rows.Close()
	ids := make([]uuid.UUID, 0, limit)
	for rows.Next() {
		var id uuid.UUID
		if err = rows.Scan(&id); err != nil {
			return nil, 0, err
		}
		ids = append(ids, id)
	}
	items := make([]Order, 0, len(ids))
	for _, id := range ids {
		item, err := r.FindOrder(ctx, id, customerID)
		if err != nil {
			return nil, 0, err
		}
		items = append(items, item)
	}
	return items, total, nil
}

func nilUUID(id uuid.UUID) any {
	if id == uuid.Nil {
		return nil
	}
	return id
}
