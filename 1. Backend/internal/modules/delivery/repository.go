package delivery

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

var ErrNotFound = errors.New("delivery not found")

type Repository struct {
	pool *pgxpool.Pool
}

func NewRepository(pool *pgxpool.Pool) *Repository {
	return &Repository{pool: pool}
}

const cols = `
	s.id, s.shop_id, s.first_name, s.last_name, s.phone, s.password_hash, s.status,
	s.created_at, s.updated_at,
	COALESCE(sh.name, ''), sh.region_id, sh.district_id, sh.mfy_id,
	COALESCE(reg.name, ''), COALESCE(dis.name, ''), COALESCE(mfy.name, '')
`

func scan(row pgx.Row) (Delivery, error) {
	var item Delivery
	err := row.Scan(
		&item.ID, &item.ShopID, &item.FirstName, &item.LastName, &item.Phone, &item.PasswordHash, &item.Status,
		&item.CreatedAt, &item.UpdatedAt,
		&item.ShopName, &item.RegionID, &item.DistrictID, &item.MFYID,
		&item.RegionName, &item.DistrictName, &item.MFYName,
	)
	if errors.Is(err, pgx.ErrNoRows) {
		return Delivery{}, ErrNotFound
	}
	return item, err
}

func (r *Repository) Create(ctx context.Context, item Delivery) (Delivery, error) {
	var id uuid.UUID
	err := r.pool.QueryRow(ctx, `
		INSERT INTO deliveries (shop_id, first_name, last_name, phone, password_hash, status)
		VALUES ($1, $2, $3, $4, $5, $6)
		RETURNING id
	`, item.ShopID, item.FirstName, item.LastName, item.Phone, item.PasswordHash, item.Status).Scan(&id)
	if err != nil {
		return Delivery{}, err
	}
	return r.FindByID(ctx, id)
}

func (r *Repository) Update(ctx context.Context, item Delivery) (Delivery, error) {
	tag, err := r.pool.Exec(ctx, `
		UPDATE deliveries
		SET shop_id = $2, first_name = $3, last_name = $4, phone = $5,
		    password_hash = $6, status = $7, updated_at = NOW()
		WHERE id = $1
	`, item.ID, item.ShopID, item.FirstName, item.LastName, item.Phone, item.PasswordHash, item.Status)
	if err != nil {
		return Delivery{}, err
	}
	if tag.RowsAffected() == 0 {
		return Delivery{}, ErrNotFound
	}
	return r.FindByID(ctx, item.ID)
}

func (r *Repository) Delete(ctx context.Context, id uuid.UUID) error {
	tag, err := r.pool.Exec(ctx, `DELETE FROM deliveries WHERE id = $1`, id)
	if err != nil {
		return err
	}
	if tag.RowsAffected() == 0 {
		return ErrNotFound
	}
	return nil
}

func (r *Repository) FindByID(ctx context.Context, id uuid.UUID) (Delivery, error) {
	return scan(r.pool.QueryRow(ctx, `
		SELECT `+cols+`
		FROM deliveries s
		JOIN local_shops sh ON sh.id = s.shop_id
		LEFT JOIN regions reg ON reg.id = sh.region_id
		LEFT JOIN regions dis ON dis.id = sh.district_id
		LEFT JOIN regions mfy ON mfy.id = sh.mfy_id
		WHERE s.id = $1
	`, id))
}

func (r *Repository) FindByPhone(ctx context.Context, phone string) (Delivery, error) {
	return scan(r.pool.QueryRow(ctx, `
		SELECT `+cols+`
		FROM deliveries s
		JOIN local_shops sh ON sh.id = s.shop_id
		LEFT JOIN regions reg ON reg.id = sh.region_id
		LEFT JOIN regions dis ON dis.id = sh.district_id
		LEFT JOIN regions mfy ON mfy.id = sh.mfy_id
		WHERE s.phone = $1
	`, phone))
}

func (r *Repository) FindShop(ctx context.Context, id uuid.UUID) (ShopRef, error) {
	var shop ShopRef
	err := r.pool.QueryRow(ctx, `
		SELECT id, name, region_id, district_id, mfy_id, status
		FROM local_shops
		WHERE id = $1
	`, id).Scan(&shop.ID, &shop.Name, &shop.RegionID, &shop.DistrictID, &shop.MFYID, &shop.Status)
	if errors.Is(err, pgx.ErrNoRows) {
		return ShopRef{}, ErrNotFound
	}
	return shop, err
}

func (r *Repository) List(ctx context.Context, q ListQuery) ([]Delivery, int, error) {
	where := []string{"TRUE"}
	args := make([]any, 0, 8)
	if q.ShopID != nil {
		args = append(args, *q.ShopID)
		where = append(where, fmt.Sprintf("s.shop_id = $%d", len(args)))
	}
	if q.RegionID != nil {
		args = append(args, *q.RegionID)
		where = append(where, fmt.Sprintf("sh.region_id = $%d", len(args)))
	}
	if q.DistrictID != nil {
		args = append(args, *q.DistrictID)
		where = append(where, fmt.Sprintf("sh.district_id = $%d", len(args)))
	}
	if q.MFYID != nil {
		args = append(args, *q.MFYID)
		where = append(where, fmt.Sprintf("sh.mfy_id = $%d", len(args)))
	}
	if q.Query != "" {
		args = append(args, "%"+q.Query+"%")
		where = append(where, fmt.Sprintf(
			"(s.first_name ILIKE $%d OR s.last_name ILIKE $%d OR s.phone ILIKE $%d OR sh.name ILIKE $%d)",
			len(args), len(args), len(args), len(args),
		))
	}
	clause := strings.Join(where, " AND ")

	var total int
	if err := r.pool.QueryRow(ctx, `
		SELECT COUNT(*)
		FROM deliveries s
		JOIN local_shops sh ON sh.id = s.shop_id
		WHERE `+clause, args...).Scan(&total); err != nil {
		return nil, 0, err
	}

	args = append(args, q.Limit, (q.Page-1)*q.Limit)
	rows, err := r.pool.Query(ctx, `
		SELECT `+cols+`
		FROM deliveries s
		JOIN local_shops sh ON sh.id = s.shop_id
		LEFT JOIN regions reg ON reg.id = sh.region_id
		LEFT JOIN regions dis ON dis.id = sh.district_id
		LEFT JOIN regions mfy ON mfy.id = sh.mfy_id
		WHERE `+clause+`
		ORDER BY s.created_at DESC
		LIMIT $`+fmt.Sprintf("%d", len(args)-1)+` OFFSET $`+fmt.Sprintf("%d", len(args))+`
	`, args...)
	if err != nil {
		return nil, 0, err
	}
	defer rows.Close()

	items := make([]Delivery, 0, q.Limit)
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
		FROM deliveries
	`).Scan(&s.Total, &s.Active)
	return s, err
}

func (r *Repository) SaveCode(ctx context.Context, deliveryID uuid.UUID, phone, purpose, code string, ttl time.Duration) error {
	if _, err := r.pool.Exec(ctx, `
		UPDATE delivery_verification_codes
		SET used_at = NOW()
		WHERE delivery_id = $1 AND purpose = $2 AND used_at IS NULL
	`, deliveryID, purpose); err != nil {
		return err
	}
	_, err := r.pool.Exec(ctx, `
		INSERT INTO delivery_verification_codes (delivery_id, phone, code_hash, purpose, expires_at)
		VALUES ($1, $2, $3, $4, $5)
	`, deliveryID, phone, hashCode(phone, purpose, code), purpose, time.Now().Add(ttl))
	return err
}

func (r *Repository) LatestCode(ctx context.Context, phone, purpose string) (uuid.UUID, uuid.UUID, string, time.Time, int, *time.Time, *time.Time, error) {
	var (
		id, deliveryID uuid.UUID
		hash         string
		expiresAt    time.Time
		attempts     int
		verifiedAt   *time.Time
		usedAt       *time.Time
	)
	err := r.pool.QueryRow(ctx, `
		SELECT id, delivery_id, code_hash, expires_at, attempts, verified_at, used_at
		FROM delivery_verification_codes
		WHERE phone = $1 AND purpose = $2
		ORDER BY created_at DESC
		LIMIT 1
	`, phone, purpose).Scan(&id, &deliveryID, &hash, &expiresAt, &attempts, &verifiedAt, &usedAt)
	if errors.Is(err, pgx.ErrNoRows) {
		return uuid.Nil, uuid.Nil, "", time.Time{}, 0, nil, nil, ErrNotFound
	}
	return id, deliveryID, hash, expiresAt, attempts, verifiedAt, usedAt, err
}

func (r *Repository) BumpAttempts(ctx context.Context, id uuid.UUID) error {
	_, err := r.pool.Exec(ctx, `UPDATE delivery_verification_codes SET attempts = attempts + 1 WHERE id = $1`, id)
	return err
}

func (r *Repository) MarkVerified(ctx context.Context, id uuid.UUID) error {
	_, err := r.pool.Exec(ctx, `UPDATE delivery_verification_codes SET verified_at = NOW() WHERE id = $1`, id)
	return err
}

func (r *Repository) MarkUsed(ctx context.Context, id uuid.UUID) error {
	_, err := r.pool.Exec(ctx, `UPDATE delivery_verification_codes SET used_at = NOW() WHERE id = $1`, id)
	return err
}

func (r *Repository) StoreRefresh(ctx context.Context, deliveryID uuid.UUID, tokenHash string, expiresAt time.Time) error {
	_, err := r.pool.Exec(ctx, `
		INSERT INTO delivery_refresh_tokens (delivery_id, token_hash, expires_at)
		VALUES ($1, $2, $3)
	`, deliveryID, tokenHash, expiresAt)
	return err
}

func (r *Repository) FindValidRefresh(ctx context.Context, tokenHash string) (uuid.UUID, error) {
	var id uuid.UUID
	err := r.pool.QueryRow(ctx, `
		SELECT delivery_id FROM delivery_refresh_tokens
		WHERE token_hash = $1 AND revoked_at IS NULL AND expires_at > NOW()
	`, tokenHash).Scan(&id)
	if err != nil {
		return uuid.Nil, ErrNotFound
	}
	return id, nil
}

func (r *Repository) RevokeRefresh(ctx context.Context, tokenHash string) error {
	_, err := r.pool.Exec(ctx, `
		UPDATE delivery_refresh_tokens
		SET revoked_at = NOW()
		WHERE token_hash = $1 AND revoked_at IS NULL
	`, tokenHash)
	return err
}

func hashCode(phone, purpose, code string) string {
	sum := sha256.Sum256([]byte(phone + "|" + purpose + "|" + code))
	return hex.EncodeToString(sum[:])
}
