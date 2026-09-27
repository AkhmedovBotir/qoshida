package customer

import (
	"context"
	"crypto/sha256"
	"database/sql"
	"encoding/hex"
	"errors"
	"strings"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

var ErrNotFound = errors.New("customer not found")

type Repository struct {
	pool *pgxpool.Pool
}

func NewRepository(pool *pgxpool.Pool) *Repository {
	return &Repository{pool: pool}
}

const cols = `
	c.id, c.name, c.first_name, c.last_name, c.birth_date, c.phone, c.password_hash, c.region_id, c.district_id, c.mfy_id, c.address, c.avatar, c.status,
	c.created_at, c.updated_at, COALESCE(reg.name, ''), COALESCE(dis.name, ''), COALESCE(mfy.name, '')
`

func scan(row pgx.Row) (Customer, error) {
	var item Customer
	var birth sql.NullTime
	err := row.Scan(
		&item.ID, &item.Name, &item.FirstName, &item.LastName, &birth, &item.Phone, &item.PasswordHash, &item.RegionID, &item.DistrictID, &item.MFYID,
		&item.Address, &item.Avatar, &item.Status, &item.CreatedAt, &item.UpdatedAt, &item.RegionName, &item.DistrictName, &item.MFYName,
	)
	if errors.Is(err, pgx.ErrNoRows) {
		return Customer{}, ErrNotFound
	}
	if err != nil {
		return Customer{}, err
	}
	item.BirthDate = scanBirth(birth)
	return item, nil
}

func (r *Repository) Create(ctx context.Context, item Customer) (Customer, error) {
	_, err := r.pool.Exec(ctx, `
		INSERT INTO customers (id, name, first_name, last_name, birth_date, phone, password_hash, region_id, district_id, mfy_id, address, avatar, status)
		VALUES ($1,$2,$3,$4,$5,$6,$7,$8,$9,$10,$11,$12,$13)
	`, item.ID, item.Name, item.FirstName, item.LastName, nullTime(item.BirthDate), item.Phone, item.PasswordHash, item.RegionID, item.DistrictID, item.MFYID, item.Address, item.Avatar, item.Status)
	if err != nil {
		return Customer{}, err
	}
	return r.FindByID(ctx, item.ID)
}

func (r *Repository) Update(ctx context.Context, item Customer) (Customer, error) {
	tag, err := r.pool.Exec(ctx, `
		UPDATE customers
		SET name = $2, first_name = $3, last_name = $4, birth_date = $5, password_hash = $6, region_id = $7, district_id = $8, mfy_id = $9, address = $10, avatar = $11, status = $12, updated_at = NOW()
		WHERE id = $1
	`, item.ID, item.Name, item.FirstName, item.LastName, nullTime(item.BirthDate), item.PasswordHash, item.RegionID, item.DistrictID, item.MFYID, item.Address, item.Avatar, item.Status)
	if err != nil {
		return Customer{}, err
	}
	if tag.RowsAffected() == 0 {
		return Customer{}, ErrNotFound
	}
	return r.FindByID(ctx, item.ID)
}

func (r *Repository) FindByID(ctx context.Context, id uuid.UUID) (Customer, error) {
	return scan(r.pool.QueryRow(ctx, `
		SELECT `+cols+`
		FROM customers c
		LEFT JOIN regions reg ON reg.id = c.region_id
		LEFT JOIN regions dis ON dis.id = c.district_id
		LEFT JOIN regions mfy ON mfy.id = c.mfy_id
		WHERE c.id = $1
	`, id))
}

func (r *Repository) FindByPhone(ctx context.Context, phone string) (Customer, error) {
	return scan(r.pool.QueryRow(ctx, `
		SELECT `+cols+`
		FROM customers c
		LEFT JOIN regions reg ON reg.id = c.region_id
		LEFT JOIN regions dis ON dis.id = c.district_id
		LEFT JOIN regions mfy ON mfy.id = c.mfy_id
		WHERE c.phone = $1
	`, phone))
}

func (r *Repository) SaveCode(ctx context.Context, customerID *uuid.UUID, phone, purpose, code string, ttl time.Duration) error {
	if _, err := r.pool.Exec(ctx, `
		UPDATE customer_verification_codes
		SET used_at = NOW()
		WHERE phone = $1 AND purpose = $2 AND used_at IS NULL
	`, phone, purpose); err != nil {
		return err
	}
	_, err := r.pool.Exec(ctx, `
		INSERT INTO customer_verification_codes (customer_id, phone, code_hash, purpose, expires_at)
		VALUES ($1, $2, $3, $4, $5)
	`, customerID, phone, hashCode(phone, purpose, code), purpose, time.Now().Add(ttl))
	return err
}

func (r *Repository) LatestCode(ctx context.Context, phone, purpose string) (uuid.UUID, *uuid.UUID, string, time.Time, int, *time.Time, *time.Time, error) {
	var (
		id         uuid.UUID
		customerID *uuid.UUID
		hash       string
		expiresAt  time.Time
		attempts   int
		verifiedAt *time.Time
		usedAt     *time.Time
	)
	err := r.pool.QueryRow(ctx, `
		SELECT id, customer_id, code_hash, expires_at, attempts, verified_at, used_at
		FROM customer_verification_codes
		WHERE phone = $1 AND purpose = $2
		ORDER BY created_at DESC
		LIMIT 1
	`, phone, purpose).Scan(&id, &customerID, &hash, &expiresAt, &attempts, &verifiedAt, &usedAt)
	if errors.Is(err, pgx.ErrNoRows) {
		return uuid.Nil, nil, "", time.Time{}, 0, nil, nil, ErrNotFound
	}
	return id, customerID, hash, expiresAt, attempts, verifiedAt, usedAt, err
}

func (r *Repository) BumpAttempts(ctx context.Context, id uuid.UUID) error {
	_, err := r.pool.Exec(ctx, `UPDATE customer_verification_codes SET attempts = attempts + 1 WHERE id = $1`, id)
	return err
}

func (r *Repository) MarkVerified(ctx context.Context, id uuid.UUID) error {
	_, err := r.pool.Exec(ctx, `UPDATE customer_verification_codes SET verified_at = NOW() WHERE id = $1`, id)
	return err
}

func (r *Repository) MarkUsed(ctx context.Context, id uuid.UUID) error {
	_, err := r.pool.Exec(ctx, `UPDATE customer_verification_codes SET used_at = NOW() WHERE id = $1`, id)
	return err
}

func (r *Repository) StoreRefresh(ctx context.Context, customerID uuid.UUID, tokenHash string, expiresAt time.Time) error {
	_, err := r.pool.Exec(ctx, `
		INSERT INTO customer_refresh_tokens (customer_id, token_hash, expires_at)
		VALUES ($1, $2, $3)
	`, customerID, tokenHash, expiresAt)
	return err
}

func (r *Repository) FindValidRefresh(ctx context.Context, tokenHash string) (uuid.UUID, error) {
	var id uuid.UUID
	err := r.pool.QueryRow(ctx, `
		SELECT customer_id FROM customer_refresh_tokens
		WHERE token_hash = $1 AND revoked_at IS NULL AND expires_at > NOW()
	`, tokenHash).Scan(&id)
	if err != nil {
		return uuid.Nil, ErrNotFound
	}
	return id, nil
}

func (r *Repository) RevokeRefresh(ctx context.Context, tokenHash string) error {
	_, err := r.pool.Exec(ctx, `
		UPDATE customer_refresh_tokens SET revoked_at = NOW()
		WHERE token_hash = $1 AND revoked_at IS NULL
	`, tokenHash)
	return err
}

func hashCode(phone, purpose, code string) string {
	sum := sha256.Sum256([]byte(phone + "|" + purpose + "|" + strings.TrimSpace(code)))
	return hex.EncodeToString(sum[:])
}
