package shopdirector

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

var ErrNotFound = errors.New("shop director not found")

type Repository struct {
	pool *pgxpool.Pool
}

func NewRepository(pool *pgxpool.Pool) *Repository {
	return &Repository{pool: pool}
}

const cols = `
	d.id, d.region_id, d.district_id, d.mfy_id, d.first_name, d.last_name,
	d.phone, d.password_hash, d.status, d.created_at, d.updated_at,
	COALESCE(reg.name, ''), COALESCE(dis.name, ''), COALESCE(mfy.name, '')
`

func scan(row pgx.Row) (Director, error) {
	var item Director
	err := row.Scan(
		&item.ID, &item.RegionID, &item.DistrictID, &item.MFYID, &item.FirstName, &item.LastName,
		&item.Phone, &item.PasswordHash, &item.Status, &item.CreatedAt, &item.UpdatedAt,
		&item.RegionName, &item.DistrictName, &item.MFYName,
	)
	if errors.Is(err, pgx.ErrNoRows) {
		return Director{}, ErrNotFound
	}
	return item, err
}

func (r *Repository) Create(ctx context.Context, item Director) (Director, error) {
	created, err := scan(r.pool.QueryRow(ctx, `
		INSERT INTO shop_directors (region_id, district_id, mfy_id, first_name, last_name, phone, password_hash, status)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8)
		RETURNING id, region_id, district_id, mfy_id, first_name, last_name, phone, password_hash, status, created_at, updated_at, '', '', ''
	`, item.RegionID, item.DistrictID, item.MFYID, item.FirstName, item.LastName, item.Phone, item.PasswordHash, item.Status))
	if err != nil {
		return Director{}, err
	}
	return r.FindByID(ctx, created.ID)
}

func (r *Repository) Update(ctx context.Context, item Director) (Director, error) {
	_, err := scan(r.pool.QueryRow(ctx, `
		UPDATE shop_directors
		SET region_id = $2, district_id = $3, mfy_id = $4, first_name = $5, last_name = $6,
		    phone = $7, password_hash = $8, status = $9, updated_at = NOW()
		WHERE id = $1
		RETURNING id, region_id, district_id, mfy_id, first_name, last_name, phone, password_hash, status, created_at, updated_at, '', '', ''
	`, item.ID, item.RegionID, item.DistrictID, item.MFYID, item.FirstName, item.LastName, item.Phone, item.PasswordHash, item.Status))
	if err != nil {
		return Director{}, err
	}
	return r.FindByID(ctx, item.ID)
}

func (r *Repository) Delete(ctx context.Context, id uuid.UUID) error {
	tag, err := r.pool.Exec(ctx, `DELETE FROM shop_directors WHERE id = $1`, id)
	if err != nil {
		return err
	}
	if tag.RowsAffected() == 0 {
		return ErrNotFound
	}
	return nil
}

func (r *Repository) FindByID(ctx context.Context, id uuid.UUID) (Director, error) {
	return scan(r.pool.QueryRow(ctx, `
		SELECT `+cols+`
		FROM shop_directors d
		LEFT JOIN regions reg ON reg.id = d.region_id
		LEFT JOIN regions dis ON dis.id = d.district_id
		LEFT JOIN regions mfy ON mfy.id = d.mfy_id
		WHERE d.id = $1
	`, id))
}

func (r *Repository) FindByPhone(ctx context.Context, phone string) (Director, error) {
	return scan(r.pool.QueryRow(ctx, `
		SELECT `+cols+`
		FROM shop_directors d
		LEFT JOIN regions reg ON reg.id = d.region_id
		LEFT JOIN regions dis ON dis.id = d.district_id
		LEFT JOIN regions mfy ON mfy.id = d.mfy_id
		WHERE d.phone = $1
	`, phone))
}

func (r *Repository) List(ctx context.Context, q ListQuery) ([]Director, int, error) {
	where := []string{"TRUE"}
	args := make([]any, 0, 8)
	if q.RegionID != nil {
		args = append(args, *q.RegionID)
		where = append(where, fmt.Sprintf("d.region_id = $%d", len(args)))
	}
	if q.DistrictID != nil {
		args = append(args, *q.DistrictID)
		where = append(where, fmt.Sprintf("d.district_id = $%d", len(args)))
	}
	if q.Query != "" {
		args = append(args, "%"+q.Query+"%")
		where = append(where, fmt.Sprintf(
			"(d.first_name ILIKE $%d OR d.last_name ILIKE $%d OR d.phone ILIKE $%d)",
			len(args), len(args), len(args),
		))
	}
	clause := strings.Join(where, " AND ")

	var total int
	if err := r.pool.QueryRow(ctx, `SELECT COUNT(*) FROM shop_directors d WHERE `+clause, args...).Scan(&total); err != nil {
		return nil, 0, err
	}

	args = append(args, q.Limit, (q.Page-1)*q.Limit)
	rows, err := r.pool.Query(ctx, `
		SELECT `+cols+`
		FROM shop_directors d
		LEFT JOIN regions reg ON reg.id = d.region_id
		LEFT JOIN regions dis ON dis.id = d.district_id
		LEFT JOIN regions mfy ON mfy.id = d.mfy_id
		WHERE `+clause+`
		ORDER BY d.created_at DESC
		LIMIT $`+fmt.Sprintf("%d", len(args)-1)+` OFFSET $`+fmt.Sprintf("%d", len(args))+`
	`, args...)
	if err != nil {
		return nil, 0, err
	}
	defer rows.Close()

	items := make([]Director, 0, q.Limit)
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
		SELECT COUNT(*), COUNT(*) FILTER (WHERE status = 'active')
		FROM shop_directors
	`).Scan(&s.Total, &s.Active)
	return s, err
}

func (r *Repository) SaveCode(ctx context.Context, directorID uuid.UUID, phone, purpose, code string, ttl time.Duration) error {
	if _, err := r.pool.Exec(ctx, `
		UPDATE shop_director_verification_codes
		SET used_at = NOW()
		WHERE shop_director_id = $1 AND purpose = $2 AND used_at IS NULL
	`, directorID, purpose); err != nil {
		return err
	}
	_, err := r.pool.Exec(ctx, `
		INSERT INTO shop_director_verification_codes (shop_director_id, phone, code_hash, purpose, expires_at)
		VALUES ($1, $2, $3, $4, $5)
	`, directorID, phone, hashCode(phone, purpose, code), purpose, time.Now().Add(ttl))
	return err
}

func (r *Repository) LatestCode(ctx context.Context, phone, purpose string) (uuid.UUID, uuid.UUID, string, time.Time, int, *time.Time, *time.Time, error) {
	var (
		id, directorID uuid.UUID
		hash           string
		expiresAt      time.Time
		attempts       int
		verifiedAt     *time.Time
		usedAt         *time.Time
	)
	err := r.pool.QueryRow(ctx, `
		SELECT id, shop_director_id, code_hash, expires_at, attempts, verified_at, used_at
		FROM shop_director_verification_codes
		WHERE phone = $1 AND purpose = $2
		ORDER BY created_at DESC
		LIMIT 1
	`, phone, purpose).Scan(&id, &directorID, &hash, &expiresAt, &attempts, &verifiedAt, &usedAt)
	if errors.Is(err, pgx.ErrNoRows) {
		return uuid.Nil, uuid.Nil, "", time.Time{}, 0, nil, nil, ErrNotFound
	}
	return id, directorID, hash, expiresAt, attempts, verifiedAt, usedAt, err
}

func (r *Repository) BumpAttempts(ctx context.Context, id uuid.UUID) error {
	_, err := r.pool.Exec(ctx, `UPDATE shop_director_verification_codes SET attempts = attempts + 1 WHERE id = $1`, id)
	return err
}

func (r *Repository) MarkVerified(ctx context.Context, id uuid.UUID) error {
	_, err := r.pool.Exec(ctx, `UPDATE shop_director_verification_codes SET verified_at = NOW() WHERE id = $1`, id)
	return err
}

func (r *Repository) MarkUsed(ctx context.Context, id uuid.UUID) error {
	_, err := r.pool.Exec(ctx, `UPDATE shop_director_verification_codes SET used_at = NOW() WHERE id = $1`, id)
	return err
}

func (r *Repository) StoreRefresh(ctx context.Context, directorID uuid.UUID, tokenHash string, expiresAt time.Time) error {
	_, err := r.pool.Exec(ctx, `
		INSERT INTO shop_director_refresh_tokens (shop_director_id, token_hash, expires_at)
		VALUES ($1, $2, $3)
	`, directorID, tokenHash, expiresAt)
	return err
}

func (r *Repository) FindValidRefresh(ctx context.Context, tokenHash string) (uuid.UUID, error) {
	var id uuid.UUID
	err := r.pool.QueryRow(ctx, `
		SELECT shop_director_id FROM shop_director_refresh_tokens
		WHERE token_hash = $1 AND revoked_at IS NULL AND expires_at > NOW()
	`, tokenHash).Scan(&id)
	if err != nil {
		return uuid.Nil, ErrNotFound
	}
	return id, nil
}

func (r *Repository) RevokeRefresh(ctx context.Context, tokenHash string) error {
	_, err := r.pool.Exec(ctx, `
		UPDATE shop_director_refresh_tokens
		SET revoked_at = NOW()
		WHERE token_hash = $1 AND revoked_at IS NULL
	`, tokenHash)
	return err
}

func hashCode(phone, purpose, code string) string {
	sum := sha256.Sum256([]byte(phone + "|" + purpose + "|" + code))
	return hex.EncodeToString(sum[:])
}
