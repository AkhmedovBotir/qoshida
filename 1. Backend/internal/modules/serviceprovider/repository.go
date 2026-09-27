package serviceprovider

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

var ErrNotFound = errors.New("service provider not found")

type Repository struct {
	pool *pgxpool.Pool
}

func NewRepository(pool *pgxpool.Pool) *Repository {
	return &Repository{pool: pool}
}

const cols = `
	p.id, p.activity_type_id, p.name, p.phone, p.image, p.region_id, p.district_id, p.mfy_id,
	p.password_hash, p.status, COALESCE(ident.status, 'none'), p.created_at, p.updated_at,
	COALESCE(at.name, ''), COALESCE(reg.name, ''), COALESCE(dis.name, ''), COALESCE(mfy.name, '')
`

const fromProviders = `
	service_providers p
	LEFT JOIN activity_types at ON at.id = p.activity_type_id
	LEFT JOIN regions reg ON reg.id = p.region_id
	LEFT JOIN regions dis ON dis.id = p.district_id
	LEFT JOIN regions mfy ON mfy.id = p.mfy_id
	LEFT JOIN service_provider_identifications ident ON ident.provider_id = p.id
`

func scan(row pgx.Row) (Provider, error) {
	var item Provider
	err := row.Scan(
		&item.ID, &item.ActivityTypeID, &item.Name, &item.Phone, &item.Image, &item.RegionID, &item.DistrictID, &item.MFYID,
		&item.PasswordHash, &item.Status, &item.IdentificationStatus, &item.CreatedAt, &item.UpdatedAt,
		&item.ActivityName, &item.RegionName, &item.DistrictName, &item.MFYName,
	)
	if errors.Is(err, pgx.ErrNoRows) {
		return Provider{}, ErrNotFound
	}
	return item, err
}

func (r *Repository) Create(ctx context.Context, item Provider) (Provider, error) {
	_, err := r.pool.Exec(ctx, `
		INSERT INTO service_providers (id, activity_type_id, name, phone, image, region_id, district_id, mfy_id, password_hash, status)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10)
	`, item.ID, item.ActivityTypeID, item.Name, item.Phone, item.Image, item.RegionID, item.DistrictID, item.MFYID, item.PasswordHash, item.Status)
	if err != nil {
		return Provider{}, err
	}
	return r.FindByID(ctx, item.ID)
}

func (r *Repository) Update(ctx context.Context, item Provider) (Provider, error) {
	tag, err := r.pool.Exec(ctx, `
		UPDATE service_providers
		SET activity_type_id = $2, name = $3, phone = $4, image = $5, region_id = $6, district_id = $7, mfy_id = $8,
		    password_hash = $9, status = $10, updated_at = NOW()
		WHERE id = $1
	`, item.ID, item.ActivityTypeID, item.Name, item.Phone, item.Image, item.RegionID, item.DistrictID, item.MFYID, item.PasswordHash, item.Status)
	if err != nil {
		return Provider{}, err
	}
	if tag.RowsAffected() == 0 {
		return Provider{}, ErrNotFound
	}
	return r.FindByID(ctx, item.ID)
}

func (r *Repository) Delete(ctx context.Context, id uuid.UUID) error {
	tag, err := r.pool.Exec(ctx, `DELETE FROM service_providers WHERE id = $1`, id)
	if err != nil {
		return err
	}
	if tag.RowsAffected() == 0 {
		return ErrNotFound
	}
	return nil
}

func (r *Repository) FindByID(ctx context.Context, id uuid.UUID) (Provider, error) {
	return scan(r.pool.QueryRow(ctx, `
		SELECT `+cols+`
		FROM `+fromProviders+`
		WHERE p.id = $1
	`, id))
}

func (r *Repository) FindByPhone(ctx context.Context, phone string) (Provider, error) {
	return scan(r.pool.QueryRow(ctx, `
		SELECT `+cols+`
		FROM `+fromProviders+`
		WHERE p.phone = $1
	`, phone))
}

func (r *Repository) List(ctx context.Context, q ListQuery) ([]Provider, int, error) {
	where := []string{"TRUE"}
	args := make([]any, 0, 8)
	if q.RegionID != nil {
		args = append(args, *q.RegionID)
		where = append(where, fmt.Sprintf("p.region_id = $%d", len(args)))
	}
	if q.DistrictID != nil {
		args = append(args, *q.DistrictID)
		where = append(where, fmt.Sprintf("p.district_id = $%d", len(args)))
	}
	if q.MFYID != nil {
		args = append(args, *q.MFYID)
		where = append(where, fmt.Sprintf("p.mfy_id = $%d", len(args)))
	}
	if q.Query != "" {
		args = append(args, "%"+q.Query+"%")
		where = append(where, fmt.Sprintf("(p.name ILIKE $%d OR p.phone ILIKE $%d)", len(args), len(args)))
	}
	clause := strings.Join(where, " AND ")

	var total int
	if err := r.pool.QueryRow(ctx, `SELECT COUNT(*) FROM `+fromProviders+` WHERE `+clause, args...).Scan(&total); err != nil {
		return nil, 0, err
	}

	args = append(args, q.Limit, (q.Page-1)*q.Limit)
	rows, err := r.pool.Query(ctx, `
		SELECT `+cols+`
		FROM `+fromProviders+`
		WHERE `+clause+`
		ORDER BY p.created_at DESC
		LIMIT $`+fmt.Sprintf("%d", len(args)-1)+` OFFSET $`+fmt.Sprintf("%d", len(args))+`
	`, args...)
	if err != nil {
		return nil, 0, err
	}
	defer rows.Close()

	items := make([]Provider, 0, q.Limit)
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
		FROM service_providers
	`).Scan(&s.Total, &s.Active)
	return s, err
}

func (r *Repository) SaveCode(ctx context.Context, providerID uuid.UUID, phone, purpose, code string, ttl time.Duration) error {
	if _, err := r.pool.Exec(ctx, `
		UPDATE service_provider_verification_codes
		SET used_at = NOW()
		WHERE provider_id = $1 AND purpose = $2 AND used_at IS NULL
	`, providerID, purpose); err != nil {
		return err
	}
	_, err := r.pool.Exec(ctx, `
		INSERT INTO service_provider_verification_codes (provider_id, phone, code_hash, purpose, expires_at)
		VALUES ($1, $2, $3, $4, $5)
	`, providerID, phone, hashCode(phone, purpose, code), purpose, time.Now().Add(ttl))
	return err
}

func (r *Repository) LatestCode(ctx context.Context, phone, purpose string) (uuid.UUID, uuid.UUID, string, time.Time, int, *time.Time, *time.Time, error) {
	var (
		id, providerID uuid.UUID
		hash           string
		expiresAt      time.Time
		attempts       int
		verifiedAt     *time.Time
		usedAt         *time.Time
	)
	err := r.pool.QueryRow(ctx, `
		SELECT id, provider_id, code_hash, expires_at, attempts, verified_at, used_at
		FROM service_provider_verification_codes
		WHERE phone = $1 AND purpose = $2
		ORDER BY created_at DESC
		LIMIT 1
	`, phone, purpose).Scan(&id, &providerID, &hash, &expiresAt, &attempts, &verifiedAt, &usedAt)
	if errors.Is(err, pgx.ErrNoRows) {
		return uuid.Nil, uuid.Nil, "", time.Time{}, 0, nil, nil, ErrNotFound
	}
	return id, providerID, hash, expiresAt, attempts, verifiedAt, usedAt, err
}

func (r *Repository) BumpAttempts(ctx context.Context, id uuid.UUID) error {
	_, err := r.pool.Exec(ctx, `UPDATE service_provider_verification_codes SET attempts = attempts + 1 WHERE id = $1`, id)
	return err
}

func (r *Repository) MarkVerified(ctx context.Context, id uuid.UUID) error {
	_, err := r.pool.Exec(ctx, `UPDATE service_provider_verification_codes SET verified_at = NOW() WHERE id = $1`, id)
	return err
}

func (r *Repository) MarkUsed(ctx context.Context, id uuid.UUID) error {
	_, err := r.pool.Exec(ctx, `UPDATE service_provider_verification_codes SET used_at = NOW() WHERE id = $1`, id)
	return err
}

func (r *Repository) StoreRefresh(ctx context.Context, providerID uuid.UUID, tokenHash string, expiresAt time.Time) error {
	_, err := r.pool.Exec(ctx, `
		INSERT INTO service_provider_refresh_tokens (provider_id, token_hash, expires_at)
		VALUES ($1, $2, $3)
	`, providerID, tokenHash, expiresAt)
	return err
}

func (r *Repository) FindValidRefresh(ctx context.Context, tokenHash string) (uuid.UUID, error) {
	var id uuid.UUID
	err := r.pool.QueryRow(ctx, `
		SELECT provider_id FROM service_provider_refresh_tokens
		WHERE token_hash = $1 AND revoked_at IS NULL AND expires_at > NOW()
	`, tokenHash).Scan(&id)
	if err != nil {
		return uuid.Nil, ErrNotFound
	}
	return id, nil
}

func (r *Repository) RevokeRefresh(ctx context.Context, tokenHash string) error {
	_, err := r.pool.Exec(ctx, `
		UPDATE service_provider_refresh_tokens
		SET revoked_at = NOW()
		WHERE token_hash = $1 AND revoked_at IS NULL
	`, tokenHash)
	return err
}

func hashCode(phone, purpose, code string) string {
	sum := sha256.Sum256([]byte(phone + "|" + purpose + "|" + code))
	return hex.EncodeToString(sum[:])
}
