package region

import (
	"context"
	"errors"
	"fmt"
	"strings"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

var ErrNotFound = errors.New("region not found")

type Repository struct {
	pool *pgxpool.Pool
}

func NewRepository(pool *pgxpool.Pool) *Repository {
	return &Repository{pool: pool}
}

const cols = `r.id, r.source_id, r.name, r.type, r.parent_id, COALESCE(r.parent_source_id, ''), COALESCE(p.name, ''), r.code, r.status, r.created_at`

func scan(row pgx.Row) (Region, error) {
	var item Region
	err := row.Scan(
		&item.ID, &item.SourceID, &item.Name, &item.Type, &item.ParentID,
		&item.ParentSourceID, &item.ParentName, &item.Code, &item.Status, &item.CreatedAt,
	)
	if errors.Is(err, pgx.ErrNoRows) {
		return Region{}, ErrNotFound
	}
	return item, err
}

func (r *Repository) Create(ctx context.Context, item Region) (Region, error) {
	created, err := scan(r.pool.QueryRow(ctx, `
		INSERT INTO regions (source_id, name, type, parent_id, parent_source_id, code, status)
		VALUES ($1, $2, $3, $4, $5, $6, $7)
		RETURNING id, source_id, name, type, parent_id, COALESCE(parent_source_id, ''), '', code, status, created_at
	`, item.SourceID, item.Name, item.Type, item.ParentID, nilIfEmpty(item.ParentSourceID), item.Code, item.Status))
	if err != nil {
		return Region{}, err
	}
	return r.FindByID(ctx, created.ID)
}

func (r *Repository) Update(ctx context.Context, item Region) (Region, error) {
	_, err := scan(r.pool.QueryRow(ctx, `
		UPDATE regions
		SET name = $2, type = $3, parent_id = $4, parent_source_id = $5, code = $6, status = $7, updated_at = NOW()
		WHERE id = $1
		RETURNING id, source_id, name, type, parent_id, COALESCE(parent_source_id, ''), '', code, status, created_at
	`, item.ID, item.Name, item.Type, item.ParentID, nilIfEmpty(item.ParentSourceID), item.Code, item.Status))
	if err != nil {
		return Region{}, err
	}
	return r.FindByID(ctx, item.ID)
}

func (r *Repository) Delete(ctx context.Context, id uuid.UUID) error {
	var children int
	if err := r.pool.QueryRow(ctx, `SELECT COUNT(*) FROM regions WHERE parent_id = $1`, id).Scan(&children); err != nil {
		return err
	}
	if children > 0 {
		return errHasChildren
	}
	tag, err := r.pool.Exec(ctx, `DELETE FROM regions WHERE id = $1`, id)
	if err != nil {
		return err
	}
	if tag.RowsAffected() == 0 {
		return ErrNotFound
	}
	return nil
}

var errHasChildren = errors.New("region has children")

func (r *Repository) FindByID(ctx context.Context, id uuid.UUID) (Region, error) {
	return scan(r.pool.QueryRow(ctx, `
		SELECT `+cols+`
		FROM regions r
		LEFT JOIN regions p ON p.id = r.parent_id
		WHERE r.id = $1
	`, id))
}

func (r *Repository) List(ctx context.Context, q ListQuery) ([]Region, int, error) {
	where := []string{"TRUE"}
	args := make([]any, 0, 6)
	if q.Type != "" {
		args = append(args, q.Type)
		where = append(where, fmt.Sprintf("r.type = $%d", len(args)))
	}
	if q.ParentID != nil {
		args = append(args, *q.ParentID)
		where = append(where, fmt.Sprintf("r.parent_id = $%d", len(args)))
	}
	if q.Query != "" {
		args = append(args, "%"+q.Query+"%")
		where = append(where, fmt.Sprintf("(r.name ILIKE $%d OR r.code ILIKE $%d)", len(args), len(args)))
	}
	clause := strings.Join(where, " AND ")

	var total int
	if err := r.pool.QueryRow(ctx, `SELECT COUNT(*) FROM regions r WHERE `+clause, args...).Scan(&total); err != nil {
		return nil, 0, err
	}

	args = append(args, q.Limit, (q.Page-1)*q.Limit)
	rows, err := r.pool.Query(ctx, `
		SELECT `+cols+`
		FROM regions r
		LEFT JOIN regions p ON p.id = r.parent_id
		WHERE `+clause+`
		ORDER BY CASE r.type WHEN 'region' THEN 0 WHEN 'district' THEN 1 ELSE 2 END, r.name
		LIMIT $`+fmt.Sprintf("%d", len(args)-1)+` OFFSET $`+fmt.Sprintf("%d", len(args))+`
	`, args...)
	if err != nil {
		return nil, 0, err
	}
	defer rows.Close()

	items := make([]Region, 0, q.Limit)
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
			COUNT(*) FILTER (WHERE type = 'region'),
			COUNT(*) FILTER (WHERE type = 'district'),
			COUNT(*) FILTER (WHERE type = 'mfy')
		FROM regions
	`).Scan(&s.Total, &s.Region, &s.District, &s.MFY)
	return s, err
}

func (r *Repository) UpsertMany(ctx context.Context, items []Region) (ImportResult, error) {
	tx, err := r.pool.Begin(ctx)
	if err != nil {
		return ImportResult{}, err
	}
	defer tx.Rollback(ctx)

	var result ImportResult
	const chunk = 400
	for start := 0; start < len(items); start += chunk {
		end := start + chunk
		if end > len(items) {
			end = len(items)
		}
		part := items[start:end]
		var b strings.Builder
		args := make([]any, 0, len(part)*6)
		b.WriteString(`
			INSERT INTO regions (source_id, name, type, parent_source_id, code, status)
			VALUES
		`)
		for i, item := range part {
			if i > 0 {
				b.WriteString(",")
			}
			n := i * 6
			fmt.Fprintf(&b, "($%d,$%d,$%d,$%d,$%d,$%d)", n+1, n+2, n+3, n+4, n+5, n+6)
			args = append(args, item.SourceID, item.Name, item.Type, nilIfEmpty(item.ParentSourceID), item.Code, item.Status)
		}
		b.WriteString(`
			ON CONFLICT (source_id) DO UPDATE SET
				name = EXCLUDED.name,
				type = EXCLUDED.type,
				parent_source_id = EXCLUDED.parent_source_id,
				code = EXCLUDED.code,
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
		UPDATE regions AS child
		SET parent_id = parent.id, updated_at = NOW()
		FROM regions AS parent
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
