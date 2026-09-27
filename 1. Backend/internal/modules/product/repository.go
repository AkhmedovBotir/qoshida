package product

import (
	"context"
	"errors"
	"strings"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

var ErrNotFound = errors.New("product not found")

type Repository struct {
	pool *pgxpool.Pool
}

func NewRepository(pool *pgxpool.Pool) *Repository {
	return &Repository{pool: pool}
}

const cols = `
	p.id, p.kontragent_id, p.category_id, p.subcategory_id, p.name, p.description,
	p.sale_price, p.cost_price, p.quantity, p.unit, p.unit_size, p.commission_percent,
	COALESCE(p.images, '{}'), p.approval_status, p.rejection_note, p.submitted_by, p.status,
	COALESCE(p.reviewed_by_name, ''), COALESCE(p.reviewed_by_role, ''), p.reviewed_at,
	p.created_at, p.updated_at,
	COALESCE(k.name, ''), COALESCE(c.name, ''), COALESCE(sc.name, ''),
	k.region_id, k.district_id, k.mfy_id,
	COALESCE(reg.name, ''), COALESCE(dis.name, ''), COALESCE(mfy.name, '')
`

func scan(row pgx.Row) (Product, error) {
	var item Product
	err := row.Scan(
		&item.ID, &item.KontragentID, &item.CategoryID, &item.SubcategoryID, &item.Name, &item.Description,
		&item.SalePrice, &item.CostPrice, &item.Quantity, &item.Unit, &item.UnitSize, &item.CommissionPercent,
		&item.Images, &item.ApprovalStatus, &item.RejectionNote, &item.SubmittedBy, &item.Status,
		&item.ReviewedByName, &item.ReviewedByRole, &item.ReviewedAt,
		&item.CreatedAt, &item.UpdatedAt,
		&item.KontragentName, &item.CategoryName, &item.SubcategoryName,
		&item.RegionID, &item.DistrictID, &item.MFYID,
		&item.RegionName, &item.DistrictName, &item.MFYName,
	)
	if errors.Is(err, pgx.ErrNoRows) {
		return Product{}, ErrNotFound
	}
	return item, err
}

func (r *Repository) Create(ctx context.Context, item Product) (Product, error) {
	if item.ID == uuid.Nil {
		item.ID = uuid.New()
	}
	err := r.pool.QueryRow(ctx, `
		INSERT INTO products (
			id, kontragent_id, category_id, subcategory_id, name, description,
			sale_price, cost_price, quantity, unit, unit_size, commission_percent,
			images, approval_status, rejection_note, submitted_by, status,
			reviewed_by_name, reviewed_by_role, reviewed_at
		) VALUES ($1,$2,$3,$4,$5,$6,$7,$8,$9,$10,$11,$12,$13,$14,$15,$16,$17,$18,$19,$20)
		RETURNING id
	`, item.ID, item.KontragentID, item.CategoryID, item.SubcategoryID, item.Name, item.Description,
		item.SalePrice, item.CostPrice, item.Quantity, item.Unit, item.UnitSize, item.CommissionPercent,
		item.Images, item.ApprovalStatus, item.RejectionNote, item.SubmittedBy, item.Status,
		item.ReviewedByName, item.ReviewedByRole, item.ReviewedAt,
	).Scan(&item.ID)
	if err != nil {
		return Product{}, err
	}
	return r.FindByID(ctx, item.ID)
}

func (r *Repository) Update(ctx context.Context, item Product) (Product, error) {
	tag, err := r.pool.Exec(ctx, `
		UPDATE products SET
			kontragent_id = $2, category_id = $3, subcategory_id = $4, name = $5, description = $6,
			sale_price = $7, cost_price = $8, quantity = $9, unit = $10, unit_size = $11,
			commission_percent = $12, images = $13, approval_status = $14, rejection_note = $15,
			submitted_by = $16, status = $17, reviewed_by_name = $18, reviewed_by_role = $19,
			reviewed_at = $20, updated_at = NOW()
		WHERE id = $1
	`, item.ID, item.KontragentID, item.CategoryID, item.SubcategoryID, item.Name, item.Description,
		item.SalePrice, item.CostPrice, item.Quantity, item.Unit, item.UnitSize, item.CommissionPercent,
		item.Images, item.ApprovalStatus, item.RejectionNote, item.SubmittedBy, item.Status,
		item.ReviewedByName, item.ReviewedByRole, item.ReviewedAt)
	if err != nil {
		return Product{}, err
	}
	if tag.RowsAffected() == 0 {
		return Product{}, ErrNotFound
	}
	return r.FindByID(ctx, item.ID)
}

func (r *Repository) Delete(ctx context.Context, id uuid.UUID) error {
	tag, err := r.pool.Exec(ctx, `DELETE FROM products WHERE id = $1`, id)
	if err != nil {
		return err
	}
	if tag.RowsAffected() == 0 {
		return ErrNotFound
	}
	return nil
}

func (r *Repository) FindByID(ctx context.Context, id uuid.UUID) (Product, error) {
	return scan(r.pool.QueryRow(ctx, `
		SELECT `+cols+`
		FROM products p
		JOIN kontragents k ON k.id = p.kontragent_id
		LEFT JOIN categories c ON c.id = p.category_id
		LEFT JOIN categories sc ON sc.id = p.subcategory_id
		LEFT JOIN regions reg ON reg.id = k.region_id
		LEFT JOIN regions dis ON dis.id = k.district_id
		LEFT JOIN regions mfy ON mfy.id = k.mfy_id
		WHERE p.id = $1
	`, id))
}

func (r *Repository) List(ctx context.Context, q ListQuery) ([]Product, int, error) {
	where := []string{"1=1"}
	args := []any{}
	add := func(cond string, val any) {
		args = append(args, val)
		where = append(where, strings.Replace(cond, "?", "$"+itoa(len(args)), 1))
	}
	if q.KontragentID != nil {
		add("p.kontragent_id = ?", *q.KontragentID)
	}
	if q.CategoryID != nil {
		add("p.category_id = ?", *q.CategoryID)
	}
	if q.ApprovalStatus != "" {
		add("p.approval_status = ?", q.ApprovalStatus)
	}
	if q.RegionID != nil {
		add("k.region_id = ?", *q.RegionID)
	}
	if q.DistrictID != nil {
		add("k.district_id = ?", *q.DistrictID)
	}
	if q.MFYID != nil {
		add("k.mfy_id = ?", *q.MFYID)
	}
	if q.Query != "" {
		args = append(args, "%"+q.Query+"%")
		n := itoa(len(args))
		where = append(where, "(p.name ILIKE $"+n+" OR k.name ILIKE $"+n+")")
	}

	clause := strings.Join(where, " AND ")
	var total int
	if err := r.pool.QueryRow(ctx, `
		SELECT COUNT(*)
		FROM products p
		JOIN kontragents k ON k.id = p.kontragent_id
		WHERE `+clause, args...).Scan(&total); err != nil {
		return nil, 0, err
	}

	offset := (q.Page - 1) * q.Limit
	args = append(args, q.Limit, offset)
	limitPos := itoa(len(args) - 1)
	offsetPos := itoa(len(args))
	rows, err := r.pool.Query(ctx, `
		SELECT `+cols+`
		FROM products p
		JOIN kontragents k ON k.id = p.kontragent_id
		LEFT JOIN categories c ON c.id = p.category_id
		LEFT JOIN categories sc ON sc.id = p.subcategory_id
		LEFT JOIN regions reg ON reg.id = k.region_id
		LEFT JOIN regions dis ON dis.id = k.district_id
		LEFT JOIN regions mfy ON mfy.id = k.mfy_id
		WHERE `+clause+`
		ORDER BY p.created_at DESC
		LIMIT $`+limitPos+` OFFSET $`+offsetPos, args...)
	if err != nil {
		return nil, 0, err
	}
	defer rows.Close()

	items := make([]Product, 0)
	for rows.Next() {
		item, err := scan(rows)
		if err != nil {
			return nil, 0, err
		}
		items = append(items, item)
	}
	return items, total, rows.Err()
}

func (r *Repository) FindKontragent(ctx context.Context, id uuid.UUID) (KontragentRef, error) {
	var item KontragentRef
	err := r.pool.QueryRow(ctx, `
		SELECT id, name, region_id, district_id, mfy_id, status
		FROM kontragents WHERE id = $1
	`, id).Scan(&item.ID, &item.Name, &item.RegionID, &item.DistrictID, &item.MFYID, &item.Status)
	if errors.Is(err, pgx.ErrNoRows) {
		return KontragentRef{}, ErrNotFound
	}
	return item, err
}

func (r *Repository) FindCategory(ctx context.Context, id uuid.UUID) (CategoryRef, error) {
	var item CategoryRef
	err := r.pool.QueryRow(ctx, `
		SELECT id, parent_id, name, status FROM categories WHERE id = $1
	`, id).Scan(&item.ID, &item.ParentID, &item.Name, &item.Status)
	if errors.Is(err, pgx.ErrNoRows) {
		return CategoryRef{}, ErrNotFound
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
