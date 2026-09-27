package localshop

import (
	"context"
	"errors"
	"regexp"
	"strings"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgconn"
	"golang.org/x/crypto/bcrypt"

	"qoshida/backend/internal/modules/region"
	"qoshida/backend/internal/shared/apperror"
	"qoshida/backend/internal/shared/audit"
)

var phoneRE = regexp.MustCompile(`^\+998[0-9]{9}$`)

type Service struct {
	repo    *Repository
	regions *region.Repository
}

func NewService(repo *Repository, regions *region.Repository) *Service {
	return &Service{repo: repo, regions: regions}
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
	trails := audit.BindMany(ctx, "local_shop", ids)
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
			return Public{}, apperror.NotFound("Mahalla do'koni topilmadi")
		}
		return Public{}, apperror.Internal()
	}
	if err = assertInScope(item, scope); err != nil {
		return Public{}, err
	}
	pub := item.Public()
	pub.Audit = audit.Bind(ctx, "local_shop", item.ID)
	return pub, nil
}

func (s *Service) Create(ctx context.Context, req CreateRequest, scope *Scope) (Public, error) {
	item, err := s.build(ctx, uuid.New(), req, scope, "", "")
	if err != nil {
		return Public{}, err
	}
	created, err := s.repo.Create(ctx, item)
	if err != nil {
		removeImageFile(item.Image)
		return Public{}, mapWriteError(err)
	}
	pub := created.Public()
	pub.Audit = audit.AfterCreate(ctx, "local_shop", created.ID)
	return pub, nil
}

func (s *Service) Update(ctx context.Context, id uuid.UUID, req CreateRequest, scope *Scope) (Public, error) {
	current, err := s.repo.FindByID(ctx, id)
	if err != nil {
		if errors.Is(err, ErrNotFound) {
			return Public{}, apperror.NotFound("Mahalla do'koni topilmadi")
		}
		return Public{}, apperror.Internal()
	}
	if err = assertInScope(current, scope); err != nil {
		return Public{}, err
	}
	item, err := s.build(ctx, id, req, scope, current.PasswordHash, current.Image)
	if err != nil {
		return Public{}, err
	}
	updated, err := s.repo.Update(ctx, item)
	if err != nil {
		return Public{}, mapWriteError(err)
	}
	if current.Image != "" && current.Image != item.Image {
		removeImageFile(current.Image)
	}
	pub := updated.Public()
	pub.Audit = audit.AfterUpdate(ctx, "local_shop", updated.ID)
	return pub, nil
}

func (s *Service) Delete(ctx context.Context, id uuid.UUID, scope *Scope) error {
	current, err := s.repo.FindByID(ctx, id)
	if err != nil {
		if errors.Is(err, ErrNotFound) {
			return apperror.NotFound("Mahalla do'koni topilmadi")
		}
		return apperror.Internal()
	}
	if err = assertInScope(current, scope); err != nil {
		return err
	}
	if err = s.repo.Delete(ctx, id); err != nil {
		if errors.Is(err, ErrNotFound) {
			return apperror.NotFound("Mahalla do'koni topilmadi")
		}
		return apperror.Internal()
	}
	removeImageFile(current.Image)
	return nil
}

func (s *Service) Stats(ctx context.Context) (Stats, error) {
	stats, err := s.repo.Stats(ctx)
	if err != nil {
		return Stats{}, apperror.Internal()
	}
	return stats, nil
}

func (s *Service) build(ctx context.Context, id uuid.UUID, req CreateRequest, scope *Scope, existingHash, existingImage string) (Shop, error) {
	req.Name = strings.TrimSpace(req.Name)
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
		return Shop{}, apperror.BadRequest("Nomi 2-160 belgi oralig'ida bo'lishi kerak")
	}
	if !phoneRE.MatchString(req.Phone) {
		return Shop{}, apperror.BadRequest("Telefon formati: +998XXXXXXXXX")
	}
	if req.Status != StatusActive && req.Status != "inactive" {
		return Shop{}, apperror.BadRequest("status: active yoki inactive")
	}

	regionID, err := uuid.Parse(strings.TrimSpace(req.RegionID))
	if err != nil {
		return Shop{}, apperror.BadRequest("Viloyat tanlanishi shart")
	}
	reg, err := s.regions.FindByID(ctx, regionID)
	if err != nil {
		if errors.Is(err, region.ErrNotFound) {
			return Shop{}, apperror.BadRequest("Viloyat topilmadi")
		}
		return Shop{}, apperror.Internal()
	}
	if reg.Type != region.TypeRegion {
		return Shop{}, apperror.BadRequest("Viloyat noto'g'ri tanlangan")
	}

	districtID, err := uuid.Parse(strings.TrimSpace(req.DistrictID))
	if err != nil {
		return Shop{}, apperror.BadRequest("Tuman tanlanishi shart")
	}
	dis, err := s.regions.FindByID(ctx, districtID)
	if err != nil {
		if errors.Is(err, region.ErrNotFound) {
			return Shop{}, apperror.BadRequest("Tuman topilmadi")
		}
		return Shop{}, apperror.Internal()
	}
	if dis.Type != region.TypeDistrict || dis.ParentID == nil || *dis.ParentID != regionID {
		return Shop{}, apperror.BadRequest("Tuman tanlangan viloyatga tegishli emas")
	}

	mfyID, err := uuid.Parse(strings.TrimSpace(req.MFYID))
	if err != nil {
		return Shop{}, apperror.BadRequest("MFY tanlanishi shart")
	}
	mfy, err := s.regions.FindByID(ctx, mfyID)
	if err != nil {
		if errors.Is(err, region.ErrNotFound) {
			return Shop{}, apperror.BadRequest("MFY topilmadi")
		}
		return Shop{}, apperror.Internal()
	}
	if mfy.Type != region.TypeMFY || mfy.ParentID == nil || *mfy.ParentID != districtID {
		return Shop{}, apperror.BadRequest("MFY tanlangan tumanga tegishli emas")
	}

	if scope != nil && scope.DistrictID != nil && districtID != *scope.DistrictID {
		return Shop{}, apperror.Forbidden("Faqat o'z tumaningizdagi do'konni boshqarish mumkin")
	}
	if scope != nil && scope.MFYID != nil && mfyID != *scope.MFYID {
		return Shop{}, apperror.Forbidden("Faqat o'z MFYingizdagi do'konni boshqarish mumkin")
	}

	hash := existingHash
	if strings.TrimSpace(req.Password) != "" {
		if err = simplePassword(req.Password); err != nil {
			return Shop{}, err
		}
		raw, err := bcrypt.GenerateFromPassword([]byte(req.Password), 12)
		if err != nil {
			return Shop{}, apperror.Internal()
		}
		hash = string(raw)
	}

	image := existingImage
	if req.ClearImage {
		image = ""
	} else if strings.TrimSpace(req.Image) != "" && strings.Contains(req.Image, ";base64,") {
		saved, err := saveImage(id, req.Image)
		if err != nil {
			return Shop{}, err
		}
		image = saved
	}

	return Shop{
		ID:           id,
		Name:         req.Name,
		Phone:        req.Phone,
		Image:        image,
		RegionID:     regionID,
		DistrictID:   districtID,
		MFYID:        mfyID,
		PasswordHash: hash,
		Status:       req.Status,
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

func assertInScope(item Shop, scope *Scope) error {
	if scope == nil {
		return nil
	}
	if item.RegionID != scope.RegionID {
		return apperror.Forbidden("Faqat o'z viloyatingizdagi do'konni boshqarish mumkin")
	}
	if scope.DistrictID != nil && item.DistrictID != *scope.DistrictID {
		return apperror.Forbidden("Faqat o'z tumaningizdagi do'konni boshqarish mumkin")
	}
	if scope.MFYID != nil && item.MFYID != *scope.MFYID {
		return apperror.Forbidden("Faqat o'z MFYingizdagi do'konni boshqarish mumkin")
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
