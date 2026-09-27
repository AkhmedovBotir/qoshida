package shoptemplate

import (
	"context"
	"errors"
	"strings"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

var ErrNotFound = errors.New("shop product template not found")

type Repository struct {
	pool *pgxpool.Pool
}

func NewRepository(pool *pgxpool.Pool) *Repository {
	return &Repository{pool: pool}
}

const cols = `
	t.id, t.category_id, t.subcategory_id, t.name, t.description, t.unit, t.unit_size,
	COALESCE(t.images, '{}'), t.created_at, t.updated_at,
	COALESCE(c.name, ''), COALESCE(sc.name, '')
`

func scan(row pgx.Row) (Template, error) {
	var item Template
	err := row.Scan(
		&item.ID, &item.CategoryID, &item.SubcategoryID, &item.Name, &item.Description, &item.Unit, &item.UnitSize,
		&item.Images, &item.CreatedAt, &item.UpdatedAt,
		&item.CategoryName, &item.SubcategoryName,
	)
	if errors.Is(err, pgx.ErrNoRows) {
		return Template{}, ErrNotFound
	}
	return item, err
}

func (r *Repository) Create(ctx context.Context, item Template) (Template, error) {
	if item.ID == uuid.Nil {
		item.ID = uuid.New()
	}
	err := r.pool.QueryRow(ctx, `
		INSERT INTO shop_product_templates (
			id, category_id, subcategory_id, name, description, unit, unit_size, images
		) VALUES ($1,$2,$3,$4,$5,$6,$7,$8)
		RETURNING id
	`, item.ID, item.CategoryID, item.SubcategoryID, item.Name, item.Description, item.Unit, item.UnitSize, item.Images).Scan(&item.ID)
	if err != nil {
		return Template{}, err
	}
	return r.FindByID(ctx, item.ID)
}

func (r *Repository) Update(ctx context.Context, item Template) (Template, error) {
	tag, err := r.pool.Exec(ctx, `
		UPDATE shop_product_templates SET
			category_id = $2, subcategory_id = $3, name = $4, description = $5,
			unit = $6, unit_size = $7, images = $8, updated_at = NOW()
		WHERE id = $1
	`, item.ID, item.CategoryID, item.SubcategoryID, item.Name, item.Description, item.Unit, item.UnitSize, item.Images)
	if err != nil {
		return Template{}, err
	}
	if tag.RowsAffected() == 0 {
		return Template{}, ErrNotFound
	}
	return r.FindByID(ctx, item.ID)
}

func (r *Repository) Delete(ctx context.Context, id uuid.UUID) error {
	tag, err := r.pool.Exec(ctx, `DELETE FROM shop_product_templates WHERE id = $1`, id)
	if err != nil {
		return err
	}
	if tag.RowsAffected() == 0 {
		return ErrNotFound
	}
	return nil
}

func (r *Repository) FindByID(ctx context.Context, id uuid.UUID) (Template, error) {
	return scan(r.pool.QueryRow(ctx, `
		SELECT `+cols+`
		FROM shop_product_templates t
		LEFT JOIN categories c ON c.id = t.category_id
		LEFT JOIN categories sc ON sc.id = t.subcategory_id
		WHERE t.id = $1
	`, id))
}

func (r *Repository) List(ctx context.Context, q ListQuery) ([]Template, int, error) {
	where := []string{"1=1"}
	args := []any{}
	add := func(cond string, val any) {
		args = append(args, val)
		where = append(where, strings.Replace(cond, "?", "$"+itoa(len(args)), 1))
	}
	if q.CategoryID != nil {
		add("t.category_id = ?", *q.CategoryID)
	}
	if q.Query != "" {
		args = append(args, "%"+q.Query+"%")
		n := itoa(len(args))
		where = append(where, "(t.name ILIKE $"+n+" OR t.description ILIKE $"+n+")")
	}

	clause := strings.Join(where, " AND ")
	var total int
	if err := r.pool.QueryRow(ctx, `
		SELECT COUNT(*) FROM shop_product_templates t WHERE `+clause, args...).Scan(&total); err != nil {
		return nil, 0, err
	}

	offset := (q.Page - 1) * q.Limit
	args = append(args, q.Limit, offset)
	limitPos := itoa(len(args) - 1)
	offsetPos := itoa(len(args))
	rows, err := r.pool.Query(ctx, `
		SELECT `+cols+`
		FROM shop_product_templates t
		LEFT JOIN categories c ON c.id = t.category_id
		LEFT JOIN categories sc ON sc.id = t.subcategory_id
		WHERE `+clause+`
		ORDER BY t.created_at DESC
		LIMIT $`+limitPos+` OFFSET $`+offsetPos, args...)
	if err != nil {
		return nil, 0, err
	}
	defer rows.Close()

	items := make([]Template, 0)
	for rows.Next() {
		item, err := scan(rows)
		if err != nil {
			return nil, 0, err
		}
		items = append(items, item)
	}
	return items, total, rows.Err()
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
