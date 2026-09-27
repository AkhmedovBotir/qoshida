package providerservice

import (
	"context"
	"errors"
	"strings"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

var ErrNotFound = errors.New("provider service not found")

type Repository struct {
	pool *pgxpool.Pool
}

func NewRepository(pool *pgxpool.Pool) *Repository {
	return &Repository{pool: pool}
}

const cols = `
	s.id, s.provider_id, s.name, s.price, COALESCE(s.images, '{}'), s.approval_status, s.rejection_note, s.submitted_by,
	COALESCE(s.reviewed_by_name, ''), COALESCE(s.reviewed_by_role, ''), s.reviewed_at,
	s.created_at, s.updated_at,
	COALESCE(p.name, ''), p.region_id, p.district_id, p.mfy_id
`

func scan(row pgx.Row) (ServiceItem, error) {
	var item ServiceItem
	err := row.Scan(
		&item.ID, &item.ProviderID, &item.Name, &item.Price, &item.Images, &item.ApprovalStatus, &item.RejectionNote, &item.SubmittedBy,
		&item.ReviewedByName, &item.ReviewedByRole, &item.ReviewedAt,
		&item.CreatedAt, &item.UpdatedAt,
		&item.ProviderName, &item.RegionID, &item.DistrictID, &item.MFYID,
	)
	if errors.Is(err, pgx.ErrNoRows) {
		return ServiceItem{}, ErrNotFound
	}
	if item.Images == nil {
		item.Images = []string{}
	}
	return item, err
}

func (r *Repository) CreateMany(ctx context.Context, items []ServiceItem) ([]ServiceItem, error) {
	tx, err := r.pool.Begin(ctx)
	if err != nil {
		return nil, err
	}
	defer tx.Rollback(ctx)

	ids := make([]uuid.UUID, 0, len(items))
	for i := range items {
		if items[i].ID == uuid.Nil {
			items[i].ID = uuid.New()
		}
		if items[i].Images == nil {
			items[i].Images = []string{}
		}
		if err = tx.QueryRow(ctx, `
			INSERT INTO provider_services (
				id, provider_id, name, price, images, approval_status, rejection_note, submitted_by,
				reviewed_by_name, reviewed_by_role, reviewed_at
			) VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11)
			RETURNING id
		`, items[i].ID, items[i].ProviderID, items[i].Name, items[i].Price, items[i].Images,
			items[i].ApprovalStatus, items[i].RejectionNote, items[i].SubmittedBy,
			items[i].ReviewedByName, items[i].ReviewedByRole, items[i].ReviewedAt,
		).Scan(&items[i].ID); err != nil {
			return nil, err
		}
		ids = append(ids, items[i].ID)
	}
	if err = tx.Commit(ctx); err != nil {
		return nil, err
	}

	out := make([]ServiceItem, 0, len(ids))
	for _, id := range ids {
		item, err := r.FindByID(ctx, id)
		if err != nil {
			return nil, err
		}
		out = append(out, item)
	}
	return out, nil
}

func (r *Repository) Update(ctx context.Context, item ServiceItem) (ServiceItem, error) {
	tag, err := r.pool.Exec(ctx, `
		UPDATE provider_services
		SET name = $2, price = $3, images = $4, approval_status = $5, rejection_note = $6, submitted_by = $7,
			reviewed_by_name = $8, reviewed_by_role = $9, reviewed_at = $10, updated_at = NOW()
		WHERE id = $1
	`, item.ID, item.Name, item.Price, item.Images, item.ApprovalStatus, item.RejectionNote, item.SubmittedBy,
		item.ReviewedByName, item.ReviewedByRole, item.ReviewedAt)
	if err != nil {
		return ServiceItem{}, err
	}
	if tag.RowsAffected() == 0 {
		return ServiceItem{}, ErrNotFound
	}
	return r.FindByID(ctx, item.ID)
}

func (r *Repository) Delete(ctx context.Context, id uuid.UUID) error {
	tag, err := r.pool.Exec(ctx, `DELETE FROM provider_services WHERE id = $1`, id)
	if err != nil {
		return err
	}
	if tag.RowsAffected() == 0 {
		return ErrNotFound
	}
	return nil
}

func (r *Repository) FindByID(ctx context.Context, id uuid.UUID) (ServiceItem, error) {
	return scan(r.pool.QueryRow(ctx, `
		SELECT `+cols+`
		FROM provider_services s
		JOIN service_providers p ON p.id = s.provider_id
		WHERE s.id = $1
	`, id))
}

func (r *Repository) List(ctx context.Context, q ListQuery) ([]ServiceItem, int, error) {
	where := []string{"1=1"}
	args := []any{}
	add := func(cond string, val any) {
		args = append(args, val)
		where = append(where, strings.Replace(cond, "?", "$"+itoa(len(args)), 1))
	}
	if q.ProviderID != nil {
		add("s.provider_id = ?", *q.ProviderID)
	}
	if q.ApprovalStatus != "" {
		add("s.approval_status = ?", q.ApprovalStatus)
	}
	if q.RegionID != nil {
		add("p.region_id = ?", *q.RegionID)
	}
	if q.DistrictID != nil {
		add("p.district_id = ?", *q.DistrictID)
	}
	if q.MFYID != nil {
		add("p.mfy_id = ?", *q.MFYID)
	}
	if q.Query != "" {
		args = append(args, "%"+q.Query+"%")
		n := itoa(len(args))
		where = append(where, "(s.name ILIKE $"+n+" OR p.name ILIKE $"+n+")")
	}

	clause := strings.Join(where, " AND ")
	var total int
	if err := r.pool.QueryRow(ctx, `
		SELECT COUNT(*)
		FROM provider_services s
		JOIN service_providers p ON p.id = s.provider_id
		WHERE `+clause, args...).Scan(&total); err != nil {
		return nil, 0, err
	}

	offset := (q.Page - 1) * q.Limit
	args = append(args, q.Limit, offset)
	limitPos := itoa(len(args) - 1)
	offsetPos := itoa(len(args))
	rows, err := r.pool.Query(ctx, `
		SELECT `+cols+`
		FROM provider_services s
		JOIN service_providers p ON p.id = s.provider_id
		WHERE `+clause+`
		ORDER BY s.created_at DESC
		LIMIT $`+limitPos+` OFFSET $`+offsetPos, args...)
	if err != nil {
		return nil, 0, err
	}
	defer rows.Close()

	items := make([]ServiceItem, 0)
	for rows.Next() {
		item, err := scan(rows)
		if err != nil {
			return nil, 0, err
		}
		items = append(items, item)
	}
	return items, total, rows.Err()
}

func (r *Repository) FindProvider(ctx context.Context, id uuid.UUID) (ProviderRef, error) {
	var item ProviderRef
	err := r.pool.QueryRow(ctx, `
		SELECT p.id, p.name, p.region_id, p.district_id, p.mfy_id, p.status, COALESCE(ident.status, 'none')
		FROM service_providers p
		LEFT JOIN service_provider_identifications ident ON ident.provider_id = p.id
		WHERE p.id = $1
	`, id).Scan(&item.ID, &item.Name, &item.RegionID, &item.DistrictID, &item.MFYID, &item.Status, &item.IdentificationStatus)
	if errors.Is(err, pgx.ErrNoRows) {
		return ProviderRef{}, ErrNotFound
	}
	return item, err
}

func itoa(n int) string {
	if n == 0 {
		return "0"
	}
	var b [12]byte
	i := len(b)
	for n > 0 {
		i--
		b[i] = byte('0' + n%10)
		n /= 10
	}
	return string(b[i:])
}
