package activitytype

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
	if q.Limit < 1 || q.Limit > 200 {
		q.Limit = 50
	}
	q.Query = strings.TrimSpace(q.Query)

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
			return Public{}, apperror.NotFound("Faoliyat turi topilmadi")
		}
		return Public{}, apperror.Internal()
	}
	return item.Public(), nil
}

func (s *Service) Create(ctx context.Context, req CreateRequest) (Public, error) {
	item, err := s.build(uuid.Nil, req, "local:"+uuid.NewString())
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
			return Public{}, apperror.NotFound("Faoliyat turi topilmadi")
		}
		return Public{}, apperror.Internal()
	}
	item, err := s.build(id, req, current.SourceID)
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
			return apperror.NotFound("Faoliyat turi topilmadi")
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

func (s *Service) Import(ctx context.Context, items []ActivityType) (ImportResult, error) {
	if len(items) == 0 {
		return ImportResult{}, apperror.BadRequest("Import qilish uchun ma'lumot yo'q")
	}
	return s.repo.UpsertMany(ctx, items)
}

func (s *Service) build(id uuid.UUID, req CreateRequest, sourceID string) (ActivityType, error) {
	req.Name = strings.TrimSpace(req.Name)
	req.Icon = strings.TrimSpace(req.Icon)
	req.Status = strings.TrimSpace(req.Status)
	if req.Status == "" {
		req.Status = StatusActive
	}
	if len(req.Name) < 2 || len(req.Name) > 160 {
		return ActivityType{}, apperror.BadRequest("Nom 2-160 belgi oralig'ida bo'lishi kerak")
	}
	if len(req.Icon) > 80 {
		return ActivityType{}, apperror.BadRequest("Ikonka nomi 80 belgidan oshmasin")
	}
	if req.Status != StatusActive && req.Status != "inactive" {
		return ActivityType{}, apperror.BadRequest("status: active yoki inactive")
	}
	return ActivityType{
		ID:       id,
		SourceID: sourceID,
		Name:     req.Name,
		Icon:     req.Icon,
		Status:   req.Status,
	}, nil
}
