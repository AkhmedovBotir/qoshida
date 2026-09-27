package region

import (
	"context"
	"errors"
	"strings"

	"github.com/google/uuid"

	"qoshida/backend/internal/shared/apperror"
)

type Service struct {
	repo *Repository
}

func NewService(repo *Repository) *Service {
	return &Service{repo: repo}
}

func (s *Service) List(ctx context.Context, q ListQuery) (ListResult, error) {
	if q.Page < 1 {
		q.Page = 1
	}
	if q.Limit < 1 || q.Limit > 300 {
		q.Limit = 20
	}
	q.Type = strings.TrimSpace(q.Type)
	q.Query = strings.TrimSpace(q.Query)
	if q.Type != "" && q.Type != TypeRegion && q.Type != TypeDistrict && q.Type != TypeMFY {
		return ListResult{}, apperror.BadRequest("type: region, district yoki mfy")
	}

	items, total, err := s.repo.List(ctx, q)
	if err != nil {
		return ListResult{}, apperror.Internal()
	}
	out := make([]Public, 0, len(items))
	for _, item := range items {
		out = append(out, item.Public())
	}
	return ListResult{Items: out, Total: total, Page: q.Page, Limit: q.Limit}, nil
}

func (s *Service) Get(ctx context.Context, id uuid.UUID) (Public, error) {
	item, err := s.repo.FindByID(ctx, id)
	if err != nil {
		if errors.Is(err, ErrNotFound) {
			return Public{}, apperror.NotFound("Hudud topilmadi")
		}
		return Public{}, apperror.Internal()
	}
	return item.Public(), nil
}

func (s *Service) Create(ctx context.Context, req CreateRequest) (Public, error) {
	item, err := s.build(ctx, uuid.Nil, req, "local:"+uuid.NewString())
	if err != nil {
		return Public{}, err
	}
	created, err := s.repo.Create(ctx, item)
	if err != nil {
		return Public{}, apperror.Internal()
	}
	return created.Public(), nil
}

func (s *Service) Update(ctx context.Context, id uuid.UUID, req UpdateRequest) (Public, error) {
	current, err := s.repo.FindByID(ctx, id)
	if err != nil {
		if errors.Is(err, ErrNotFound) {
			return Public{}, apperror.NotFound("Hudud topilmadi")
		}
		return Public{}, apperror.Internal()
	}
	item, err := s.build(ctx, id, req, current.SourceID)
	if err != nil {
		return Public{}, err
	}
	updated, err := s.repo.Update(ctx, item)
	if err != nil {
		return Public{}, apperror.Internal()
	}
	return updated.Public(), nil
}

func (s *Service) Delete(ctx context.Context, id uuid.UUID) error {
	if err := s.repo.Delete(ctx, id); err != nil {
		if errors.Is(err, ErrNotFound) {
			return apperror.NotFound("Hudud topilmadi")
		}
		if errors.Is(err, errHasChildren) {
			return apperror.Conflict("Avval bolalar hududlarini o'chiring")
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

func (s *Service) Import(ctx context.Context, items []Region) (ImportResult, error) {
	if len(items) == 0 {
		return ImportResult{}, apperror.BadRequest("Import qilish uchun ma'lumot yo'q")
	}
	result, err := s.repo.UpsertMany(ctx, items)
	if err != nil {
		return ImportResult{}, err
	}
	return result, nil
}

func (s *Service) build(ctx context.Context, id uuid.UUID, req CreateRequest, sourceID string) (Region, error) {
	req.Name = strings.TrimSpace(req.Name)
	req.Code = strings.TrimSpace(req.Code)
	req.Type = strings.TrimSpace(req.Type)
	req.Status = strings.TrimSpace(req.Status)
	if req.Status == "" {
		req.Status = StatusActive
	}
	if len(req.Name) < 2 || len(req.Name) > 160 {
		return Region{}, apperror.BadRequest("Nom 2-160 belgi oralig'ida bo'lishi kerak")
	}
	if len(req.Code) < 1 || len(req.Code) > 80 {
		return Region{}, apperror.BadRequest("Code 1-80 belgi oralig'ida bo'lishi kerak")
	}
	if req.Type != TypeRegion && req.Type != TypeDistrict && req.Type != TypeMFY {
		return Region{}, apperror.BadRequest("type: region, district yoki mfy")
	}
	if req.Status != StatusActive && req.Status != "inactive" {
		return Region{}, apperror.BadRequest("status: active yoki inactive")
	}

	item := Region{
		ID:       id,
		SourceID: sourceID,
		Name:     req.Name,
		Type:     req.Type,
		Code:     req.Code,
		Status:   req.Status,
	}

	if req.Type == TypeRegion {
		if req.ParentID != nil && strings.TrimSpace(*req.ParentID) != "" {
			return Region{}, apperror.BadRequest("Viloyatning parenti bo'lmasligi kerak")
		}
		return item, nil
	}
	if req.ParentID == nil || strings.TrimSpace(*req.ParentID) == "" {
		return Region{}, apperror.BadRequest("Tuman va MFY uchun parent majburiy")
	}
	parentID, err := uuid.Parse(strings.TrimSpace(*req.ParentID))
	if err != nil {
		return Region{}, apperror.BadRequest("parent_id noto'g'ri")
	}
	if parentID == id {
		return Region{}, apperror.BadRequest("Hudud o'ziga parent bo'la olmaydi")
	}
	parent, err := s.repo.FindByID(ctx, parentID)
	if err != nil {
		return Region{}, apperror.BadRequest("Parent hudud topilmadi")
	}
	if req.Type == TypeDistrict && parent.Type != TypeRegion {
		return Region{}, apperror.BadRequest("Tuman parenti viloyat bo'lishi kerak")
	}
	if req.Type == TypeMFY && parent.Type != TypeDistrict {
		return Region{}, apperror.BadRequest("MFY parenti tuman bo'lishi kerak")
	}
	item.ParentID = &parentID
	item.ParentSourceID = parent.SourceID
	return item, nil
}
