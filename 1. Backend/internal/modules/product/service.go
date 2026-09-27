package product

import (
	"context"
	"errors"
	"strings"
	"time"

	"github.com/google/uuid"

	"qoshida/backend/internal/shared/apperror"
	"qoshida/backend/internal/shared/audit"
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
	trails := audit.BindMany(ctx, "product", ids)
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
			return Public{}, apperror.NotFound("Mahsulot topilmadi")
		}
		return Public{}, apperror.Internal()
	}
	if err = assertInScope(item, scope); err != nil {
		return Public{}, err
	}
	pub := item.Public()
	pub.Audit = audit.Bind(ctx, "product", item.ID)
	attachReviewAudit(&pub)
	return pub, nil
}

func (s *Service) Create(ctx context.Context, req CreateRequest, scope *Scope) (Public, error) {
	item, err := s.build(ctx, uuid.New(), req, scope, Product{})
	if err != nil {
		return Public{}, err
	}
	if isStaff(scope) {
		item.ApprovalStatus = ApprovalApproved
		item.SubmittedBy = SubmittedStaff
		item.RejectionNote = ""
		stampReview(ctx, &item)
	} else {
		item.ApprovalStatus = ApprovalPending
		item.SubmittedBy = SubmittedKontragent
		item.RejectionNote = ""
	}
	created, err := s.repo.Create(ctx, item)
	if err != nil {
		return Public{}, apperror.Internal()
	}
	pub := created.Public()
	pub.Audit = audit.AfterCreate(ctx, "product", created.ID)
	if item.ApprovalStatus == ApprovalApproved {
		pub.Audit = audit.AfterApprove(ctx, "product", created.ID)
	}
	attachReviewAudit(&pub)
	return pub, nil
}

func (s *Service) Update(ctx context.Context, id uuid.UUID, req CreateRequest, scope *Scope) (Public, error) {
	current, err := s.repo.FindByID(ctx, id)
	if err != nil {
		if errors.Is(err, ErrNotFound) {
			return Public{}, apperror.NotFound("Mahsulot topilmadi")
		}
		return Public{}, apperror.Internal()
	}
	if err = assertInScope(current, scope); err != nil {
		return Public{}, err
	}
	item, err := s.build(ctx, id, req, scope, current)
	if err != nil {
		return Public{}, err
	}
	if isStaff(scope) {
		item.ApprovalStatus = current.ApprovalStatus
		item.SubmittedBy = current.SubmittedBy
		item.RejectionNote = current.RejectionNote
		item.ReviewedByName = current.ReviewedByName
		item.ReviewedByRole = current.ReviewedByRole
		item.ReviewedAt = current.ReviewedAt
	} else {
		item.KontragentID = current.KontragentID
		item.SubmittedBy = current.SubmittedBy
		item.ApprovalStatus = ApprovalPending
		item.RejectionNote = ""
		item.ReviewedByName = ""
		item.ReviewedByRole = ""
		item.ReviewedAt = nil
	}
	updated, err := s.repo.Update(ctx, item)
	if err != nil {
		return Public{}, apperror.Internal()
	}
	pub := updated.Public()
	pub.Audit = audit.AfterUpdate(ctx, "product", updated.ID)
	return pub, nil
}

func (s *Service) Delete(ctx context.Context, id uuid.UUID, scope *Scope) error {
	current, err := s.repo.FindByID(ctx, id)
	if err != nil {
		if errors.Is(err, ErrNotFound) {
			return apperror.NotFound("Mahsulot topilmadi")
		}
		return apperror.Internal()
	}
	if err = assertInScope(current, scope); err != nil {
		return err
	}
	if err = s.repo.Delete(ctx, id); err != nil {
		if errors.Is(err, ErrNotFound) {
			return apperror.NotFound("Mahsulot topilmadi")
		}
		return apperror.Internal()
	}
	removeAllImages(current.Images)
	return nil
}

func (s *Service) Approve(ctx context.Context, id uuid.UUID, scope *Scope) (Public, error) {
	if !isStaff(scope) {
		return Public{}, apperror.Forbidden("Mahsulotni tasdiqlash huquqi yo‘q")
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
	pub.Audit = audit.AfterApprove(ctx, "product", updated.ID)
	attachReviewAudit(&pub)
	return pub, nil
}

func (s *Service) Reject(ctx context.Context, id uuid.UUID, note string, scope *Scope) (Public, error) {
	if !isStaff(scope) {
		return Public{}, apperror.Forbidden("Mahsulotni bekor qilish huquqi yo‘q")
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
	pub.Audit = audit.AfterReject(ctx, "product", updated.ID)
	attachReviewAudit(&pub)
	return pub, nil
}

func (s *Service) requirePending(ctx context.Context, id uuid.UUID, scope *Scope) (Product, error) {
	current, err := s.repo.FindByID(ctx, id)
	if err != nil {
		if errors.Is(err, ErrNotFound) {
			return Product{}, apperror.NotFound("Mahsulot topilmadi")
		}
		return Product{}, apperror.Internal()
	}
	if err = assertInScope(current, scope); err != nil {
		return Product{}, err
	}
	if current.ApprovalStatus != ApprovalPending {
		return Product{}, apperror.BadRequest("Faqat kutilayotgan mahsulotni tasdiqlash yoki bekor qilish mumkin")
	}
	return current, nil
}

func (s *Service) build(ctx context.Context, id uuid.UUID, req CreateRequest, scope *Scope, current Product) (Product, error) {
	req.Name = strings.TrimSpace(req.Name)
	req.Description = strings.TrimSpace(req.Description)
	req.Unit = strings.TrimSpace(req.Unit)
	req.Status = strings.TrimSpace(req.Status)
	if req.Status == "" {
		req.Status = StatusActive
	}
	if scope != nil && scope.KontragentID != nil {
		req.KontragentID = scope.KontragentID.String()
	}
	if len(req.Name) < 2 || len(req.Name) > 120 {
		return Product{}, apperror.BadRequest("Nomi 2-120 belgi oralig‘ida bo‘lishi kerak")
	}
	if len(req.Description) > 2000 {
		return Product{}, apperror.BadRequest("Tavsif 2000 belgidan oshmasin")
	}
	if req.Unit != UnitDona && req.Unit != UnitLitr && req.Unit != UnitKg {
		return Product{}, apperror.BadRequest("Birlik: dona, litr yoki kg")
	}
	if req.Status != StatusActive && req.Status != "inactive" {
		return Product{}, apperror.BadRequest("status: active yoki inactive")
	}
	if req.SalePrice < 0 || req.CostPrice < 0 {
		return Product{}, apperror.BadRequest("Narx manfiy bo‘lishi mumkin emas")
	}
	if req.SalePrice < req.CostPrice {
		return Product{}, apperror.BadRequest("Sotuv narxi asl narxdan kichik bo‘lmasin")
	}
	if req.Quantity < 0 {
		return Product{}, apperror.BadRequest("Miqdor manfiy bo‘lishi mumkin emas")
	}
	if req.UnitSize <= 0 {
		return Product{}, apperror.BadRequest("Birlik o‘lchami 0 dan katta bo‘lishi kerak")
	}
	if req.CommissionPercent < 0 || req.CommissionPercent > 100 {
		return Product{}, apperror.BadRequest("Komissiya 0-100 foiz oralig‘ida bo‘lishi kerak")
	}

	kontragentID, err := uuid.Parse(strings.TrimSpace(req.KontragentID))
	if err != nil {
		return Product{}, apperror.BadRequest("Kontragent tanlanishi shart")
	}
	kontragent, err := s.repo.FindKontragent(ctx, kontragentID)
	if err != nil {
		if errors.Is(err, ErrNotFound) {
			return Product{}, apperror.BadRequest("Kontragent topilmadi")
		}
		return Product{}, apperror.Internal()
	}
	if err = assertKontragentInScope(kontragent, scope); err != nil {
		return Product{}, err
	}

	categoryID, err := uuid.Parse(strings.TrimSpace(req.CategoryID))
	if err != nil {
		return Product{}, apperror.BadRequest("Kategoriya tanlanishi shart")
	}
	subID, err := uuid.Parse(strings.TrimSpace(req.SubcategoryID))
	if err != nil {
		return Product{}, apperror.BadRequest("Subkategoriya tanlanishi shart")
	}
	category, err := s.repo.FindCategory(ctx, categoryID)
	if err != nil {
		if errors.Is(err, ErrNotFound) {
			return Product{}, apperror.BadRequest("Kategoriya topilmadi")
		}
		return Product{}, apperror.Internal()
	}
	if category.ParentID != nil {
		return Product{}, apperror.BadRequest("Asosiy kategoriya tanlang")
	}
	sub, err := s.repo.FindCategory(ctx, subID)
	if err != nil {
		if errors.Is(err, ErrNotFound) {
			return Product{}, apperror.BadRequest("Subkategoriya topilmadi")
		}
		return Product{}, apperror.Internal()
	}
	if sub.ParentID == nil || *sub.ParentID != category.ID {
		return Product{}, apperror.BadRequest("Subkategoriya tanlangan kategoriyaga tegishli emas")
	}

	images, err := saveImages(id, req.Images, current.Images)
	if err != nil {
		return Product{}, err
	}

	return Product{
		ID:                id,
		KontragentID:      kontragent.ID,
		CategoryID:        category.ID,
		SubcategoryID:     sub.ID,
		Name:              req.Name,
		Description:       req.Description,
		SalePrice:         req.SalePrice,
		CostPrice:         req.CostPrice,
		Quantity:          req.Quantity,
		Unit:              req.Unit,
		UnitSize:          req.UnitSize,
		CommissionPercent: req.CommissionPercent,
		Images:            images,
		Status:            req.Status,
	}, nil
}

func applyListScope(q *ListQuery, scope *Scope) {
	if scope == nil {
		return
	}
	if scope.KontragentID != nil {
		q.KontragentID = scope.KontragentID
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

func assertInScope(item Product, scope *Scope) error {
	if scope == nil {
		return nil
	}
	if scope.KontragentID != nil && item.KontragentID != *scope.KontragentID {
		return apperror.Forbidden("Faqat o‘z mahsulotingizni boshqarish mumkin")
	}
	if scope.RegionID != nil && item.RegionID != *scope.RegionID {
		return apperror.Forbidden("Faqat o‘z viloyatingizdagi mahsulotni boshqarish mumkin")
	}
	if scope.DistrictID != nil && item.DistrictID != *scope.DistrictID {
		return apperror.Forbidden("Faqat o‘z tumaningizdagi mahsulotni boshqarish mumkin")
	}
	if scope.MFYID != nil && item.MFYID != *scope.MFYID {
		return apperror.Forbidden("Faqat o‘z MFYingizdagi mahsulotni boshqarish mumkin")
	}
	return nil
}

func assertKontragentInScope(item KontragentRef, scope *Scope) error {
	if scope == nil {
		return nil
	}
	if scope.KontragentID != nil && item.ID != *scope.KontragentID {
		return apperror.Forbidden("Faqat o‘z mahsulotingizni qo‘shish mumkin")
	}
	if scope.RegionID != nil && item.RegionID != *scope.RegionID {
		return apperror.Forbidden("Kontragent sizning viloyatingizga tegishli emas")
	}
	if scope.DistrictID != nil && item.DistrictID != *scope.DistrictID {
		return apperror.Forbidden("Kontragent sizning tumaningizga tegishli emas")
	}
	if scope.MFYID != nil && item.MFYID != *scope.MFYID {
		return apperror.Forbidden("Kontragent sizning MFYingizga tegishli emas")
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

func stampReview(ctx context.Context, item *Product) {
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
