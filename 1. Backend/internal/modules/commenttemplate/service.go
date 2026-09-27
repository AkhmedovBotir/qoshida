package commenttemplate

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
	if q.Limit < 1 {
		q.Limit = 10
	}
	if q.Limit > 100 {
		q.Limit = 100
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
	totalPages := (total + q.Limit - 1) / q.Limit
	if totalPages == 0 {
		totalPages = 1
	}
	return ListResult{Items: out, Total: total, Page: q.Page, Limit: q.Limit, TotalPages: totalPages}, nil
}

func (s *Service) Get(ctx context.Context, id uuid.UUID) (Public, error) {
	item, err := s.repo.FindByID(ctx, id)
	if err != nil {
		if errors.Is(err, ErrNotFound) {
			return Public{}, apperror.NotFound("Kommentariya shabloni topilmadi")
		}
		return Public{}, apperror.Internal()
	}
	return item.Public(), nil
}

func (s *Service) Create(ctx context.Context, req CreateRequest) (Public, error) {
	item, err := s.build(uuid.Nil, req)
	if err != nil {
		return Public{}, err
	}
	maxOrder, err := s.repo.MaxSortOrder(ctx)
	if err != nil {
		return Public{}, apperror.Internal()
	}
	item.SortOrder = maxOrder + 1
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
			return Public{}, apperror.NotFound("Kommentariya shabloni topilmadi")
		}
		return Public{}, apperror.Internal()
	}
	item, err := s.build(id, req)
	if err != nil {
		return Public{}, err
	}
	if strings.TrimSpace(req.Status) == "" {
		item.Status = current.Status
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
			return apperror.NotFound("Kommentariya shabloni topilmadi")
		}
		return apperror.Internal()
	}
	return nil
}

func (s *Service) Reorder(ctx context.Context, req ReorderRequest) error {
	fromID, err := uuid.Parse(strings.TrimSpace(req.FromID))
	if err != nil {
		return apperror.BadRequest("from_id noto'g'ri")
	}
	toID, err := uuid.Parse(strings.TrimSpace(req.ToID))
	if err != nil {
		return apperror.BadRequest("to_id noto'g'ri")
	}
	if fromID == toID {
		return apperror.BadRequest("from_id va to_id bir xil bo'lmasin")
	}
	if err = s.repo.SwapSortOrder(ctx, fromID, toID); err != nil {
		if errors.Is(err, ErrNotFound) {
			return apperror.NotFound("Kommentariya shabloni topilmadi")
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

func (s *Service) build(id uuid.UUID, req CreateRequest) (CommentTemplate, error) {
	req.Comment = strings.TrimSpace(req.Comment)
	req.Status = strings.TrimSpace(req.Status)
	if req.Status == "" {
		req.Status = StatusActive
	}
	if req.Comment == "" {
		return CommentTemplate{}, apperror.BadRequest("Kommentariya matni majburiy")
	}
	if len(req.Comment) > 4000 {
		return CommentTemplate{}, apperror.BadRequest("Kommentariya 4000 belgidan oshmasin")
	}
	if req.Status != StatusActive && req.Status != "inactive" {
		return CommentTemplate{}, apperror.BadRequest("status: active yoki inactive")
	}
	return CommentTemplate{
		ID:      id,
		Comment: req.Comment,
		Status:  req.Status,
	}, nil
}
