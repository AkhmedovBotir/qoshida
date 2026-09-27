package shoptemplate

import (
	"context"
	"errors"
	"strings"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgconn"

	"qoshida/backend/internal/shared/apperror"
	"qoshida/backend/internal/shared/audit"
)

const entity = "shop_product_template"

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
	if q.Limit < 1 || q.Limit > 100 {
		q.Limit = 20
	}
	q.Query = strings.TrimSpace(q.Query)
	items, total, err := s.repo.List(ctx, q)
	if err != nil {
		return ListResult{}, apperror.Internal()
	}
	ids := make([]uuid.UUID, 0, len(items))
	for _, item := range items {
		ids = append(ids, item.ID)
	}
	trails := audit.BindMany(ctx, entity, ids)
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
			return Public{}, apperror.NotFound("Shablon topilmadi")
		}
		return Public{}, apperror.Internal()
	}
	pub := item.Public()
	pub.Audit = audit.Bind(ctx, entity, item.ID)
	return pub, nil
}

func (s *Service) Create(ctx context.Context, req CreateRequest) (Public, error) {
	item, err := s.build(ctx, uuid.New(), req, Template{})
	if err != nil {
		return Public{}, err
	}
	created, err := s.repo.Create(ctx, item)
	if err != nil {
		return Public{}, apperror.Internal()
	}
	pub := created.Public()
	pub.Audit = audit.AfterCreate(ctx, entity, created.ID)
	return pub, nil
}

func (s *Service) Update(ctx context.Context, id uuid.UUID, req CreateRequest) (Public, error) {
	current, err := s.repo.FindByID(ctx, id)
	if err != nil {
		if errors.Is(err, ErrNotFound) {
			return Public{}, apperror.NotFound("Shablon topilmadi")
		}
		return Public{}, apperror.Internal()
	}
	item, err := s.build(ctx, id, req, current)
	if err != nil {
		return Public{}, err
	}
	updated, err := s.repo.Update(ctx, item)
	if err != nil {
		return Public{}, apperror.Internal()
	}
	pub := updated.Public()
	pub.Audit = audit.AfterUpdate(ctx, entity, updated.ID)
	return pub, nil
}

func (s *Service) Delete(ctx context.Context, id uuid.UUID) error {
	current, err := s.repo.FindByID(ctx, id)
	if err != nil {
		if errors.Is(err, ErrNotFound) {
			return apperror.NotFound("Shablon topilmadi")
		}
		return apperror.Internal()
	}
	if err = s.repo.Delete(ctx, id); err != nil {
		if isFK(err) {
			return apperror.Conflict("Shablon do‘konlarga biriktirilgan. Avval biriktirishni olib tashlang")
		}
		if errors.Is(err, ErrNotFound) {
			return apperror.NotFound("Shablon topilmadi")
		}
		return apperror.Internal()
	}
	removeAllImages(current.Images)
	return nil
}

func (s *Service) build(ctx context.Context, id uuid.UUID, req CreateRequest, current Template) (Template, error) {
	req.Name = strings.TrimSpace(req.Name)
	req.Description = strings.TrimSpace(req.Description)
	req.Unit = strings.TrimSpace(req.Unit)
	if len(req.Name) < 2 || len(req.Name) > 120 {
		return Template{}, apperror.BadRequest("Nomi 2-120 belgi oralig‘ida bo‘lishi kerak")
	}
	if len(req.Description) > 2000 {
		return Template{}, apperror.BadRequest("Tavsif 2000 belgidan oshmasin")
	}
	if req.Unit != UnitDona && req.Unit != UnitLitr && req.Unit != UnitKg {
		return Template{}, apperror.BadRequest("Birlik: dona, litr yoki kg")
	}
	if req.UnitSize <= 0 {
		return Template{}, apperror.BadRequest("Birlik o‘lchami 0 dan katta bo‘lishi kerak")
	}

	categoryID, err := uuid.Parse(strings.TrimSpace(req.CategoryID))
	if err != nil {
		return Template{}, apperror.BadRequest("Kategoriya tanlanishi shart")
	}
	subID, err := uuid.Parse(strings.TrimSpace(req.SubcategoryID))
	if err != nil {
		return Template{}, apperror.BadRequest("Subkategoriya tanlanishi shart")
	}
	category, err := s.repo.FindCategory(ctx, categoryID)
	if err != nil {
		if errors.Is(err, ErrNotFound) {
			return Template{}, apperror.BadRequest("Kategoriya topilmadi")
		}
		return Template{}, apperror.Internal()
	}
	if category.ParentID != nil {
		return Template{}, apperror.BadRequest("Asosiy kategoriya tanlang")
	}
	sub, err := s.repo.FindCategory(ctx, subID)
	if err != nil {
		if errors.Is(err, ErrNotFound) {
			return Template{}, apperror.BadRequest("Subkategoriya topilmadi")
		}
		return Template{}, apperror.Internal()
	}
	if sub.ParentID == nil || *sub.ParentID != category.ID {
		return Template{}, apperror.BadRequest("Subkategoriya tanlangan kategoriyaga tegishli emas")
	}

	images, err := saveImages(id, req.Images, current.Images)
	if err != nil {
		return Template{}, err
	}

	return Template{
		ID:            id,
		CategoryID:    category.ID,
		SubcategoryID: sub.ID,
		Name:          req.Name,
		Description:   req.Description,
		Unit:          req.Unit,
		UnitSize:      req.UnitSize,
		Images:        images,
	}, nil
}

func isFK(err error) bool {
	var pgErr *pgconn.PgError
	return errors.As(err, &pgErr) && pgErr.Code == "23503"
}
