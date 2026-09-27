package kontragent

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

var ErrNotFound = errors.New("kontragent not found")

type Repository struct {
	pool *pgxpool.Pool
}

func NewRepository(pool *pgxpool.Pool) *Repository {
	return &Repository{pool: pool}
}

const cols = `
	k.id, k.activity_type_id, k.name, k.inn, k.region_id, k.district_id, k.mfy_id,
	k.phone, k.logo, k.password_hash, k.status, k.created_at, k.updated_at,
	COALESCE(at.name, ''), COALESCE(reg.name, ''), COALESCE(dis.name, ''), COALESCE(mfy.name, '')
`

func scan(row pgx.Row) (Kontragent, error) {
	var item Kontragent
	err := row.Scan(
		&item.ID, &item.ActivityTypeID, &item.Name, &item.INN, &item.RegionID, &item.DistrictID, &item.MFYID,
		&item.Phone, &item.Logo, &item.PasswordHash, &item.Status, &item.CreatedAt, &item.UpdatedAt,
		&item.ActivityName, &item.RegionName, &item.DistrictName, &item.MFYName,
	)
	if errors.Is(err, pgx.ErrNoRows) {
		return Kontragent{}, ErrNotFound
	}
	return item, err
}

func (r *Repository) Create(ctx context.Context, item Kontragent) (Kontragent, error) {
	_, err := scan(r.pool.QueryRow(ctx, `
		INSERT INTO kontragents (id, activity_type_id, name, inn, region_id, district_id, mfy_id, phone, logo, password_hash, status)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11)
		RETURNING id, activity_type_id, name, inn, region_id, district_id, mfy_id, phone, logo, password_hash, status, created_at, updated_at, '', '', '', ''
	`, item.ID, item.ActivityTypeID, item.Name, item.INN, item.RegionID, item.DistrictID, item.MFYID, item.Phone, item.Logo, item.PasswordHash, item.Status))
	if err != nil {
		return Kontragent{}, err
	}
	return r.FindByID(ctx, item.ID)
}

func (r *Repository) Update(ctx context.Context, item Kontragent) (Kontragent, error) {
	_, err := scan(r.pool.QueryRow(ctx, `
		UPDATE kontragents
		SET activity_type_id = $2, name = $3, inn = $4, region_id = $5, district_id = $6, mfy_id = $7,
		    phone = $8, logo = $9, password_hash = $10, status = $11, updated_at = NOW()
		WHERE id = $1
		RETURNING id, activity_type_id, name, inn, region_id, district_id, mfy_id, phone, logo, password_hash, status, created_at, updated_at, '', '', '', ''
	`, item.ID, item.ActivityTypeID, item.Name, item.INN, item.RegionID, item.DistrictID, item.MFYID, item.Phone, item.Logo, item.PasswordHash, item.Status))
	if err != nil {
		return Kontragent{}, err
	}
	return r.FindByID(ctx, item.ID)
}

func (r *Repository) Delete(ctx context.Context, id uuid.UUID) error {
	tag, err := r.pool.Exec(ctx, `DELETE FROM kontragents WHERE id = $1`, id)
	if err != nil {
		return err
	}
	if tag.RowsAffected() == 0 {
		return ErrNotFound
	}
	return nil
}

func (r *Repository) FindByID(ctx context.Context, id uuid.UUID) (Kontragent, error) {
	return scan(r.pool.QueryRow(ctx, `
		SELECT `+cols+`
		FROM kontragents k
		LEFT JOIN activity_types at ON at.id = k.activity_type_id
		LEFT JOIN regions reg ON reg.id = k.region_id
		LEFT JOIN regions dis ON dis.id = k.district_id
		LEFT JOIN regions mfy ON mfy.id = k.mfy_id
		WHERE k.id = $1
	`, id))
}

func (r *Repository) FindByPhone(ctx context.Context, phone string) (Kontragent, error) {
	return scan(r.pool.QueryRow(ctx, `
		SELECT `+cols+`
		FROM kontragents k
		LEFT JOIN activity_types at ON at.id = k.activity_type_id
		LEFT JOIN regions reg ON reg.id = k.region_id
		LEFT JOIN regions dis ON dis.id = k.district_id
		LEFT JOIN regions mfy ON mfy.id = k.mfy_id
		WHERE k.phone = $1
	`, phone))
}

func (r *Repository) List(ctx context.Context, q ListQuery) ([]Kontragent, int, error) {
	where := []string{"TRUE"}
	args := make([]any, 0, 8)
	if q.RegionID != nil {
		args = append(args, *q.RegionID)
		where = append(where, fmt.Sprintf("k.region_id = $%d", len(args)))
	}
	if q.DistrictID != nil {
		args = append(args, *q.DistrictID)
		where = append(where, fmt.Sprintf("k.district_id = $%d", len(args)))
	}
	if q.MFYID != nil {
		args = append(args, *q.MFYID)
		where = append(where, fmt.Sprintf("k.mfy_id = $%d", len(args)))
	}
	if q.Query != "" {
		args = append(args, "%"+q.Query+"%")
		where = append(where, fmt.Sprintf(
			"(k.name ILIKE $%d OR k.inn ILIKE $%d OR k.phone ILIKE $%d)",
			len(args), len(args), len(args),
		))
	}
	clause := strings.Join(where, " AND ")

	var total int
	if err := r.pool.QueryRow(ctx, `SELECT COUNT(*) FROM kontragents k WHERE `+clause, args...).Scan(&total); err != nil {
		return nil, 0, err
	}

	args = append(args, q.Limit, (q.Page-1)*q.Limit)
	rows, err := r.pool.Query(ctx, `
		SELECT `+cols+`
		FROM kontragents k
		LEFT JOIN activity_types at ON at.id = k.activity_type_id
		LEFT JOIN regions reg ON reg.id = k.region_id
		LEFT JOIN regions dis ON dis.id = k.district_id
		LEFT JOIN regions mfy ON mfy.id = k.mfy_id
		WHERE `+clause+`
		ORDER BY k.created_at DESC
		LIMIT $`+fmt.Sprintf("%d", len(args)-1)+` OFFSET $`+fmt.Sprintf("%d", len(args))+`
	`, args...)
	if err != nil {
		return nil, 0, err
	}
	defer rows.Close()

	items := make([]Kontragent, 0, q.Limit)
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
		FROM kontragents
	`).Scan(&s.Total, &s.Active)
	return s, err
}

func (r *Repository) SaveCode(ctx context.Context, kontragentID uuid.UUID, phone, purpose, code string, ttl time.Duration) error {
	if _, err := r.pool.Exec(ctx, `
		UPDATE kontragent_verification_codes
		SET used_at = NOW()
		WHERE kontragent_id = $1 AND purpose = $2 AND used_at IS NULL
	`, kontragentID, purpose); err != nil {
		return err
	}
	_, err := r.pool.Exec(ctx, `
		INSERT INTO kontragent_verification_codes (kontragent_id, phone, code_hash, purpose, expires_at)
		VALUES ($1, $2, $3, $4, $5)
	`, kontragentID, phone, hashCode(phone, purpose, code), purpose, time.Now().Add(ttl))
	return err
}

func (r *Repository) LatestCode(ctx context.Context, phone, purpose string) (uuid.UUID, uuid.UUID, string, time.Time, int, *time.Time, *time.Time, error) {
	var (
		id, kontragentID uuid.UUID
		hash             string
		expiresAt        time.Time
		attempts         int
		verifiedAt       *time.Time
		usedAt           *time.Time
	)
	err := r.pool.QueryRow(ctx, `
		SELECT id, kontragent_id, code_hash, expires_at, attempts, verified_at, used_at
		FROM kontragent_verification_codes
		WHERE phone = $1 AND purpose = $2
		ORDER BY created_at DESC
		LIMIT 1
	`, phone, purpose).Scan(&id, &kontragentID, &hash, &expiresAt, &attempts, &verifiedAt, &usedAt)
	if errors.Is(err, pgx.ErrNoRows) {
		return uuid.Nil, uuid.Nil, "", time.Time{}, 0, nil, nil, ErrNotFound
	}
	return id, kontragentID, hash, expiresAt, attempts, verifiedAt, usedAt, err
}

func (r *Repository) BumpAttempts(ctx context.Context, id uuid.UUID) error {
	_, err := r.pool.Exec(ctx, `UPDATE kontragent_verification_codes SET attempts = attempts + 1 WHERE id = $1`, id)
	return err
}

func (r *Repository) MarkVerified(ctx context.Context, id uuid.UUID) error {
	_, err := r.pool.Exec(ctx, `UPDATE kontragent_verification_codes SET verified_at = NOW() WHERE id = $1`, id)
	return err
}

func (r *Repository) MarkUsed(ctx context.Context, id uuid.UUID) error {
	_, err := r.pool.Exec(ctx, `UPDATE kontragent_verification_codes SET used_at = NOW() WHERE id = $1`, id)
	return err
}

func (r *Repository) StoreRefresh(ctx context.Context, kontragentID uuid.UUID, tokenHash string, expiresAt time.Time) error {
	_, err := r.pool.Exec(ctx, `
		INSERT INTO kontragent_refresh_tokens (kontragent_id, token_hash, expires_at)
		VALUES ($1, $2, $3)
	`, kontragentID, tokenHash, expiresAt)
	return err
}

func (r *Repository) FindValidRefresh(ctx context.Context, tokenHash string) (uuid.UUID, error) {
	var id uuid.UUID
	err := r.pool.QueryRow(ctx, `
		SELECT kontragent_id FROM kontragent_refresh_tokens
		WHERE token_hash = $1 AND revoked_at IS NULL AND expires_at > NOW()
	`, tokenHash).Scan(&id)
	if err != nil {
		return uuid.Nil, ErrNotFound
	}
	return id, nil
}

func (r *Repository) RevokeRefresh(ctx context.Context, tokenHash string) error {
	_, err := r.pool.Exec(ctx, `
		UPDATE kontragent_refresh_tokens
		SET revoked_at = NOW()
		WHERE token_hash = $1 AND revoked_at IS NULL
	`, tokenHash)
	return err
}

func hashCode(phone, purpose, code string) string {
	sum := sha256.Sum256([]byte(phone + "|" + purpose + "|" + code))
	return hex.EncodeToString(sum[:])
}
