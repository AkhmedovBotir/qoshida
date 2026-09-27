package providerservice

import (
	"context"
	"errors"
	"strings"
	"time"

	"github.com/google/uuid"

	"qoshida/backend/internal/shared/apperror"
	"qoshida/backend/internal/shared/audit"
)

const (
	minName = 2
	maxName = 160
	maxPrice = 99999999999.99
	maxBatch = 50
)

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
	q.Query = strings.TrimSpace(q.Query)
	q.ApprovalStatus = strings.TrimSpace(q.ApprovalStatus)
	applyListScope(&q, scope)
	items, total, err := s.repo.List(ctx, q)
	if err != nil {
		return ListResult{}, apperror.Internal()
	}
	ids := make([]uuid.UUID, 0, len(items))
	for _, item := range items {
		ids = append(ids, item.ID)
	}
	trails := audit.BindMany(ctx, "provider_service", ids)
	out := make([]Public, 0, len(items))
	for _, item := range items {
		pub := item.Public()
		if t, ok := trails[item.ID]; ok {
			pub.Audit = &t
		}
		attachReviewAudit(&pub)
		out = append(out, pub)
	}
	return ListResult{Items: out, Total: total, Page: q.Page, Limit: q.Limit}, nil
}

func (s *Service) Get(ctx context.Context, id uuid.UUID, scope *Scope) (Public, error) {
	item, err := s.repo.FindByID(ctx, id)
	if err != nil {
		if errors.Is(err, ErrNotFound) {
			return Public{}, apperror.NotFound("Xizmat topilmadi")
		}
		return Public{}, apperror.Internal()
	}
	if err = assertInScope(item, scope); err != nil {
		return Public{}, err
	}
	pub := item.Public()
	pub.Audit = audit.Bind(ctx, "provider_service", item.ID)
	attachReviewAudit(&pub)
	return pub, nil
}

func (s *Service) Create(ctx context.Context, req CreateRequest, scope *Scope) (CreateResult, error) {
	rows := req.Items
	if len(rows) == 0 && strings.TrimSpace(req.Name) != "" {
		rows = []ItemRequest{{Name: req.Name, Price: req.Price, Images: req.Images}}
	}
	if len(rows) == 0 {
		return CreateResult{}, apperror.BadRequest("Kamida bitta xizmat kiriting")
	}
	if len(rows) > maxBatch {
		return CreateResult{}, apperror.BadRequest("Bir martada 50 tadan ko‘p xizmat qo‘shib bo‘lmaydi")
	}

	provider, err := s.resolveProvider(ctx, req.ProviderID, scope)
	if err != nil {
		return CreateResult{}, err
	}

	staff := isStaff(scope)
	if !staff && provider.IdentificationStatus != ApprovalApproved {
		return CreateResult{}, apperror.Forbidden("Identifikatsiyadan o‘tmaguncha xizmat joylab bo‘lmaydi")
	}
	items := make([]ServiceItem, 0, len(rows))
	for _, row := range rows {
		name, price, err := validateItem(row.Name, row.Price)
		if err != nil {
			return CreateResult{}, err
		}
		item := ServiceItem{
			ID:         uuid.New(),
			ProviderID: provider.ID,
			Name:       name,
			Price:      price,
		}
		images, err := saveImages(item.ID, row.Images, nil)
		if err != nil {
			return CreateResult{}, err
		}
		item.Images = images
		if staff {
			item.ApprovalStatus = ApprovalApproved
			item.SubmittedBy = SubmittedStaff
			stampReview(ctx, &item)
		} else {
			item.ApprovalStatus = ApprovalPending
			item.SubmittedBy = SubmittedProvider
		}
		items = append(items, item)
	}

	created, err := s.repo.CreateMany(ctx, items)
	if err != nil {
		return CreateResult{}, apperror.Internal()
	}
	out := make([]Public, 0, len(created))
	for _, item := range created {
		pub := item.Public()
		pub.Audit = audit.AfterCreate(ctx, "provider_service", item.ID)
		if item.ApprovalStatus == ApprovalApproved {
			pub.Audit = audit.AfterApprove(ctx, "provider_service", item.ID)
		}
		attachReviewAudit(&pub)
		out = append(out, pub)
	}
	return CreateResult{Items: out}, nil
}

func (s *Service) Update(ctx context.Context, id uuid.UUID, req UpdateRequest, scope *Scope) (Public, error) {
	current, err := s.repo.FindByID(ctx, id)
	if err != nil {
		if errors.Is(err, ErrNotFound) {
			return Public{}, apperror.NotFound("Xizmat topilmadi")
		}
		return Public{}, apperror.Internal()
	}
	if err = assertInScope(current, scope); err != nil {
		return Public{}, err
	}
	if !isStaff(scope) {
		provider, err := s.repo.FindProvider(ctx, current.ProviderID)
		if err != nil {
			return Public{}, apperror.Internal()
		}
		if provider.IdentificationStatus != ApprovalApproved {
			return Public{}, apperror.Forbidden("Identifikatsiyadan o‘tmaguncha xizmat joylab bo‘lmaydi")
		}
	}
	name, price, err := validateItem(req.Name, req.Price)
	if err != nil {
		return Public{}, err
	}
	images, err := saveImages(current.ID, req.Images, current.Images)
	if err != nil {
		return Public{}, err
	}
	current.Name = name
	current.Price = price
	current.Images = images
	if !isStaff(scope) {
		current.ApprovalStatus = ApprovalPending
		current.RejectionNote = ""
		current.ReviewedByName = ""
		current.ReviewedByRole = ""
		current.ReviewedAt = nil
	}
	updated, err := s.repo.Update(ctx, current)
	if err != nil {
		return Public{}, apperror.Internal()
	}
	pub := updated.Public()
	pub.Audit = audit.AfterUpdate(ctx, "provider_service", updated.ID)
	return pub, nil
}

func (s *Service) Delete(ctx context.Context, id uuid.UUID, scope *Scope) error {
	current, err := s.repo.FindByID(ctx, id)
	if err != nil {
		if errors.Is(err, ErrNotFound) {
			return apperror.NotFound("Xizmat topilmadi")
		}
		return apperror.Internal()
	}
	if err = assertInScope(current, scope); err != nil {
		return err
	}
	if err = s.repo.Delete(ctx, id); err != nil {
		if errors.Is(err, ErrNotFound) {
			return apperror.NotFound("Xizmat topilmadi")
		}
		return apperror.Internal()
	}
	removeAllImages(current.Images)
	return nil
}

func (s *Service) Approve(ctx context.Context, id uuid.UUID, scope *Scope) (Public, error) {
	if !isStaff(scope) {
		return Public{}, apperror.Forbidden("Xizmatni tasdiqlash huquqi yo‘q")
	}
	current, err := s.requirePending(ctx, id, scope)
	if err != nil {
		return Public{}, err
	}
	current.ApprovalStatus = ApprovalApproved
	current.RejectionNote = ""
	stampReview(ctx, &current)
	updated, err := s.repo.Update(ctx, current)
	if err != nil {
		return Public{}, apperror.Internal()
	}
	pub := updated.Public()
	pub.Audit = audit.AfterApprove(ctx, "provider_service", updated.ID)
	attachReviewAudit(&pub)
	return pub, nil
}

func (s *Service) Reject(ctx context.Context, id uuid.UUID, note string, scope *Scope) (Public, error) {
	if !isStaff(scope) {
		return Public{}, apperror.Forbidden("Xizmatni bekor qilish huquqi yo‘q")
	}
	note = strings.TrimSpace(note)
	if len(note) < 3 {
		return Public{}, apperror.BadRequest("Bekor qilish sababi kamida 3 belgi bo‘lishi kerak")
	}
	if len(note) > 500 {
		return Public{}, apperror.BadRequest("Bekor qilish sababi 500 belgidan oshmasin")
	}
	current, err := s.requirePending(ctx, id, scope)
	if err != nil {
		return Public{}, err
	}
	current.ApprovalStatus = ApprovalRejected
	current.RejectionNote = note
	stampReview(ctx, &current)
	updated, err := s.repo.Update(ctx, current)
	if err != nil {
		return Public{}, apperror.Internal()
	}
	pub := updated.Public()
	pub.Audit = audit.AfterReject(ctx, "provider_service", updated.ID)
	attachReviewAudit(&pub)
	return pub, nil
}

func (s *Service) requirePending(ctx context.Context, id uuid.UUID, scope *Scope) (ServiceItem, error) {
	current, err := s.repo.FindByID(ctx, id)
	if err != nil {
		if errors.Is(err, ErrNotFound) {
			return ServiceItem{}, apperror.NotFound("Xizmat topilmadi")
		}
		return ServiceItem{}, apperror.Internal()
	}
	if err = assertInScope(current, scope); err != nil {
		return ServiceItem{}, err
	}
	if current.ApprovalStatus != ApprovalPending {
		return ServiceItem{}, apperror.BadRequest("Faqat kutilayotgan xizmatni tasdiqlash yoki bekor qilish mumkin")
	}
	return current, nil
}

func (s *Service) resolveProvider(ctx context.Context, raw string, scope *Scope) (ProviderRef, error) {
	if scope != nil && scope.ProviderID != nil {
		raw = scope.ProviderID.String()
	}
	id, err := uuid.Parse(strings.TrimSpace(raw))
	if err != nil {
		return ProviderRef{}, apperror.BadRequest("Xizmat ko‘rsatuvchi tanlanishi shart")
	}
	provider, err := s.repo.FindProvider(ctx, id)
	if err != nil {
		if errors.Is(err, ErrNotFound) {
			return ProviderRef{}, apperror.BadRequest("Xizmat ko‘rsatuvchi topilmadi")
		}
		return ProviderRef{}, apperror.Internal()
	}
	if err = assertProviderInScope(provider, scope); err != nil {
		return ProviderRef{}, err
	}
	return provider, nil
}

func validateItem(name string, price float64) (string, float64, error) {
	name = strings.TrimSpace(name)
	if len(name) < minName || len(name) > maxName {
		return "", 0, apperror.BadRequest("Xizmat nomi 2-160 belgi oralig‘ida bo‘lishi kerak")
	}
	if price < 0 {
		return "", 0, apperror.BadRequest("Narx manfiy bo‘lishi mumkin emas")
	}
	if price > maxPrice {
		return "", 0, apperror.BadRequest("Narx juda katta")
	}
	return name, roundMoney(price), nil
}

func roundMoney(v float64) float64 {
	return float64(int(v*100+0.5)) / 100
}

func applyListScope(q *ListQuery, scope *Scope) {
	if scope == nil {
		return
	}
	if scope.ProviderID != nil {
		q.ProviderID = scope.ProviderID
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

func assertInScope(item ServiceItem, scope *Scope) error {
	if scope == nil {
		return nil
	}
	if scope.ProviderID != nil && item.ProviderID != *scope.ProviderID {
		return apperror.Forbidden("Faqat o‘z xizmatingizni boshqarish mumkin")
	}
	if scope.RegionID != nil && item.RegionID != *scope.RegionID {
		return apperror.Forbidden("Faqat o‘z viloyatingizdagi xizmatni boshqarish mumkin")
	}
	if scope.DistrictID != nil && item.DistrictID != *scope.DistrictID {
		return apperror.Forbidden("Faqat o‘z tumaningizdagi xizmatni boshqarish mumkin")
	}
	if scope.MFYID != nil && item.MFYID != *scope.MFYID {
		return apperror.Forbidden("Faqat o‘z MFYingizdagi xizmatni boshqarish mumkin")
	}
	return nil
}

func assertProviderInScope(item ProviderRef, scope *Scope) error {
	if scope == nil {
		return nil
	}
	if scope.ProviderID != nil && item.ID != *scope.ProviderID {
		return apperror.Forbidden("Faqat o‘z xizmatingizni qo‘shish mumkin")
	}
	if scope.RegionID != nil && item.RegionID != *scope.RegionID {
		return apperror.Forbidden("Xizmat ko‘rsatuvchi sizning viloyatingizga tegishli emas")
	}
	if scope.DistrictID != nil && item.DistrictID != *scope.DistrictID {
		return apperror.Forbidden("Xizmat ko‘rsatuvchi sizning tumaningizga tegishli emas")
	}
	if scope.MFYID != nil && item.MFYID != *scope.MFYID {
		return apperror.Forbidden("Xizmat ko‘rsatuvchi sizning MFYingizga tegishli emas")
	}
	return nil
}

func attachReviewAudit(pub *Public) {
	if pub.ReviewedByName == "" {
		return
	}
	if pub.Audit == nil {
		pub.Audit = &audit.Trail{}
	}
	entry := &audit.Entry{Name: pub.ReviewedByName, Role: pub.ReviewedByRole, At: pub.ReviewedAt}
	if pub.ApprovalStatus == ApprovalRejected && pub.Audit.Rejected == nil {
		pub.Audit.Rejected = entry
	}
	if pub.ApprovalStatus == ApprovalApproved && pub.Audit.Approved == nil {
		pub.Audit.Approved = entry
	}
}

func stampReview(ctx context.Context, item *ServiceItem) {
	now := time.Now().UTC()
	item.ReviewedAt = &now
	if actor, ok := audit.From(ctx); ok {
		item.ReviewedByName = actor.Name
		item.ReviewedByRole = actor.Role
		return
	}
	item.ReviewedByName = "Tizim"
	item.ReviewedByRole = "Admin"
}
