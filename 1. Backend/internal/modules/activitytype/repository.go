package activitytype

import (
	"context"
	"errors"
	"fmt"
	"strings"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

var ErrNotFound = errors.New("activity type not found")

type Repository struct {
	pool *pgxpool.Pool
}

func NewRepository(pool *pgxpool.Pool) *Repository {
	return &Repository{pool: pool}
}

const cols = `id, source_id, name, icon, status, created_at`

func scan(row pgx.Row) (ActivityType, error) {
	var item ActivityType
	err := row.Scan(&item.ID, &item.SourceID, &item.Name, &item.Icon, &item.Status, &item.CreatedAt)
	if errors.Is(err, pgx.ErrNoRows) {
		return ActivityType{}, ErrNotFound
	}
	return item, err
}

func (r *Repository) Create(ctx context.Context, item ActivityType) (ActivityType, error) {
	return scan(r.pool.QueryRow(ctx, `
		INSERT INTO activity_types (source_id, name, icon, status)
		VALUES ($1, $2, $3, $4)
		RETURNING `+cols+`
	`, item.SourceID, item.Name, item.Icon, item.Status))
}

func (r *Repository) Update(ctx context.Context, item ActivityType) (ActivityType, error) {
	return scan(r.pool.QueryRow(ctx, `
		UPDATE activity_types
		SET name = $2, icon = $3, status = $4, updated_at = NOW()
		WHERE id = $1
		RETURNING `+cols+`
	`, item.ID, item.Name, item.Icon, item.Status))
}

func (r *Repository) Delete(ctx context.Context, id uuid.UUID) error {
	tag, err := r.pool.Exec(ctx, `DELETE FROM activity_types WHERE id = $1`, id)
	if err != nil {
		return err
	}
	if tag.RowsAffected() == 0 {
		return ErrNotFound
	}
	return nil
}

func (r *Repository) FindByID(ctx context.Context, id uuid.UUID) (ActivityType, error) {
	return scan(r.pool.QueryRow(ctx, `SELECT `+cols+` FROM activity_types WHERE id = $1`, id))
}

func (r *Repository) List(ctx context.Context, q ListQuery) ([]ActivityType, int, error) {
	where := []string{"TRUE"}
	args := make([]any, 0, 4)
	if q.Query != "" {
		args = append(args, "%"+q.Query+"%")
		where = append(where, fmt.Sprintf("(name ILIKE $%d OR icon ILIKE $%d)", len(args), len(args)))
	}
	if q.Status != "" {
		args = append(args, q.Status)
		where = append(where, fmt.Sprintf("status = $%d", len(args)))
	}
	clause := strings.Join(where, " AND ")

	var total int
	if err := r.pool.QueryRow(ctx, `SELECT COUNT(*) FROM activity_types WHERE `+clause, args...).Scan(&total); err != nil {
		return nil, 0, err
	}

	args = append(args, q.Limit, (q.Page-1)*q.Limit)
	rows, err := r.pool.Query(ctx, `
		SELECT `+cols+`
		FROM activity_types
		WHERE `+clause+`
		ORDER BY name
		LIMIT $`+fmt.Sprintf("%d", len(args)-1)+` OFFSET $`+fmt.Sprintf("%d", len(args))+`
	`, args...)
	if err != nil {
		return nil, 0, err
	}
	defer rows.Close()

	items := make([]ActivityType, 0, q.Limit)
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
			COUNT(*) FILTER (WHERE status = 'active'),
			COUNT(*) FILTER (WHERE status = 'inactive')
		FROM activity_types
	`).Scan(&s.Total, &s.Active, &s.Inactive)
	return s, err
}

func (r *Repository) UpsertMany(ctx context.Context, items []ActivityType) (ImportResult, error) {
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
		args := make([]any, 0, len(part)*4)
		b.WriteString(`INSERT INTO activity_types (source_id, name, icon, status) VALUES `)
		for i, item := range part {
			if i > 0 {
				b.WriteString(",")
			}
			n := i * 4
			fmt.Fprintf(&b, "($%d,$%d,$%d,$%d)", n+1, n+2, n+3, n+4)
			args = append(args, item.SourceID, item.Name, item.Icon, item.Status)
		}
		b.WriteString(`
			ON CONFLICT (source_id) DO UPDATE SET
				name = EXCLUDED.name,
				icon = EXCLUDED.icon,
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

	if err = tx.Commit(ctx); err != nil {
		return ImportResult{}, err
	}
	result.Total = result.Inserted + result.Updated
	return result, nil
}
