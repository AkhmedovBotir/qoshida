package shopstock

import (
	"context"
	"errors"
	"strings"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgconn"

	"qoshida/backend/internal/shared/apperror"
	"qoshida/backend/internal/shared/audit"
)

const (
	productEntity  = "shop_product"
	incomingEntity = "shop_product_incoming"
)

type Service struct {
	repo *Repository
}

func NewService(repo *Repository) *Service {
	return &Service{repo: repo}
}

func (s *Service) ListProducts(ctx context.Context, q ProductListQuery, scope *Scope) (ProductListResult, error) {
	normalizePage(&q.Page, &q.Limit)
	q.Query = strings.TrimSpace(q.Query)
	applyProductScope(&q, scope)
	items, total, err := s.repo.ListProducts(ctx, q)
	if err != nil {
		return ProductListResult{}, apperror.Internal()
	}
	ids := make([]uuid.UUID, 0, len(items))
	for _, item := range items {
		ids = append(ids, item.ID)
	}
	trails := audit.BindMany(ctx, productEntity, ids)
	out := make([]ProductPublic, 0, len(items))
	for _, item := range items {
		pub := item.Public()
		if t, ok := trails[item.ID]; ok {
			pub.Audit = &t
		}
		out = append(out, pub)
	}
	return ProductListResult{Items: out, Total: total, Page: q.Page, Limit: q.Limit}, nil
}

func (s *Service) GetProduct(ctx context.Context, id uuid.UUID, scope *Scope) (ProductPublic, error) {
	item, err := s.requireProduct(ctx, id, scope)
	if err != nil {
		return ProductPublic{}, err
	}
	pub := item.Public()
	pub.Audit = audit.Bind(ctx, productEntity, item.ID)
	return pub, nil
}

func (s *Service) Attach(ctx context.Context, req AttachRequest, scope *Scope) (ProductPublic, error) {
	if scope != nil && scope.ShopID != nil {
		req.ShopID = scope.ShopID.String()
	}
	shopID, err := uuid.Parse(strings.TrimSpace(req.ShopID))
	if err != nil {
		return ProductPublic{}, apperror.BadRequest("Do‘kon tanlanishi shart")
	}
	templateID, err := uuid.Parse(strings.TrimSpace(req.TemplateID))
	if err != nil {
		return ProductPublic{}, apperror.BadRequest("Shablon tanlanishi shart")
	}
	if req.Quantity < 0 {
		return ProductPublic{}, apperror.BadRequest("Miqdor manfiy bo‘lishi mumkin emas")
	}
	if req.SalePrice < 0 || req.CostPrice < 0 {
		return ProductPublic{}, apperror.BadRequest("Narx manfiy bo‘lishi mumkin emas")
	}
	if req.SalePrice < req.CostPrice {
		return ProductPublic{}, apperror.BadRequest("Sotuv narxi asl narxdan kichik bo‘lmasin")
	}

	shop, err := s.repo.FindShop(ctx, shopID)
	if err != nil {
		if errors.Is(err, ErrNotFound) {
			return ProductPublic{}, apperror.BadRequest("Do‘kon topilmadi")
		}
		return ProductPublic{}, apperror.Internal()
	}
	if err = assertShopInScope(shop, scope); err != nil {
		return ProductPublic{}, err
	}
	if _, err = s.repo.FindTemplate(ctx, templateID); err != nil {
		if errors.Is(err, ErrNotFound) {
			return ProductPublic{}, apperror.BadRequest("Shablon topilmadi")
		}
		return ProductPublic{}, apperror.Internal()
	}
	if _, err = s.repo.FindByShopAndTemplate(ctx, shopID, templateID); err == nil {
		return ProductPublic{}, apperror.Conflict("Bu shablon allaqachon do‘konga biriktirilgan")
	} else if !errors.Is(err, ErrNotFound) {
		return ProductPublic{}, apperror.Internal()
	}

	created, err := s.repo.CreateProduct(ctx, ShopProduct{
		ID:         uuid.New(),
		ShopID:     shopID,
		TemplateID: templateID,
		Quantity:   req.Quantity,
		SalePrice:  req.SalePrice,
		CostPrice:  req.CostPrice,
	})
	if err != nil {
		if isUnique(err) {
			return ProductPublic{}, apperror.Conflict("Bu shablon allaqachon do‘konga biriktirilgan")
		}
		return ProductPublic{}, apperror.Internal()
	}
	pub := created.Public()
	pub.Audit = audit.AfterCreate(ctx, productEntity, created.ID)
	return pub, nil
}

func (s *Service) UpdateProduct(ctx context.Context, id uuid.UUID, req UpdateRequest, scope *Scope) (ProductPublic, error) {
	current, err := s.requireProduct(ctx, id, scope)
	if err != nil {
		return ProductPublic{}, err
	}
	if req.SalePrice < 0 || req.CostPrice < 0 {
		return ProductPublic{}, apperror.BadRequest("Narx manfiy bo‘lishi mumkin emas")
	}
	if req.SalePrice < req.CostPrice {
		return ProductPublic{}, apperror.BadRequest("Sotuv narxi asl narxdan kichik bo‘lmasin")
	}
	current.SalePrice = req.SalePrice
	current.CostPrice = req.CostPrice
	updated, err := s.repo.UpdateProduct(ctx, current)
	if err != nil {
		return ProductPublic{}, apperror.Internal()
	}
	pub := updated.Public()
	pub.Audit = audit.AfterUpdate(ctx, productEntity, updated.ID)
	return pub, nil
}

func (s *Service) DeleteProduct(ctx context.Context, id uuid.UUID, scope *Scope) error {
	if _, err := s.requireProduct(ctx, id, scope); err != nil {
		return err
	}
	if err := s.repo.DeleteProduct(ctx, id); err != nil {
		if errors.Is(err, ErrNotFound) {
			return apperror.NotFound("Do‘kon mahsuloti topilmadi")
		}
		return apperror.Internal()
	}
	return nil
}

func (s *Service) ListIncomings(ctx context.Context, q IncomingListQuery, scope *Scope) (IncomingListResult, error) {
	normalizePage(&q.Page, &q.Limit)
	q.Query = strings.TrimSpace(q.Query)
	applyIncomingScope(&q, scope)
	items, total, err := s.repo.ListIncomings(ctx, q)
	if err != nil {
		return IncomingListResult{}, apperror.Internal()
	}
	ids := make([]uuid.UUID, 0, len(items))
	for _, item := range items {
		ids = append(ids, item.ID)
	}
	trails := audit.BindMany(ctx, incomingEntity, ids)
	out := make([]IncomingPublic, 0, len(items))
	for _, item := range items {
		pub := item.Public()
		if t, ok := trails[item.ID]; ok {
			pub.Audit = &t
		}
		out = append(out, pub)
	}
	return IncomingListResult{Items: out, Total: total, Page: q.Page, Limit: q.Limit}, nil
}

func (s *Service) CreateIncoming(ctx context.Context, req IncomingRequest, scope *Scope) (IncomingPublic, error) {
	productID, err := uuid.Parse(strings.TrimSpace(req.ShopProductID))
	if err != nil {
		return IncomingPublic{}, apperror.BadRequest("Mahsulot tanlanishi shart")
	}
	if req.Quantity <= 0 {
		return IncomingPublic{}, apperror.BadRequest("Kirim miqdori 0 dan katta bo‘lishi kerak")
	}
	if _, err = s.requireProduct(ctx, productID, scope); err != nil {
		return IncomingPublic{}, err
	}
	created, err := s.repo.CreateIncoming(ctx, Incoming{
		ID:            uuid.New(),
		ShopProductID: productID,
		Quantity:      req.Quantity,
	})
	if err != nil {
		return IncomingPublic{}, apperror.Internal()
	}
	pub := created.Public()
	pub.Audit = audit.AfterCreate(ctx, incomingEntity, created.ID)
	return pub, nil
}

func (s *Service) requireProduct(ctx context.Context, id uuid.UUID, scope *Scope) (ShopProduct, error) {
	item, err := s.repo.FindProduct(ctx, id)
	if err != nil {
		if errors.Is(err, ErrNotFound) {
			return ShopProduct{}, apperror.NotFound("Do‘kon mahsuloti topilmadi")
		}
		return ShopProduct{}, apperror.Internal()
	}
	if err = assertProductInScope(item, scope); err != nil {
		return ShopProduct{}, err
	}
	return item, nil
}

func applyProductScope(q *ProductListQuery, scope *Scope) {
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

func applyIncomingScope(q *IncomingListQuery, scope *Scope) {
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

func assertProductInScope(item ShopProduct, scope *Scope) error {
	if scope == nil {
		return nil
	}
	if scope.ShopID != nil && item.ShopID != *scope.ShopID {
		return apperror.Forbidden("Faqat o‘z do‘koningizdagi mahsulotni boshqarish mumkin")
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

func assertShopInScope(item ShopRef, scope *Scope) error {
	if scope == nil {
		return nil
	}
	if scope.ShopID != nil && item.ID != *scope.ShopID {
		return apperror.Forbidden("Faqat o‘z do‘koningizga mahsulot biriktirish mumkin")
	}
	if scope.RegionID != nil && item.RegionID != *scope.RegionID {
		return apperror.Forbidden("Do‘kon sizning viloyatingizga tegishli emas")
	}
	if scope.DistrictID != nil && item.DistrictID != *scope.DistrictID {
		return apperror.Forbidden("Do‘kon sizning tumaningizga tegishli emas")
	}
	if scope.MFYID != nil && item.MFYID != *scope.MFYID {
		return apperror.Forbidden("Do‘kon sizning MFYingizga tegishli emas")
	}
	return nil
}

func normalizePage(page, limit *int) {
	if *page < 1 {
		*page = 1
	}
	if *limit < 1 || *limit > 100 {
		*limit = 20
	}
}

func isUnique(err error) bool {
	var pgErr *pgconn.PgError
	return errors.As(err, &pgErr) && pgErr.Code == "23505"
}
