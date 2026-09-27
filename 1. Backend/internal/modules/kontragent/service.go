package kontragent

import (
	"context"
	"errors"
	"regexp"
	"strings"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgconn"
	"golang.org/x/crypto/bcrypt"

	"qoshida/backend/internal/modules/activitytype"
	"qoshida/backend/internal/modules/region"
	"qoshida/backend/internal/shared/apperror"
	"qoshida/backend/internal/shared/audit"
)

var (
	phoneRE = regexp.MustCompile(`^\+998[0-9]{9}$`)
	innRE   = regexp.MustCompile(`^[0-9]{9}$`)
)

type Service struct {
	repo       *Repository
	regions    *region.Repository
	activities *activitytype.Repository
}

func NewService(repo *Repository, regions *region.Repository, activities *activitytype.Repository) *Service {
	return &Service{repo: repo, regions: regions, activities: activities}
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
	trails := audit.BindMany(ctx, "kontragent", ids)
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
			return Public{}, apperror.NotFound("Kontragent topilmadi")
		}
		return Public{}, apperror.Internal()
	}
	if err = assertInScope(item, scope); err != nil {
		return Public{}, err
	}
	pub := item.Public()
	pub.Audit = audit.Bind(ctx, "kontragent", item.ID)
	return pub, nil
}

func (s *Service) Create(ctx context.Context, req CreateRequest, scope *Scope) (Public, error) {
	item, err := s.build(ctx, uuid.New(), req, scope, "", "")
	if err != nil {
		return Public{}, err
	}
	created, err := s.repo.Create(ctx, item)
	if err != nil {
		removeLogoFile(item.Logo)
		return Public{}, mapWriteError(err)
	}
	pub := created.Public()
	pub.Audit = audit.AfterCreate(ctx, "kontragent", created.ID)
	return pub, nil
}

func (s *Service) Update(ctx context.Context, id uuid.UUID, req CreateRequest, scope *Scope) (Public, error) {
	current, err := s.repo.FindByID(ctx, id)
	if err != nil {
		if errors.Is(err, ErrNotFound) {
			return Public{}, apperror.NotFound("Kontragent topilmadi")
		}
		return Public{}, apperror.Internal()
	}
	if err = assertInScope(current, scope); err != nil {
		return Public{}, err
	}
	item, err := s.build(ctx, id, req, scope, current.PasswordHash, current.Logo)
	if err != nil {
		return Public{}, err
	}
	updated, err := s.repo.Update(ctx, item)
	if err != nil {
		return Public{}, mapWriteError(err)
	}
	if current.Logo != "" && current.Logo != item.Logo {
		removeLogoFile(current.Logo)
	}
	pub := updated.Public()
	pub.Audit = audit.AfterUpdate(ctx, "kontragent", updated.ID)
	return pub, nil
}

func (s *Service) Delete(ctx context.Context, id uuid.UUID, scope *Scope) error {
	current, err := s.repo.FindByID(ctx, id)
	if err != nil {
		if errors.Is(err, ErrNotFound) {
			return apperror.NotFound("Kontragent topilmadi")
		}
		return apperror.Internal()
	}
	if err = assertInScope(current, scope); err != nil {
		return err
	}
	if err = s.repo.Delete(ctx, id); err != nil {
		if errors.Is(err, ErrNotFound) {
			return apperror.NotFound("Kontragent topilmadi")
		}
		return apperror.Internal()
	}
	removeLogoFile(current.Logo)
	return nil
}

func (s *Service) Stats(ctx context.Context) (Stats, error) {
	stats, err := s.repo.Stats(ctx)
	if err != nil {
		return Stats{}, apperror.Internal()
	}
	return stats, nil
}

func (s *Service) build(ctx context.Context, id uuid.UUID, req CreateRequest, scope *Scope, existingHash, existingLogo string) (Kontragent, error) {
	req.Name = strings.TrimSpace(req.Name)
	req.INN = strings.TrimSpace(req.INN)
	req.Phone = strings.TrimSpace(req.Phone)
	req.Status = strings.TrimSpace(req.Status)
	if req.Status == "" {
		req.Status = StatusActive
	}
	if scope != nil {
		req.RegionID = scope.RegionID.String()
		if scope.DistrictID != nil {
			req.DistrictID = scope.DistrictID.String()
		}
		if scope.MFYID != nil {
			req.MFYID = scope.MFYID.String()
		}
	}
	if len(req.Name) < 2 || len(req.Name) > 160 {
		return Kontragent{}, apperror.BadRequest("Nomi 2-160 belgi oralig'ida bo'lishi kerak")
	}
	if !innRE.MatchString(req.INN) {
		return Kontragent{}, apperror.BadRequest("INN 9 xonali raqam bo'lishi kerak")
	}
	if !phoneRE.MatchString(req.Phone) {
		return Kontragent{}, apperror.BadRequest("Telefon formati: +998XXXXXXXXX")
	}
	if req.Status != StatusActive && req.Status != "inactive" {
		return Kontragent{}, apperror.BadRequest("status: active yoki inactive")
	}

	activityID, err := uuid.Parse(strings.TrimSpace(req.ActivityTypeID))
	if err != nil {
		return Kontragent{}, apperror.BadRequest("Faoliyat turi tanlanishi shart")
	}
	activity, err := s.activities.FindByID(ctx, activityID)
	if err != nil {
		if errors.Is(err, activitytype.ErrNotFound) {
			return Kontragent{}, apperror.BadRequest("Faoliyat turi topilmadi")
		}
		return Kontragent{}, apperror.Internal()
	}
	if activity.Status != activitytype.StatusActive {
		return Kontragent{}, apperror.BadRequest("Faoliyat turi faol emas")
	}

	regionID, err := uuid.Parse(strings.TrimSpace(req.RegionID))
	if err != nil {
		return Kontragent{}, apperror.BadRequest("Viloyat tanlanishi shart")
	}
	reg, err := s.regions.FindByID(ctx, regionID)
	if err != nil {
		if errors.Is(err, region.ErrNotFound) {
			return Kontragent{}, apperror.BadRequest("Viloyat topilmadi")
		}
		return Kontragent{}, apperror.Internal()
	}
	if reg.Type != region.TypeRegion {
		return Kontragent{}, apperror.BadRequest("Viloyat noto'g'ri tanlangan")
	}

	districtID, err := uuid.Parse(strings.TrimSpace(req.DistrictID))
	if err != nil {
		return Kontragent{}, apperror.BadRequest("Tuman tanlanishi shart")
	}
	dis, err := s.regions.FindByID(ctx, districtID)
	if err != nil {
		if errors.Is(err, region.ErrNotFound) {
			return Kontragent{}, apperror.BadRequest("Tuman topilmadi")
		}
		return Kontragent{}, apperror.Internal()
	}
	if dis.Type != region.TypeDistrict || dis.ParentID == nil || *dis.ParentID != regionID {
		return Kontragent{}, apperror.BadRequest("Tuman tanlangan viloyatga tegishli emas")
	}

	mfyID, err := uuid.Parse(strings.TrimSpace(req.MFYID))
	if err != nil {
		return Kontragent{}, apperror.BadRequest("MFY tanlanishi shart")
	}
	mfy, err := s.regions.FindByID(ctx, mfyID)
	if err != nil {
		if errors.Is(err, region.ErrNotFound) {
			return Kontragent{}, apperror.BadRequest("MFY topilmadi")
		}
		return Kontragent{}, apperror.Internal()
	}
	if mfy.Type != region.TypeMFY || mfy.ParentID == nil || *mfy.ParentID != districtID {
		return Kontragent{}, apperror.BadRequest("MFY tanlangan tumanga tegishli emas")
	}

	if scope != nil && scope.DistrictID != nil && districtID != *scope.DistrictID {
		return Kontragent{}, apperror.Forbidden("Faqat o'z tumaningizdagi kontragentni boshqarish mumkin")
	}
	if scope != nil && scope.MFYID != nil && mfyID != *scope.MFYID {
		return Kontragent{}, apperror.Forbidden("Faqat o'z MFYingizdagi kontragentni boshqarish mumkin")
	}

	hash := existingHash
	if strings.TrimSpace(req.Password) != "" {
		if err = simplePassword(req.Password); err != nil {
			return Kontragent{}, err
		}
		raw, err := bcrypt.GenerateFromPassword([]byte(req.Password), 12)
		if err != nil {
			return Kontragent{}, apperror.Internal()
		}
		hash = string(raw)
	}

	logo := existingLogo
	if req.ClearLogo {
		logo = ""
	} else if strings.TrimSpace(req.Logo) != "" && strings.Contains(req.Logo, ";base64,") {
		saved, err := saveLogo(id, req.Logo)
		if err != nil {
			return Kontragent{}, err
		}
		logo = saved
	}

	return Kontragent{
		ID:             id,
		ActivityTypeID: activityID,
		Name:           req.Name,
		INN:            req.INN,
		RegionID:       regionID,
		DistrictID:     districtID,
		MFYID:          mfyID,
		Phone:          req.Phone,
		Logo:           logo,
		PasswordHash:   hash,
		Status:         req.Status,
	}, nil
}

func applyListScope(q *ListQuery, scope *Scope) {
	if scope == nil {
		return
	}
	q.RegionID = &scope.RegionID
	if scope.DistrictID != nil {
		q.DistrictID = scope.DistrictID
	}
	if scope.MFYID != nil {
		q.MFYID = scope.MFYID
	}
}

func assertInScope(item Kontragent, scope *Scope) error {
	if scope == nil {
		return nil
	}
	if item.RegionID != scope.RegionID {
		return apperror.Forbidden("Faqat o'z viloyatingizdagi kontragentni boshqarish mumkin")
	}
	if scope.DistrictID != nil && item.DistrictID != *scope.DistrictID {
		return apperror.Forbidden("Faqat o'z tumaningizdagi kontragentni boshqarish mumkin")
	}
	if scope.MFYID != nil && item.MFYID != *scope.MFYID {
		return apperror.Forbidden("Faqat o'z MFYingizdagi kontragentni boshqarish mumkin")
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
		if strings.Contains(pgErr.ConstraintName, "inn") {
			return apperror.Conflict("INN allaqachon mavjud")
		}
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
