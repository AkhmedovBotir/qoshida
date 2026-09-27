package category

import (
	"context"
	"errors"
	"fmt"
	"strings"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

var ErrNotFound = errors.New("category not found")
var errHasChildren = errors.New("category has children")

type Repository struct {
	pool *pgxpool.Pool
}

func NewRepository(pool *pgxpool.Pool) *Repository {
	return &Repository{pool: pool}
}

const cols = `c.id, c.source_id, c.name, c.slug, c.parent_id, COALESCE(c.parent_source_id, ''), COALESCE(p.name, ''), COALESCE(c.image_url, ''), c.censored, c.status, c.created_at`

func scan(row pgx.Row) (Category, error) {
	var item Category
	err := row.Scan(
		&item.ID, &item.SourceID, &item.Name, &item.Slug, &item.ParentID,
		&item.ParentSourceID, &item.ParentName, &item.ImageURL, &item.Censored, &item.Status, &item.CreatedAt,
	)
	if errors.Is(err, pgx.ErrNoRows) {
		return Category{}, ErrNotFound
	}
	return item, err
}

func (r *Repository) Create(ctx context.Context, item Category) (Category, error) {
	created, err := scan(r.pool.QueryRow(ctx, `
		INSERT INTO categories (source_id, name, slug, parent_id, parent_source_id, image_url, censored, status)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8)
		RETURNING id, source_id, name, slug, parent_id, COALESCE(parent_source_id, ''), '', COALESCE(image_url, ''), censored, status, created_at
	`, item.SourceID, item.Name, item.Slug, item.ParentID, nilIfEmpty(item.ParentSourceID), nilIfEmpty(item.ImageURL), item.Censored, item.Status))
	if err != nil {
		return Category{}, err
	}
	return r.FindByID(ctx, created.ID)
}

func (r *Repository) Update(ctx context.Context, item Category) (Category, error) {
	_, err := scan(r.pool.QueryRow(ctx, `
		UPDATE categories
		SET name = $2, slug = $3, parent_id = $4, parent_source_id = $5, image_url = $6, censored = $7, status = $8, updated_at = NOW()
		WHERE id = $1
		RETURNING id, source_id, name, slug, parent_id, COALESCE(parent_source_id, ''), '', COALESCE(image_url, ''), censored, status, created_at
	`, item.ID, item.Name, item.Slug, item.ParentID, nilIfEmpty(item.ParentSourceID), nilIfEmpty(item.ImageURL), item.Censored, item.Status))
	if err != nil {
		return Category{}, err
	}
	return r.FindByID(ctx, item.ID)
}

func (r *Repository) Delete(ctx context.Context, id uuid.UUID) error {
	var children int
	if err := r.pool.QueryRow(ctx, `SELECT COUNT(*) FROM categories WHERE parent_id = $1`, id).Scan(&children); err != nil {
		return err
	}
	if children > 0 {
		return errHasChildren
	}
	tag, err := r.pool.Exec(ctx, `DELETE FROM categories WHERE id = $1`, id)
	if err != nil {
		return err
	}
	if tag.RowsAffected() == 0 {
		return ErrNotFound
	}
	return nil
}

func (r *Repository) FindByID(ctx context.Context, id uuid.UUID) (Category, error) {
	return scan(r.pool.QueryRow(ctx, `
		SELECT `+cols+`
		FROM categories c
		LEFT JOIN categories p ON p.id = c.parent_id
		WHERE c.id = $1
	`, id))
}

func (r *Repository) List(ctx context.Context, q ListQuery) ([]Category, int, error) {
	where := []string{"TRUE"}
	args := make([]any, 0, 6)
	if q.Roots {
		where = append(where, "c.parent_id IS NULL")
	}
	if q.ParentID != nil {
		args = append(args, *q.ParentID)
		where = append(where, fmt.Sprintf("c.parent_id = $%d", len(args)))
	}
	if q.Query != "" {
		args = append(args, "%"+q.Query+"%")
		where = append(where, fmt.Sprintf("(c.name ILIKE $%d OR c.slug ILIKE $%d)", len(args), len(args)))
	}
	clause := strings.Join(where, " AND ")

	var total int
	if err := r.pool.QueryRow(ctx, `SELECT COUNT(*) FROM categories c WHERE `+clause, args...).Scan(&total); err != nil {
		return nil, 0, err
	}

	args = append(args, q.Limit, (q.Page-1)*q.Limit)
	rows, err := r.pool.Query(ctx, `
		SELECT `+cols+`
		FROM categories c
		LEFT JOIN categories p ON p.id = c.parent_id
		WHERE `+clause+`
		ORDER BY CASE WHEN c.parent_id IS NULL THEN 0 ELSE 1 END, c.name
		LIMIT $`+fmt.Sprintf("%d", len(args)-1)+` OFFSET $`+fmt.Sprintf("%d", len(args))+`
	`, args...)
	if err != nil {
		return nil, 0, err
	}
	defer rows.Close()

	items := make([]Category, 0, q.Limit)
	for rows.Next() {
		item, err := scan(rows)
		if err != nil {
			return nil, 0, err
		}
		items = append(items, item)
	}
	return items, total, rows.Err()
}

func (r *Repository) Stats(ctx context.Context) (Stats, error) {
	var s Stats
	err := r.pool.QueryRow(ctx, `
		SELECT
			COUNT(*),
			COUNT(*) FILTER (WHERE parent_id IS NULL),
			COUNT(*) FILTER (WHERE parent_id IS NOT NULL)
		FROM categories
	`).Scan(&s.Total, &s.Root, &s.Child)
	return s, err
}

func (r *Repository) UpsertMany(ctx context.Context, items []Category) (ImportResult, error) {
	tx, err := r.pool.Begin(ctx)
	if err != nil {
		return ImportResult{}, err
	}
	defer tx.Rollback(ctx)

	var result ImportResult
	const chunk = 200
	for start := 0; start < len(items); start += chunk {
		end := start + chunk
		if end > len(items) {
			end = len(items)
		}
		part := items[start:end]
		var b strings.Builder
		args := make([]any, 0, len(part)*7)
		b.WriteString(`INSERT INTO categories (source_id, name, slug, parent_source_id, image_url, censored, status) VALUES `)
		for i, item := range part {
			if i > 0 {
				b.WriteString(",")
			}
			n := i * 7
			fmt.Fprintf(&b, "($%d,$%d,$%d,$%d,$%d,$%d,$%d)", n+1, n+2, n+3, n+4, n+5, n+6, n+7)
			args = append(args, item.SourceID, item.Name, item.Slug, nilIfEmpty(item.ParentSourceID), nilIfEmpty(item.ImageURL), item.Censored, item.Status)
		}
		b.WriteString(`
			ON CONFLICT (source_id) DO UPDATE SET
				name = EXCLUDED.name,
				slug = EXCLUDED.slug,
				parent_source_id = EXCLUDED.parent_source_id,
				image_url = COALESCE(EXCLUDED.image_url, categories.image_url),
				censored = EXCLUDED.censored,
				status = EXCLUDED.status,
				updated_at = NOW()
			RETURNING (xmax = 0) AS inserted
		`)
		rows, err := tx.Query(ctx, b.String(), args...)
		if err != nil {
			return ImportResult{}, err
		}
		for rows.Next() {
			var inserted bool
			if err = rows.Scan(&inserted); err != nil {
				rows.Close()
				return ImportResult{}, err
			}
			if inserted {
				result.Inserted++
			} else {
				result.Updated++
			}
		}
		rows.Close()
		if err = rows.Err(); err != nil {
			return ImportResult{}, err
		}
	}

	if _, err = tx.Exec(ctx, `
		UPDATE categories AS child
		SET parent_id = parent.id, updated_at = NOW()
		FROM categories AS parent
		WHERE child.parent_source_id IS NOT NULL
		  AND child.parent_source_id = parent.source_id
		  AND child.parent_id IS DISTINCT FROM parent.id
	`); err != nil {
		return ImportResult{}, err
	}

	if err = tx.Commit(ctx); err != nil {
		return ImportResult{}, err
	}
	result.Total = result.Inserted + result.Updated
	return result, nil
}

func nilIfEmpty(v string) any {
	if v == "" {
		return nil
	}
	return v
}
