package category

import (
	"context"
	"errors"
	"regexp"
	"strings"
	"unicode"

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

func (s *Service) List(ctx context.Context, q ListQuery) (ListResult, error) {
	if q.Page < 1 {
		q.Page = 1
	}
	if q.Limit < 1 || q.Limit > 300 {
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
	trails := audit.BindMany(ctx, "category", ids)
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
			return Public{}, apperror.NotFound("Kategoriya topilmadi")
		}
		return Public{}, apperror.Internal()
	}
	pub := item.Public()
	pub.Audit = audit.Bind(ctx, "category", item.ID)
	return pub, nil
}

func (s *Service) Create(ctx context.Context, req CreateRequest) (Public, error) {
	item, err := s.build(ctx, uuid.Nil, req, "local:"+uuid.NewString(), "")
	if err != nil {
		return Public{}, err
	}
	created, err := s.repo.Create(ctx, item)
	if err != nil {
		return Public{}, apperror.Internal()
	}
	pub := created.Public()
	pub.Audit = audit.AfterCreate(ctx, "category", created.ID)
	return pub, nil
}

func (s *Service) Update(ctx context.Context, id uuid.UUID, req UpdateRequest) (Public, error) {
	current, err := s.repo.FindByID(ctx, id)
	if err != nil {
		if errors.Is(err, ErrNotFound) {
			return Public{}, apperror.NotFound("Kategoriya topilmadi")
		}
		return Public{}, apperror.Internal()
	}
	item, err := s.build(ctx, id, req, current.SourceID, current.ImageURL)
	if err != nil {
		return Public{}, err
	}
	updated, err := s.repo.Update(ctx, item)
	if err != nil {
		return Public{}, apperror.Internal()
	}
	pub := updated.Public()
	pub.Audit = audit.AfterUpdate(ctx, "category", updated.ID)
	return pub, nil
}

func (s *Service) Delete(ctx context.Context, id uuid.UUID) error {
	if err := s.repo.Delete(ctx, id); err != nil {
		if errors.Is(err, ErrNotFound) {
			return apperror.NotFound("Kategoriya topilmadi")
		}
		if errors.Is(err, errHasChildren) {
			return apperror.Conflict("Avval ichki kategoriyalarni o'chiring")
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

func (s *Service) Import(ctx context.Context, items []Category) (ImportResult, error) {
	if len(items) == 0 {
		return ImportResult{}, apperror.BadRequest("Import qilish uchun ma'lumot yo'q")
	}
	result, err := s.repo.UpsertMany(ctx, items)
	if err != nil {
		return ImportResult{}, err
	}
	for _, item := range items {
		if item.ImageURL != "" {
			result.Images++
		}
	}
	return result, nil
}

func (s *Service) build(ctx context.Context, id uuid.UUID, req CreateRequest, sourceID, existingImage string) (Category, error) {
	req.Name = strings.TrimSpace(req.Name)
	req.Slug = strings.TrimSpace(req.Slug)
	req.Status = strings.TrimSpace(req.Status)
	if req.Status == "" {
		req.Status = StatusActive
	}
	if len(req.Name) < 2 || len(req.Name) > 160 {
		return Category{}, apperror.BadRequest("Nom 2-160 belgi oralig'ida bo'lishi kerak")
	}
	if req.Slug == "" {
		req.Slug = slugify(req.Name)
	}
	if len(req.Slug) < 1 || len(req.Slug) > 180 {
		return Category{}, apperror.BadRequest("Slug 1-180 belgi oralig'ida bo'lishi kerak")
	}
	if req.Status != StatusActive && req.Status != "inactive" {
		return Category{}, apperror.BadRequest("status: active yoki inactive")
	}

	item := Category{
		ID:       id,
		SourceID: sourceID,
		Name:     req.Name,
		Slug:     req.Slug,
		Censored: req.Censored,
		Status:   req.Status,
	}
	if err := applyCategoryImage(&item, req, existingImage); err != nil {
		return Category{}, err
	}

	if req.ParentID == nil || strings.TrimSpace(*req.ParentID) == "" {
		return item, nil
	}
	parentID, err := uuid.Parse(strings.TrimSpace(*req.ParentID))
	if err != nil {
		return Category{}, apperror.BadRequest("parent_id noto'g'ri")
	}
	if parentID == id {
		return Category{}, apperror.BadRequest("Kategoriya o'ziga parent bo'la olmaydi")
	}
	parent, err := s.repo.FindByID(ctx, parentID)
	if err != nil {
		return Category{}, apperror.BadRequest("Parent kategoriya topilmadi")
	}
	if parent.ParentID != nil {
		return Category{}, apperror.BadRequest("Ichki kategoriya ostiga yana kategoriya qo'shib bo'lmaydi")
	}
	item.ParentID = &parentID
	item.ParentSourceID = parent.SourceID
	return item, nil
}

func applyCategoryImage(item *Category, req CreateRequest, existing string) error {
	if req.ClearImage {
		item.ImageURL = ""
		return nil
	}
	raw := strings.TrimSpace(req.Image)
	if raw == "" {
		item.ImageURL = existing
		return nil
	}
	url, err := saveDataImage(item.SourceID, raw, 10*1024*1024)
	if err != nil {
		return apperror.BadRequest(err.Error())
	}
	item.ImageURL = url
	return nil
}

var nonSlug = regexp.MustCompile(`[^a-z0-9]+`)

func slugify(name string) string {
	var b strings.Builder
	for _, r := range strings.ToLower(name) {
		if unicode.IsLetter(r) || unicode.IsDigit(r) {
			switch r {
			case 'o', 'ö':
				b.WriteByte('o')
			default:
				b.WriteRune(r)
			}
			continue
		}
		if r == ' ' || r == '-' || r == '_' {
			b.WriteByte('-')
		}
	}
	s := nonSlug.ReplaceAllString(b.String(), "-")
	s = strings.Trim(s, "-")
	if s == "" {
		s = "kategoriya"
	}
	if len(s) > 180 {
		s = s[:180]
	}
	return s
}
