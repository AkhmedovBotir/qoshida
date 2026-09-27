package manager

import (
	"context"
	"crypto/sha256"
	"encoding/hex"
	"errors"
	"fmt"
	"strings"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

var ErrNotFound = errors.New("manager not found")

type Repository struct {
	pool *pgxpool.Pool
}

func NewRepository(pool *pgxpool.Pool) *Repository {
	return &Repository{pool: pool}
}

const cols = `
	m.id, m.type, m.region_id, m.district_id, m.first_name, m.last_name,
	m.phone, m.username, m.password_hash, m.status, m.created_at, m.updated_at,
	COALESCE(reg.name, ''), COALESCE(dis.name, '')
`

func scan(row pgx.Row) (Manager, error) {
	var item Manager
	err := row.Scan(
		&item.ID, &item.Type, &item.RegionID, &item.DistrictID, &item.FirstName, &item.LastName,
		&item.Phone, &item.Username, &item.PasswordHash, &item.Status, &item.CreatedAt, &item.UpdatedAt,
		&item.RegionName, &item.DistrictName,
	)
	if errors.Is(err, pgx.ErrNoRows) {
		return Manager{}, ErrNotFound
	}
	return item, err
}

func (r *Repository) Create(ctx context.Context, item Manager) (Manager, error) {
	created, err := scan(r.pool.QueryRow(ctx, `
		INSERT INTO managers (type, region_id, district_id, first_name, last_name, phone, username, password_hash, status)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9)
		RETURNING id, type, region_id, district_id, first_name, last_name, phone, username, password_hash, status, created_at, updated_at, '', ''
	`, item.Type, item.RegionID, item.DistrictID, item.FirstName, item.LastName, item.Phone, item.Username, item.PasswordHash, item.Status))
	if err != nil {
		return Manager{}, err
	}
	return r.FindByID(ctx, created.ID)
}

func (r *Repository) Update(ctx context.Context, item Manager) (Manager, error) {
	_, err := scan(r.pool.QueryRow(ctx, `
		UPDATE managers
		SET type = $2, region_id = $3, district_id = $4, first_name = $5, last_name = $6,
		    phone = $7, username = $8, password_hash = $9, status = $10, updated_at = NOW()
		WHERE id = $1
		RETURNING id, type, region_id, district_id, first_name, last_name, phone, username, password_hash, status, created_at, updated_at, '', ''
	`, item.ID, item.Type, item.RegionID, item.DistrictID, item.FirstName, item.LastName, item.Phone, item.Username, item.PasswordHash, item.Status))
	if err != nil {
		return Manager{}, err
	}
	return r.FindByID(ctx, item.ID)
}

func (r *Repository) Delete(ctx context.Context, id uuid.UUID) error {
	tag, err := r.pool.Exec(ctx, `DELETE FROM managers WHERE id = $1`, id)
	if err != nil {
		return err
	}
	if tag.RowsAffected() == 0 {
		return ErrNotFound
	}
	return nil
}

func (r *Repository) FindByID(ctx context.Context, id uuid.UUID) (Manager, error) {
	return scan(r.pool.QueryRow(ctx, `
		SELECT `+cols+`
		FROM managers m
		LEFT JOIN regions reg ON reg.id = m.region_id
		LEFT JOIN regions dis ON dis.id = m.district_id
		WHERE m.id = $1
	`, id))
}

func (r *Repository) FindByPhone(ctx context.Context, phone string) (Manager, error) {
	return scan(r.pool.QueryRow(ctx, `
		SELECT `+cols+`
		FROM managers m
		LEFT JOIN regions reg ON reg.id = m.region_id
		LEFT JOIN regions dis ON dis.id = m.district_id
		WHERE m.phone = $1
	`, phone))
}

func (r *Repository) List(ctx context.Context, q ListQuery) ([]Manager, int, error) {
	where := []string{"TRUE"}
	args := make([]any, 0, 6)
	if q.Type != "" {
		args = append(args, q.Type)
		where = append(where, fmt.Sprintf("m.type = $%d", len(args)))
	}
	if q.RegionID != nil {
		args = append(args, *q.RegionID)
		where = append(where, fmt.Sprintf("m.region_id = $%d", len(args)))
	}
	if q.Query != "" {
		args = append(args, "%"+q.Query+"%")
		where = append(where, fmt.Sprintf(
			"(m.first_name ILIKE $%d OR m.last_name ILIKE $%d OR m.phone ILIKE $%d OR m.username ILIKE $%d)",
			len(args), len(args), len(args), len(args),
		))
	}
	clause := strings.Join(where, " AND ")

	var total int
	if err := r.pool.QueryRow(ctx, `SELECT COUNT(*) FROM managers m WHERE `+clause, args...).Scan(&total); err != nil {
		return nil, 0, err
	}

	args = append(args, q.Limit, (q.Page-1)*q.Limit)
	rows, err := r.pool.Query(ctx, `
		SELECT `+cols+`
		FROM managers m
		LEFT JOIN regions reg ON reg.id = m.region_id
		LEFT JOIN regions dis ON dis.id = m.district_id
		WHERE `+clause+`
		ORDER BY m.created_at DESC
		LIMIT $`+fmt.Sprintf("%d", len(args)-1)+` OFFSET $`+fmt.Sprintf("%d", len(args))+`
	`, args...)
	if err != nil {
		return nil, 0, err
	}
	defer rows.Close()

	items := make([]Manager, 0, q.Limit)
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
			COUNT(*) FILTER (WHERE status = 'active')
		FROM managers
	`).Scan(&s.Total, &s.Region, &s.District, &s.Active)
	return s, err
}

func (r *Repository) SaveCode(ctx context.Context, managerID uuid.UUID, phone, purpose, code string, ttl time.Duration) error {
	if _, err := r.pool.Exec(ctx, `
		UPDATE manager_verification_codes
		SET used_at = NOW()
		WHERE manager_id = $1 AND purpose = $2 AND used_at IS NULL
	`, managerID, purpose); err != nil {
		return err
	}
	_, err := r.pool.Exec(ctx, `
		INSERT INTO manager_verification_codes (manager_id, phone, code_hash, purpose, expires_at)
		VALUES ($1, $2, $3, $4, $5)
	`, managerID, phone, hashCode(phone, purpose, code), purpose, time.Now().Add(ttl))
	return err
}

func (r *Repository) LatestCode(ctx context.Context, phone, purpose string) (uuid.UUID, uuid.UUID, string, time.Time, int, *time.Time, *time.Time, error) {
	var (
		id, managerID uuid.UUID
		hash          string
		expiresAt     time.Time
		attempts      int
		verifiedAt    *time.Time
		usedAt        *time.Time
	)
	err := r.pool.QueryRow(ctx, `
		SELECT id, manager_id, code_hash, expires_at, attempts, verified_at, used_at
		FROM manager_verification_codes
		WHERE phone = $1 AND purpose = $2
		ORDER BY created_at DESC
		LIMIT 1
	`, phone, purpose).Scan(&id, &managerID, &hash, &expiresAt, &attempts, &verifiedAt, &usedAt)
	if errors.Is(err, pgx.ErrNoRows) {
		return uuid.Nil, uuid.Nil, "", time.Time{}, 0, nil, nil, ErrNotFound
	}
	return id, managerID, hash, expiresAt, attempts, verifiedAt, usedAt, err
}

func (r *Repository) BumpAttempts(ctx context.Context, id uuid.UUID) error {
	_, err := r.pool.Exec(ctx, `UPDATE manager_verification_codes SET attempts = attempts + 1 WHERE id = $1`, id)
	return err
}

func (r *Repository) MarkVerified(ctx context.Context, id uuid.UUID) error {
	_, err := r.pool.Exec(ctx, `UPDATE manager_verification_codes SET verified_at = NOW() WHERE id = $1`, id)
	return err
}

func (r *Repository) MarkUsed(ctx context.Context, id uuid.UUID) error {
	_, err := r.pool.Exec(ctx, `UPDATE manager_verification_codes SET used_at = NOW() WHERE id = $1`, id)
	return err
}

func (r *Repository) StoreRefresh(ctx context.Context, managerID uuid.UUID, tokenHash string, expiresAt time.Time) error {
	_, err := r.pool.Exec(ctx, `
		INSERT INTO manager_refresh_tokens (manager_id, token_hash, expires_at)
		VALUES ($1, $2, $3)
	`, managerID, tokenHash, expiresAt)
	return err
}

func (r *Repository) FindValidRefresh(ctx context.Context, tokenHash string) (uuid.UUID, error) {
	var id uuid.UUID
	err := r.pool.QueryRow(ctx, `
		SELECT manager_id FROM manager_refresh_tokens
		WHERE token_hash = $1 AND revoked_at IS NULL AND expires_at > NOW()
	`, tokenHash).Scan(&id)
	if err != nil {
		return uuid.Nil, ErrNotFound
	}
	return id, nil
}

func (r *Repository) RevokeRefresh(ctx context.Context, tokenHash string) error {
	_, err := r.pool.Exec(ctx, `
		UPDATE manager_refresh_tokens
		SET revoked_at = NOW()
		WHERE token_hash = $1 AND revoked_at IS NULL
	`, tokenHash)
	return err
}

func hashCode(phone, purpose, code string) string {
	sum := sha256.Sum256([]byte(phone + "|" + purpose + "|" + code))
	return hex.EncodeToString(sum[:])
}
