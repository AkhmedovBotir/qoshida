package seller

import (
	"context"
	"errors"
	"regexp"
	"strings"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgconn"
	"golang.org/x/crypto/bcrypt"

	"qoshida/backend/internal/shared/apperror"
	"qoshida/backend/internal/shared/audit"
)

var phoneRE = regexp.MustCompile(`^\+998[0-9]{9}$`)

type Service struct {
	repo *Repository
}

func NewService(repo *Repository) *Service {
	return &Service{repo: repo}
}

func (s *Service) List(ctx context.Context, q ListQuery, scope *Scope) (ListResult, error) {
	if q.Page < 1 {
		q.Page = 1
	}
	if q.Limit < 1 || q.Limit > 100 {
		q.Limit = 20
	}
	applyListScope(&q, scope)
	items, total, err := s.repo.List(ctx, q)
	if err != nil {
		return ListResult{}, apperror.Internal()
	}
	ids := make([]uuid.UUID, 0, len(items))
	for _, item := range items {
		ids = append(ids, item.ID)
	}
	trails := audit.BindMany(ctx, "seller", ids)
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

func (s *Service) Get(ctx context.Context, id uuid.UUID, scope *Scope) (Public, error) {
	item, err := s.repo.FindByID(ctx, id)
	if err != nil {
		if errors.Is(err, ErrNotFound) {
			return Public{}, apperror.NotFound("Sotuvchi topilmadi")
		}
		return Public{}, apperror.Internal()
	}
	if err = assertInScope(item, scope); err != nil {
		return Public{}, err
	}
	pub := item.Public()
	pub.Audit = audit.Bind(ctx, "seller", item.ID)
	return pub, nil
}

func (s *Service) Create(ctx context.Context, req CreateRequest, scope *Scope) (Public, error) {
	item, err := s.build(ctx, uuid.Nil, req, scope, "")
	if err != nil {
		return Public{}, err
	}
	created, err := s.repo.Create(ctx, item)
	if err != nil {
		return Public{}, mapWriteError(err)
	}
	pub := created.Public()
	pub.Audit = audit.AfterCreate(ctx, "seller", created.ID)
	return pub, nil
}

func (s *Service) Update(ctx context.Context, id uuid.UUID, req CreateRequest, scope *Scope) (Public, error) {
	current, err := s.repo.FindByID(ctx, id)
	if err != nil {
		if errors.Is(err, ErrNotFound) {
			return Public{}, apperror.NotFound("Sotuvchi topilmadi")
		}
		return Public{}, apperror.Internal()
	}
	if err = assertInScope(current, scope); err != nil {
		return Public{}, err
	}
	item, err := s.build(ctx, id, req, scope, current.PasswordHash)
	if err != nil {
		return Public{}, err
	}
	updated, err := s.repo.Update(ctx, item)
	if err != nil {
		return Public{}, mapWriteError(err)
	}
	pub := updated.Public()
	pub.Audit = audit.AfterUpdate(ctx, "seller", updated.ID)
	return pub, nil
}

func (s *Service) Delete(ctx context.Context, id uuid.UUID, scope *Scope) error {
	current, err := s.repo.FindByID(ctx, id)
	if err != nil {
		if errors.Is(err, ErrNotFound) {
			return apperror.NotFound("Sotuvchi topilmadi")
		}
		return apperror.Internal()
	}
	if err = assertInScope(current, scope); err != nil {
		return err
	}
	if err = s.repo.Delete(ctx, id); err != nil {
		if errors.Is(err, ErrNotFound) {
			return apperror.NotFound("Sotuvchi topilmadi")
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

func (s *Service) build(ctx context.Context, id uuid.UUID, req CreateRequest, scope *Scope, existingHash string) (Seller, error) {
	req.FirstName = strings.TrimSpace(req.FirstName)
	req.LastName = strings.TrimSpace(req.LastName)
	req.Phone = strings.TrimSpace(req.Phone)
	req.Status = strings.TrimSpace(req.Status)
	if req.Status == "" {
		req.Status = StatusActive
	}
	if scope != nil && scope.ShopID != nil {
		req.ShopID = scope.ShopID.String()
	}
	if err := validateIdentity(req.FirstName, req.LastName, req.Phone); err != nil {
		return Seller{}, err
	}
	if req.Status != StatusActive && req.Status != "inactive" {
		return Seller{}, apperror.BadRequest("status: active yoki inactive")
	}

	shopID, err := uuid.Parse(strings.TrimSpace(req.ShopID))
	if err != nil {
		return Seller{}, apperror.BadRequest("Do‘kon tanlanishi shart")
	}
	shop, err := s.repo.FindShop(ctx, shopID)
	if err != nil {
		if errors.Is(err, ErrNotFound) {
			return Seller{}, apperror.BadRequest("Do‘kon topilmadi")
		}
		return Seller{}, apperror.Internal()
	}
	if err = assertShopInScope(shop, scope); err != nil {
		return Seller{}, err
	}

	hash := existingHash
	if strings.TrimSpace(req.Password) != "" {
		if err = simplePassword(req.Password); err != nil {
			return Seller{}, err
		}
		raw, err := bcrypt.GenerateFromPassword([]byte(req.Password), 12)
		if err != nil {
			return Seller{}, apperror.Internal()
		}
		hash = string(raw)
	}

	return Seller{
		ID:           id,
		ShopID:       shop.ID,
		FirstName:    req.FirstName,
		LastName:     req.LastName,
		Phone:        req.Phone,
		PasswordHash: hash,
		Status:       req.Status,
	}, nil
}

func applyListScope(q *ListQuery, scope *Scope) {
	if scope == nil {
		return
	}
	if scope.ShopID != nil {
		q.ShopID = scope.ShopID
	}
	if scope.RegionID != nil {
		q.RegionID = scope.RegionID
	}
	if scope.DistrictID != nil {
		q.DistrictID = scope.DistrictID
	}
	if scope.MFYID != nil {
		q.MFYID = scope.MFYID
	}
}

func assertInScope(item Seller, scope *Scope) error {
	if scope == nil {
		return nil
	}
	if scope.ShopID != nil && item.ShopID != *scope.ShopID {
		return apperror.Forbidden("Faqat o‘z do‘koningizdagi sotuvchini boshqarish mumkin")
	}
	if scope.RegionID != nil && item.RegionID != *scope.RegionID {
		return apperror.Forbidden("Faqat o‘z viloyatingizdagi sotuvchini boshqarish mumkin")
	}
	if scope.DistrictID != nil && item.DistrictID != *scope.DistrictID {
		return apperror.Forbidden("Faqat o‘z tumaningizdagi sotuvchini boshqarish mumkin")
	}
	if scope.MFYID != nil && item.MFYID != *scope.MFYID {
		return apperror.Forbidden("Faqat o‘z MFYingizdagi sotuvchini boshqarish mumkin")
	}
	return nil
}

func assertShopInScope(shop ShopRef, scope *Scope) error {
	if scope == nil {
		return nil
	}
	if scope.ShopID != nil && shop.ID != *scope.ShopID {
		return apperror.Forbidden("Faqat o‘z do‘koningizga sotuvchi qo‘shish mumkin")
	}
	if scope.RegionID != nil && shop.RegionID != *scope.RegionID {
		return apperror.Forbidden("Do‘kon sizning viloyatingizga tegishli emas")
	}
	if scope.DistrictID != nil && shop.DistrictID != *scope.DistrictID {
		return apperror.Forbidden("Do‘kon sizning tumaningizga tegishli emas")
	}
	if scope.MFYID != nil && shop.MFYID != *scope.MFYID {
		return apperror.Forbidden("Do‘kon sizning MFYingizga tegishli emas")
	}
	return nil
}

func validateIdentity(firstName, lastName, phone string) error {
	if len(firstName) < 2 || len(firstName) > 80 {
		return apperror.BadRequest("Ism 2-80 belgi oralig'ida bo'lishi kerak")
	}
	if len(lastName) < 2 || len(lastName) > 80 {
		return apperror.BadRequest("Familiya 2-80 belgi oralig'ida bo'lishi kerak")
	}
	if !phoneRE.MatchString(phone) {
		return apperror.BadRequest("Telefon formati: +998XXXXXXXXX")
	}
	return nil
}

func simplePassword(password string) error {
	password = strings.TrimSpace(password)
	if password == "" {
		return apperror.BadRequest("Parol kiritilishi shart")
	}
	if len(password) > 72 {
		return apperror.BadRequest("Parol juda uzun")
	}
	return nil
}

func mapWriteError(err error) error {
	var pgErr *pgconn.PgError
	if errors.As(err, &pgErr) && pgErr.Code == "23505" {
		return apperror.Conflict("Telefon allaqachon mavjud")
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
