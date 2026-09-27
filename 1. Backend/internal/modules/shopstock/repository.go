package shopstock

import (
	"context"
	"errors"
	"strings"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

var ErrNotFound = errors.New("shop product not found")

type Repository struct {
	pool *pgxpool.Pool
}

func NewRepository(pool *pgxpool.Pool) *Repository {
	return &Repository{pool: pool}
}

const productCols = `
	sp.id, sp.shop_id, sp.template_id, sp.quantity, sp.sale_price, sp.cost_price, sp.created_at, sp.updated_at,
	COALESCE(t.name, ''), COALESCE(t.description, ''), t.category_id, t.subcategory_id,
	COALESCE(c.name, ''), COALESCE(sc.name, ''), COALESCE(t.unit, ''), COALESCE(t.unit_size, 0),
	COALESCE(t.images, '{}'), COALESCE(s.name, ''),
	s.region_id, s.district_id, s.mfy_id,
	COALESCE(reg.name, ''), COALESCE(dis.name, ''), COALESCE(mfy.name, '')
`

const incomingCols = `
	i.id, i.shop_product_id, i.quantity, i.created_at,
	COALESCE(t.name, ''), s.id, COALESCE(s.name, ''), COALESCE(t.unit, ''), COALESCE(t.unit_size, 0),
	s.region_id, s.district_id, s.mfy_id
`

func scanProduct(row pgx.Row) (ShopProduct, error) {
	var item ShopProduct
	err := row.Scan(
		&item.ID, &item.ShopID, &item.TemplateID, &item.Quantity, &item.SalePrice, &item.CostPrice, &item.CreatedAt, &item.UpdatedAt,
		&item.Name, &item.Description, &item.CategoryID, &item.SubcategoryID,
		&item.CategoryName, &item.SubcategoryName, &item.Unit, &item.UnitSize,
		&item.Images, &item.ShopName,
		&item.RegionID, &item.DistrictID, &item.MFYID,
		&item.RegionName, &item.DistrictName, &item.MFYName,
	)
	if errors.Is(err, pgx.ErrNoRows) {
		return ShopProduct{}, ErrNotFound
	}
	return item, err
}

func scanIncoming(row pgx.Row) (Incoming, error) {
	var item Incoming
	err := row.Scan(
		&item.ID, &item.ShopProductID, &item.Quantity, &item.CreatedAt,
		&item.ProductName, &item.ShopID, &item.ShopName, &item.Unit, &item.UnitSize,
		&item.RegionID, &item.DistrictID, &item.MFYID,
	)
	if errors.Is(err, pgx.ErrNoRows) {
		return Incoming{}, ErrNotFound
	}
	return item, err
}

func productJoins() string {
	return `
		FROM shop_products sp
		JOIN local_shops s ON s.id = sp.shop_id
		JOIN shop_product_templates t ON t.id = sp.template_id
		LEFT JOIN categories c ON c.id = t.category_id
		LEFT JOIN categories sc ON sc.id = t.subcategory_id
		LEFT JOIN regions reg ON reg.id = s.region_id
		LEFT JOIN regions dis ON dis.id = s.district_id
		LEFT JOIN regions mfy ON mfy.id = s.mfy_id
	`
}

func incomingJoins() string {
	return `
		FROM shop_product_incomings i
		JOIN shop_products sp ON sp.id = i.shop_product_id
		JOIN local_shops s ON s.id = sp.shop_id
		JOIN shop_product_templates t ON t.id = sp.template_id
	`
}

func (r *Repository) CreateProduct(ctx context.Context, item ShopProduct) (ShopProduct, error) {
	if item.ID == uuid.Nil {
		item.ID = uuid.New()
	}
	err := r.pool.QueryRow(ctx, `
		INSERT INTO shop_products (id, shop_id, template_id, quantity, sale_price, cost_price)
		VALUES ($1,$2,$3,$4,$5,$6)
		RETURNING id
	`, item.ID, item.ShopID, item.TemplateID, item.Quantity, item.SalePrice, item.CostPrice).Scan(&item.ID)
	if err != nil {
		return ShopProduct{}, err
	}
	return r.FindProduct(ctx, item.ID)
}

func (r *Repository) UpdateProduct(ctx context.Context, item ShopProduct) (ShopProduct, error) {
	tag, err := r.pool.Exec(ctx, `
		UPDATE shop_products SET sale_price = $2, cost_price = $3, updated_at = NOW()
		WHERE id = $1
	`, item.ID, item.SalePrice, item.CostPrice)
	if err != nil {
		return ShopProduct{}, err
	}
	if tag.RowsAffected() == 0 {
		return ShopProduct{}, ErrNotFound
	}
	return r.FindProduct(ctx, item.ID)
}

func (r *Repository) DeleteProduct(ctx context.Context, id uuid.UUID) error {
	tag, err := r.pool.Exec(ctx, `DELETE FROM shop_products WHERE id = $1`, id)
	if err != nil {
		return err
	}
	if tag.RowsAffected() == 0 {
		return ErrNotFound
	}
	return nil
}

func (r *Repository) FindProduct(ctx context.Context, id uuid.UUID) (ShopProduct, error) {
	return scanProduct(r.pool.QueryRow(ctx, `SELECT `+productCols+productJoins()+` WHERE sp.id = $1`, id))
}

func (r *Repository) FindByShopAndTemplate(ctx context.Context, shopID, templateID uuid.UUID) (ShopProduct, error) {
	return scanProduct(r.pool.QueryRow(ctx, `SELECT `+productCols+productJoins()+` WHERE sp.shop_id = $1 AND sp.template_id = $2`, shopID, templateID))
}

func (r *Repository) ListProducts(ctx context.Context, q ProductListQuery) ([]ShopProduct, int, error) {
	where := []string{"1=1"}
	args := []any{}
	add := func(cond string, val any) {
		args = append(args, val)
		where = append(where, strings.Replace(cond, "?", "$"+itoa(len(args)), 1))
	}
	if q.ShopID != nil {
		add("sp.shop_id = ?", *q.ShopID)
	}
	if q.TemplateID != nil {
		add("sp.template_id = ?", *q.TemplateID)
	}
	if q.RegionID != nil {
		add("s.region_id = ?", *q.RegionID)
	}
	if q.DistrictID != nil {
		add("s.district_id = ?", *q.DistrictID)
	}
	if q.MFYID != nil {
		add("s.mfy_id = ?", *q.MFYID)
	}
	if q.Query != "" {
		args = append(args, "%"+q.Query+"%")
		n := itoa(len(args))
		where = append(where, "(t.name ILIKE $"+n+" OR s.name ILIKE $"+n+")")
	}

	clause := strings.Join(where, " AND ")
	var total int
	if err := r.pool.QueryRow(ctx, `SELECT COUNT(*) `+productJoins()+` WHERE `+clause, args...).Scan(&total); err != nil {
		return nil, 0, err
	}

	offset := (q.Page - 1) * q.Limit
	args = append(args, q.Limit, offset)
	limitPos := itoa(len(args) - 1)
	offsetPos := itoa(len(args))
	rows, err := r.pool.Query(ctx, `
		SELECT `+productCols+productJoins()+`
		WHERE `+clause+`
		ORDER BY sp.created_at DESC
		LIMIT $`+limitPos+` OFFSET $`+offsetPos, args...)
	if err != nil {
		return nil, 0, err
	}
	defer rows.Close()

	items := make([]ShopProduct, 0)
	for rows.Next() {
		item, err := scanProduct(rows)
		if err != nil {
			return nil, 0, err
		}
		items = append(items, item)
	}
	return items, total, rows.Err()
}

func (r *Repository) CreateIncoming(ctx context.Context, item Incoming) (Incoming, error) {
	if item.ID == uuid.Nil {
		item.ID = uuid.New()
	}
	tx, err := r.pool.Begin(ctx)
	if err != nil {
		return Incoming{}, err
	}
	defer tx.Rollback(ctx)

	if err = tx.QueryRow(ctx, `
		INSERT INTO shop_product_incomings (id, shop_product_id, quantity)
		VALUES ($1,$2,$3)
		RETURNING id
	`, item.ID, item.ShopProductID, item.Quantity).Scan(&item.ID); err != nil {
		return Incoming{}, err
	}
	tag, err := tx.Exec(ctx, `
		UPDATE shop_products SET quantity = quantity + $2, updated_at = NOW()
		WHERE id = $1
	`, item.ShopProductID, item.Quantity)
	if err != nil {
		return Incoming{}, err
	}
	if tag.RowsAffected() == 0 {
		return Incoming{}, ErrNotFound
	}
	if err = tx.Commit(ctx); err != nil {
		return Incoming{}, err
	}
	return r.FindIncoming(ctx, item.ID)
}

func (r *Repository) FindIncoming(ctx context.Context, id uuid.UUID) (Incoming, error) {
	return scanIncoming(r.pool.QueryRow(ctx, `SELECT `+incomingCols+incomingJoins()+` WHERE i.id = $1`, id))
}

func (r *Repository) ListIncomings(ctx context.Context, q IncomingListQuery) ([]Incoming, int, error) {
	where := []string{"1=1"}
	args := []any{}
	add := func(cond string, val any) {
		args = append(args, val)
		where = append(where, strings.Replace(cond, "?", "$"+itoa(len(args)), 1))
	}
	if q.ShopID != nil {
		add("s.id = ?", *q.ShopID)
	}
	if q.ShopProductID != nil {
		add("i.shop_product_id = ?", *q.ShopProductID)
	}
	if q.RegionID != nil {
		add("s.region_id = ?", *q.RegionID)
	}
	if q.DistrictID != nil {
		add("s.district_id = ?", *q.DistrictID)
	}
	if q.MFYID != nil {
		add("s.mfy_id = ?", *q.MFYID)
	}
	if q.Query != "" {
		args = append(args, "%"+q.Query+"%")
		n := itoa(len(args))
		where = append(where, "(t.name ILIKE $"+n+" OR s.name ILIKE $"+n+")")
	}

	clause := strings.Join(where, " AND ")
	var total int
	if err := r.pool.QueryRow(ctx, `SELECT COUNT(*) `+incomingJoins()+` WHERE `+clause, args...).Scan(&total); err != nil {
		return nil, 0, err
	}

	offset := (q.Page - 1) * q.Limit
	args = append(args, q.Limit, offset)
	limitPos := itoa(len(args) - 1)
	offsetPos := itoa(len(args))
	rows, err := r.pool.Query(ctx, `
		SELECT `+incomingCols+incomingJoins()+`
		WHERE `+clause+`
		ORDER BY i.created_at DESC
		LIMIT $`+limitPos+` OFFSET $`+offsetPos, args...)
	if err != nil {
		return nil, 0, err
	}
	defer rows.Close()

	items := make([]Incoming, 0)
	for rows.Next() {
		item, err := scanIncoming(rows)
		if err != nil {
			return nil, 0, err
		}
		items = append(items, item)
	}
	return items, total, rows.Err()
}

func (r *Repository) FindShop(ctx context.Context, id uuid.UUID) (ShopRef, error) {
	var item ShopRef
	err := r.pool.QueryRow(ctx, `
		SELECT id, name, region_id, district_id, mfy_id, status FROM local_shops WHERE id = $1
	`, id).Scan(&item.ID, &item.Name, &item.RegionID, &item.DistrictID, &item.MFYID, &item.Status)
	if errors.Is(err, pgx.ErrNoRows) {
		return ShopRef{}, ErrNotFound
	}
	return item, err
}

func (r *Repository) FindTemplate(ctx context.Context, id uuid.UUID) (TemplateRef, error) {
	var item TemplateRef
	err := r.pool.QueryRow(ctx, `SELECT id, name FROM shop_product_templates WHERE id = $1`, id).Scan(&item.ID, &item.Name)
	if errors.Is(err, pgx.ErrNoRows) {
		return TemplateRef{}, ErrNotFound
	}
	return item, err
}

func itoa(n int) string {
	if n == 0 {
		return "0"
	}
	var b [12]byte
	i := len(b)
	for n > 0 {
		i--
		b[i] = byte('0' + n%10)
		n /= 10
	}
	return string(b[i:])
}
