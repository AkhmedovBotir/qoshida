package commenttemplate

import (
	"context"
	"errors"
	"fmt"
	"strings"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

var ErrNotFound = errors.New("comment template not found")

type Repository struct {
	pool *pgxpool.Pool
}

func NewRepository(pool *pgxpool.Pool) *Repository {
	return &Repository{pool: pool}
}

const cols = `id, comment, sort_order, status, created_at, updated_at`

func scan(row pgx.Row) (CommentTemplate, error) {
	var item CommentTemplate
	err := row.Scan(&item.ID, &item.Comment, &item.SortOrder, &item.Status, &item.CreatedAt, &item.UpdatedAt)
	if errors.Is(err, pgx.ErrNoRows) {
		return CommentTemplate{}, ErrNotFound
	}
	return item, err
}

func (r *Repository) Create(ctx context.Context, item CommentTemplate) (CommentTemplate, error) {
	return scan(r.pool.QueryRow(ctx, `
		INSERT INTO comment_templates (comment, sort_order, status)
		VALUES ($1, $2, $3)
		RETURNING `+cols+`
	`, item.Comment, item.SortOrder, item.Status))
}

func (r *Repository) Update(ctx context.Context, item CommentTemplate) (CommentTemplate, error) {
	return scan(r.pool.QueryRow(ctx, `
		UPDATE comment_templates
		SET comment = $2, status = $3, updated_at = NOW()
		WHERE id = $1
		RETURNING `+cols+`
	`, item.ID, item.Comment, item.Status))
}

func (r *Repository) Delete(ctx context.Context, id uuid.UUID) error {
	tag, err := r.pool.Exec(ctx, `DELETE FROM comment_templates WHERE id = $1`, id)
	if err != nil {
		return err
	}
	if tag.RowsAffected() == 0 {
		return ErrNotFound
	}
	return nil
}

func (r *Repository) FindByID(ctx context.Context, id uuid.UUID) (CommentTemplate, error) {
	return scan(r.pool.QueryRow(ctx, `SELECT `+cols+` FROM comment_templates WHERE id = $1`, id))
}

func (r *Repository) MaxSortOrder(ctx context.Context) (int, error) {
	var max int
	err := r.pool.QueryRow(ctx, `SELECT COALESCE(MAX(sort_order), 0) FROM comment_templates`).Scan(&max)
	return max, err
}

func (r *Repository) List(ctx context.Context, q ListQuery) ([]CommentTemplate, int, error) {
	where := []string{"TRUE"}
	args := make([]any, 0, 4)
	if q.Query != "" {
		args = append(args, "%"+q.Query+"%")
		where = append(where, fmt.Sprintf("comment ILIKE $%d", len(args)))
	}
	clause := strings.Join(where, " AND ")

	var total int
	if err := r.pool.QueryRow(ctx, `SELECT COUNT(*) FROM comment_templates WHERE `+clause, args...).Scan(&total); err != nil {
		return nil, 0, err
	}

	args = append(args, q.Limit, (q.Page-1)*q.Limit)
	rows, err := r.pool.Query(ctx, `
		SELECT `+cols+`
		FROM comment_templates
		WHERE `+clause+`
		ORDER BY sort_order ASC, id ASC
		LIMIT $`+fmt.Sprintf("%d", len(args)-1)+` OFFSET $`+fmt.Sprintf("%d", len(args))+`
	`, args...)
	if err != nil {
		return nil, 0, err
	}
	defer rows.Close()

	items := make([]CommentTemplate, 0, q.Limit)
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
		FROM comment_templates
	`).Scan(&s.Total, &s.Active, &s.Inactive)
	return s, err
}

func (r *Repository) SwapSortOrder(ctx context.Context, fromID, toID uuid.UUID) error {
	tx, err := r.pool.Begin(ctx)
	if err != nil {
		return err
	}
	defer tx.Rollback(ctx)

	var fromOrder, toOrder int
	if err = tx.QueryRow(ctx, `SELECT sort_order FROM comment_templates WHERE id = $1 FOR UPDATE`, fromID).Scan(&fromOrder); err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return ErrNotFound
		}
		return err
	}
	if err = tx.QueryRow(ctx, `SELECT sort_order FROM comment_templates WHERE id = $1 FOR UPDATE`, toID).Scan(&toOrder); err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return ErrNotFound
		}
		return err
	}

	if _, err = tx.Exec(ctx, `
		UPDATE comment_templates SET sort_order = $2, updated_at = NOW() WHERE id = $1
	`, fromID, toOrder); err != nil {
		return err
	}
	if _, err = tx.Exec(ctx, `
		UPDATE comment_templates SET sort_order = $2, updated_at = NOW() WHERE id = $1
	`, toID, fromOrder); err != nil {
		return err
	}
	return tx.Commit(ctx)
}
