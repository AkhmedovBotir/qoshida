package admin

import (
	"context"
	"errors"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

var ErrNotFound = errors.New("admin not found")

type Repository struct {
	pool *pgxpool.Pool
}

func NewRepository(pool *pgxpool.Pool) *Repository {
	return &Repository{pool: pool}
}

const adminCols = `id, first_name, last_name, phone, username, role, password_hash, is_active, created_at`

func scanAdmin(row pgx.Row) (Admin, error) {
	var item Admin
	err := row.Scan(
		&item.ID, &item.FirstName, &item.LastName, &item.Phone, &item.Username,
		&item.Role, &item.PasswordHash, &item.IsActive, &item.CreatedAt,
	)
	if errors.Is(err, pgx.ErrNoRows) {
		return Admin{}, ErrNotFound
	}
	return item, err
}

func (r *Repository) Create(ctx context.Context, item Admin) (Admin, error) {
	return scanAdmin(r.pool.QueryRow(ctx, `
		INSERT INTO admins (first_name, last_name, phone, username, role, password_hash)
		VALUES ($1, $2, $3, $4, $5, $6)
		RETURNING `+adminCols+`
	`, item.FirstName, item.LastName, item.Phone, item.Username, item.Role, item.PasswordHash))
}

func (r *Repository) Update(ctx context.Context, item Admin) (Admin, error) {
	return scanAdmin(r.pool.QueryRow(ctx, `
		UPDATE admins
		SET first_name = $2, last_name = $3, phone = $4, username = $5, password_hash = $6, updated_at = NOW()
		WHERE id = $1 AND role = 'admin'
		RETURNING `+adminCols+`
	`, item.ID, item.FirstName, item.LastName, item.Phone, item.Username, item.PasswordHash))
}

func (r *Repository) Delete(ctx context.Context, id uuid.UUID) error {
	tag, err := r.pool.Exec(ctx, `DELETE FROM admins WHERE id = $1 AND role = 'admin'`, id)
	if err != nil {
		return err
	}
	if tag.RowsAffected() == 0 {
		return ErrNotFound
	}
	return nil
}

func (r *Repository) List(ctx context.Context) ([]Admin, error) {
	rows, err := r.pool.Query(ctx, `
		SELECT `+adminCols+`
		FROM admins
		ORDER BY CASE WHEN role = 'general' THEN 0 ELSE 1 END, created_at ASC
	`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	items := make([]Admin, 0)
	for rows.Next() {
		item, err := scanAdmin(rows)
		if err != nil {
			return nil, err
		}
		items = append(items, item)
	}
	return items, rows.Err()
}

func (r *Repository) FindByID(ctx context.Context, id uuid.UUID) (Admin, error) {
	return scanAdmin(r.pool.QueryRow(ctx, `SELECT `+adminCols+` FROM admins WHERE id = $1`, id))
}

func (r *Repository) FindByUsername(ctx context.Context, username string) (Admin, error) {
	return scanAdmin(r.pool.QueryRow(ctx, `SELECT `+adminCols+` FROM admins WHERE username = $1`, username))
}

func (r *Repository) HasGeneral(ctx context.Context) (bool, error) {
	var exists bool
	err := r.pool.QueryRow(ctx, `SELECT EXISTS(SELECT 1 FROM admins WHERE role = 'general')`).Scan(&exists)
	return exists, err
}

func (r *Repository) Count(ctx context.Context) (int, error) {
	var n int
	err := r.pool.QueryRow(ctx, `SELECT COUNT(*) FROM admins`).Scan(&n)
	return n, err
}
