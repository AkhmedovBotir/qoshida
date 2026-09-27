package serviceprovider

import (
	"context"
	"errors"
	"fmt"
	"strings"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
)

const identCols = `
	i.id, i.provider_id, i.full_name, i.pinfl, i.passport, i.birth_date, i.address, i.experience_years, i.about,
	i.passport_image, i.selfie_image, i.documents, i.status, i.rejection_note, i.reviewed_by_name, i.reviewed_by_role,
	i.reviewed_at, i.submitted_at, i.created_at, i.updated_at,
	p.name, p.phone, COALESCE(at.name, ''), COALESCE(reg.name, ''), COALESCE(dis.name, ''), COALESCE(mfy.name, ''),
	p.region_id, p.district_id, p.mfy_id
`

const identFrom = `
	service_provider_identifications i
	JOIN service_providers p ON p.id = i.provider_id
	LEFT JOIN activity_types at ON at.id = p.activity_type_id
	LEFT JOIN regions reg ON reg.id = p.region_id
	LEFT JOIN regions dis ON dis.id = p.district_id
	LEFT JOIN regions mfy ON mfy.id = p.mfy_id
`

func scanIdent(row pgx.Row) (Identification, error) {
	var (
		item Identification
		docs []byte
	)
	err := row.Scan(
		&item.ID, &item.ProviderID, &item.FullName, &item.PINFL, &item.Passport, &item.BirthDate, &item.Address,
		&item.ExperienceYears, &item.About, &item.PassportImage, &item.SelfieImage, &docs, &item.Status,
		&item.RejectionNote, &item.ReviewedByName, &item.ReviewedByRole, &item.ReviewedAt, &item.SubmittedAt,
		&item.CreatedAt, &item.UpdatedAt, &item.ProviderName, &item.Phone, &item.ActivityName, &item.RegionName,
		&item.DistrictName, &item.MFYName, &item.RegionID, &item.DistrictID, &item.MFYID,
	)
	if errors.Is(err, pgx.ErrNoRows) {
		return Identification{}, ErrNotFound
	}
	if err != nil {
		return Identification{}, err
	}
	item.Documents = unmarshalDocs(docs)
	return item, nil
}

func (r *Repository) FindIdentByID(ctx context.Context, id uuid.UUID) (Identification, error) {
	return scanIdent(r.pool.QueryRow(ctx, `SELECT `+identCols+` FROM `+identFrom+` WHERE i.id = $1`, id))
}

func (r *Repository) FindIdentByProvider(ctx context.Context, providerID uuid.UUID) (Identification, error) {
	return scanIdent(r.pool.QueryRow(ctx, `SELECT `+identCols+` FROM `+identFrom+` WHERE i.provider_id = $1`, providerID))
}

func (r *Repository) CreateIdent(ctx context.Context, item Identification) (Identification, error) {
	_, err := r.pool.Exec(ctx, `
		INSERT INTO service_provider_identifications (
			id, provider_id, full_name, pinfl, passport, birth_date, address, experience_years, about,
			passport_image, selfie_image, documents, status, rejection_note, submitted_at
		) VALUES ($1,$2,$3,$4,$5,$6,$7,$8,$9,$10,$11,$12,$13,'', $14)
	`, item.ID, item.ProviderID, item.FullName, item.PINFL, item.Passport, item.BirthDate, item.Address,
		item.ExperienceYears, item.About, item.PassportImage, item.SelfieImage, marshalDocs(item.Documents),
		item.Status, item.SubmittedAt)
	if err != nil {
		return Identification{}, err
	}
	return r.FindIdentByID(ctx, item.ID)
}

func (r *Repository) UpdateIdent(ctx context.Context, item Identification) (Identification, error) {
	tag, err := r.pool.Exec(ctx, `
		UPDATE service_provider_identifications
		SET full_name = $2, pinfl = $3, passport = $4, birth_date = $5, address = $6, experience_years = $7, about = $8,
		    passport_image = $9, selfie_image = $10, documents = $11, status = $12, rejection_note = '',
		    reviewed_by_name = '', reviewed_by_role = '', reviewed_at = NULL, submitted_at = $13, updated_at = NOW()
		WHERE id = $1
	`, item.ID, item.FullName, item.PINFL, item.Passport, item.BirthDate, item.Address, item.ExperienceYears, item.About,
		item.PassportImage, item.SelfieImage, marshalDocs(item.Documents), item.Status, item.SubmittedAt)
	if err != nil {
		return Identification{}, err
	}
	if tag.RowsAffected() == 0 {
		return Identification{}, ErrNotFound
	}
	return r.FindIdentByID(ctx, item.ID)
}

func (r *Repository) UpdateIdentReview(ctx context.Context, item Identification) (Identification, error) {
	tag, err := r.pool.Exec(ctx, `
		UPDATE service_provider_identifications
		SET status = $2, rejection_note = $3, reviewed_by_name = $4, reviewed_by_role = $5, reviewed_at = $6, updated_at = NOW()
		WHERE id = $1
	`, item.ID, item.Status, item.RejectionNote, item.ReviewedByName, item.ReviewedByRole, item.ReviewedAt)
	if err != nil {
		return Identification{}, err
	}
	if tag.RowsAffected() == 0 {
		return Identification{}, ErrNotFound
	}
	return r.FindIdentByID(ctx, item.ID)
}

func (r *Repository) ListIdent(ctx context.Context, q IdentListQuery) ([]Identification, int, error) {
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
	if q.Status == IdentPending || q.Status == IdentApproved || q.Status == IdentRejected {
		args = append(args, q.Status)
		where = append(where, fmt.Sprintf("i.status = $%d", len(args)))
	}
	if q.Query != "" {
		args = append(args, "%"+q.Query+"%")
		where = append(where, fmt.Sprintf("(p.name ILIKE $%d OR p.phone ILIKE $%d OR i.full_name ILIKE $%d OR i.pinfl ILIKE $%d)", len(args), len(args), len(args), len(args)))
	}
	clause := strings.Join(where, " AND ")

	var total int
	if err := r.pool.QueryRow(ctx, `SELECT COUNT(*) FROM `+identFrom+` WHERE `+clause, args...).Scan(&total); err != nil {
		return nil, 0, err
	}

	args = append(args, q.Limit, (q.Page-1)*q.Limit)
	rows, err := r.pool.Query(ctx, `
		SELECT `+identCols+`
		FROM `+identFrom+`
		WHERE `+clause+`
		ORDER BY CASE i.status WHEN 'pending' THEN 0 WHEN 'rejected' THEN 1 ELSE 2 END, i.submitted_at DESC
		LIMIT $`+fmt.Sprintf("%d", len(args)-1)+` OFFSET $`+fmt.Sprintf("%d", len(args))+`
	`, args...)
	if err != nil {
		return nil, 0, err
	}
	defer rows.Close()

	items := make([]Identification, 0, q.Limit)
	for rows.Next() {
		item, err := scanIdent(rows)
		if err != nil {
			return nil, 0, err
		}
		items = append(items, item)
	}
	return items, total, rows.Err()
}

func identPrefix(providerID uuid.UUID) string {
	return "/uploads/service-provider-ident/" + providerID.String() + "/"
}

func identPathOK(providerID uuid.UUID, path string) bool {
	path = strings.TrimSpace(path)
	return path != "" && strings.HasPrefix(path, identPrefix(providerID))
}
