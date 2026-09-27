package manager

import (
	"context"
	"errors"
	"regexp"
	"strings"
	"unicode"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgconn"
	"golang.org/x/crypto/bcrypt"

	"qoshida/backend/internal/modules/region"
	"qoshida/backend/internal/shared/apperror"
	"qoshida/backend/internal/shared/audit"
)

var (
	phoneRE    = regexp.MustCompile(`^\+998[0-9]{9}$`)
	usernameRE = regexp.MustCompile(`^[a-zA-Z0-9._]{3,32}$`)
)

type Service struct {
	repo    *Repository
	regions *region.Repository
}

func NewService(repo *Repository, regions *region.Repository) *Service {
	return &Service{repo: repo, regions: regions}
}

func (s *Service) List(ctx context.Context, q ListQuery) (ListResult, error) {
	if q.Page < 1 {
		q.Page = 1
	}
	if q.Limit < 1 || q.Limit > 100 {
		q.Limit = 20
	}
	items, total, err := s.repo.List(ctx, q)
	if err != nil {
		return ListResult{}, apperror.Internal()
	}
	ids := make([]uuid.UUID, 0, len(items))
	for _, item := range items {
		ids = append(ids, item.ID)
	}
	trails := audit.BindMany(ctx, "manager", ids)
	out := make([]Public, 0, len(items))
	for _, item := range items {
		pub := item.Public()
		if t, ok := trails[item.ID]; ok {
			pub.Audit = &t
		}
		out = append(out, pub)
	}
	return ListResult{Items: out, Total: total, Page: q.Page, Limit: q.Limit}, nil
}

func (s *Service) Get(ctx context.Context, id uuid.UUID) (Public, error) {
	item, err := s.repo.FindByID(ctx, id)
	if err != nil {
		if errors.Is(err, ErrNotFound) {
			return Public{}, apperror.NotFound("Menejer topilmadi")
		}
		return Public{}, apperror.Internal()
	}
	pub := item.Public()
	pub.Audit = audit.Bind(ctx, "manager", item.ID)
	return pub, nil
}

func (s *Service) Create(ctx context.Context, req CreateRequest, forced *Manager) (Public, error) {
	item, err := s.build(ctx, uuid.Nil, req, forced, "")
	if err != nil {
		return Public{}, err
	}
	created, err := s.repo.Create(ctx, item)
	if err != nil {
		return Public{}, mapWriteError(err)
	}
	pub := created.Public()
	pub.Audit = audit.AfterCreate(ctx, "manager", created.ID)
	return pub, nil
}

func (s *Service) Update(ctx context.Context, id uuid.UUID, req CreateRequest, forced *Manager) (Public, error) {
	current, err := s.repo.FindByID(ctx, id)
	if err != nil {
		if errors.Is(err, ErrNotFound) {
			return Public{}, apperror.NotFound("Menejer topilmadi")
		}
		return Public{}, apperror.Internal()
	}
	if forced != nil && (current.Type != TypeDistrict || current.RegionID != forced.RegionID) {
		return Public{}, apperror.Forbidden("Faqat o'z viloyatingizdagi tuman menejerini tahrirlash mumkin")
	}
	item, err := s.build(ctx, id, req, forced, current.PasswordHash)
	if err != nil {
		return Public{}, err
	}
	updated, err := s.repo.Update(ctx, item)
	if err != nil {
		return Public{}, mapWriteError(err)
	}
	pub := updated.Public()
	pub.Audit = audit.AfterUpdate(ctx, "manager", updated.ID)
	return pub, nil
}

func (s *Service) Delete(ctx context.Context, id uuid.UUID, forced *Manager) error {
	current, err := s.repo.FindByID(ctx, id)
	if err != nil {
		if errors.Is(err, ErrNotFound) {
			return apperror.NotFound("Menejer topilmadi")
		}
		return apperror.Internal()
	}
	if forced != nil && (current.Type != TypeDistrict || current.RegionID != forced.RegionID) {
		return apperror.Forbidden("Faqat o'z viloyatingizdagi tuman menejerini o'chirish mumkin")
	}
	if err = s.repo.Delete(ctx, id); err != nil {
		if errors.Is(err, ErrNotFound) {
			return apperror.NotFound("Menejer topilmadi")
		}
		return apperror.Internal()
	}
	return nil
}

func (s *Service) Stats(ctx context.Context) (Stats, error) {
	stats, err := s.repo.Stats(ctx)
	if err != nil {
		return Stats{}, apperror.Internal()
	}
	return stats, nil
}

func (s *Service) build(ctx context.Context, id uuid.UUID, req CreateRequest, forced *Manager, existingHash string) (Manager, error) {
	req.Type = strings.TrimSpace(req.Type)
	req.FirstName = strings.TrimSpace(req.FirstName)
	req.LastName = strings.TrimSpace(req.LastName)
	req.Phone = strings.TrimSpace(req.Phone)
	req.Username = strings.TrimSpace(req.Username)
	req.Status = strings.TrimSpace(req.Status)
	if req.Status == "" {
		req.Status = StatusActive
	}
	if forced != nil {
		req.Type = TypeDistrict
		req.RegionID = forced.RegionID.String()
	}
	if req.Type != TypeRegion && req.Type != TypeDistrict {
		return Manager{}, apperror.BadRequest("Tur: region yoki district")
	}
	if err := validateIdentity(req.FirstName, req.LastName, req.Phone, req.Username); err != nil {
		return Manager{}, err
	}
	if req.Status != StatusActive && req.Status != "inactive" {
		return Manager{}, apperror.BadRequest("status: active yoki inactive")
	}

	regionID, err := uuid.Parse(strings.TrimSpace(req.RegionID))
	if err != nil {
		return Manager{}, apperror.BadRequest("Viloyat tanlanishi shart")
	}
	reg, err := s.regions.FindByID(ctx, regionID)
	if err != nil {
		if errors.Is(err, region.ErrNotFound) {
			return Manager{}, apperror.BadRequest("Viloyat topilmadi")
		}
		return Manager{}, apperror.Internal()
	}
	if reg.Type != region.TypeRegion {
		return Manager{}, apperror.BadRequest("Viloyat noto'g'ri tanlangan")
	}

	var districtID *uuid.UUID
	if req.Type == TypeDistrict {
		did, err := uuid.Parse(strings.TrimSpace(req.DistrictID))
		if err != nil {
			return Manager{}, apperror.BadRequest("Tuman tanlanishi shart")
		}
		dis, err := s.regions.FindByID(ctx, did)
		if err != nil {
			if errors.Is(err, region.ErrNotFound) {
				return Manager{}, apperror.BadRequest("Tuman topilmadi")
			}
			return Manager{}, apperror.Internal()
		}
		if dis.Type != region.TypeDistrict || dis.ParentID == nil || *dis.ParentID != regionID {
			return Manager{}, apperror.BadRequest("Tuman tanlangan viloyatga tegishli emas")
		}
		districtID = &did
	}

	hash := existingHash
	if strings.TrimSpace(req.Password) != "" {
		if err = validatePassword(req.Password); err != nil {
			return Manager{}, err
		}
		raw, err := bcrypt.GenerateFromPassword([]byte(req.Password), 12)
		if err != nil {
			return Manager{}, apperror.Internal()
		}
		hash = string(raw)
	}

	return Manager{
		ID:           id,
		Type:         req.Type,
		RegionID:     regionID,
		DistrictID:   districtID,
		FirstName:    req.FirstName,
		LastName:     req.LastName,
		Phone:        req.Phone,
		Username:     req.Username,
		PasswordHash: hash,
		Status:       req.Status,
	}, nil
}

func validateIdentity(firstName, lastName, phone, username string) error {
	if len(firstName) < 2 || len(firstName) > 80 {
		return apperror.BadRequest("Ism 2-80 belgi oralig'ida bo'lishi kerak")
	}
	if len(lastName) < 2 || len(lastName) > 80 {
		return apperror.BadRequest("Familiya 2-80 belgi oralig'ida bo'lishi kerak")
	}
	if !phoneRE.MatchString(phone) {
		return apperror.BadRequest("Telefon formati: +998XXXXXXXXX")
	}
	if !usernameRE.MatchString(username) {
		return apperror.BadRequest("Username 3-32 belgi: harf, raqam, nuqta yoki pastki chiziq")
	}
	return nil
}

func validatePassword(password string) error {
	if len(password) < 10 || len(password) > 72 {
		return apperror.BadRequest("Parol 10-72 belgi oralig'ida bo'lishi kerak")
	}
	var hasLetter, hasDigit bool
	for _, r := range password {
		if unicode.IsLetter(r) {
			hasLetter = true
		}
		if unicode.IsDigit(r) {
			hasDigit = true
		}
	}
	if !hasLetter || !hasDigit {
		return apperror.BadRequest("Parolda kamida bitta harf va bitta raqam bo'lishi kerak")
	}
	return nil
}

func mapWriteError(err error) error {
	var pgErr *pgconn.PgError
	if errors.As(err, &pgErr) && pgErr.Code == "23505" {
		return apperror.Conflict("Username yoki telefon allaqachon mavjud")
	}
	return apperror.Internal()
}

func NormalizePhone(phone string) (string, error) {
	p := strings.TrimSpace(phone)
	p = strings.ReplaceAll(p, " ", "")
	p = strings.ReplaceAll(p, "-", "")
	if strings.HasPrefix(p, "998") && len(p) == 12 {
		p = "+" + p
	}
	if !strings.HasPrefix(p, "+") && len(p) == 9 {
		p = "+998" + p
	}
	if !phoneRE.MatchString(p) {
		return "", apperror.BadRequest("Telefon formati: +998XXXXXXXXX")
	}
	return p, nil
}
