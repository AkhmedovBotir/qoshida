package market

import (
	"context"
	"errors"
	"strings"
	"unicode/utf8"

	"github.com/google/uuid"

	"qoshida/backend/internal/shared/apperror"
)

type Service struct {
	repo *Repository
}

func NewService(repo *Repository) *Service {
	return &Service{repo: repo}
}

func (s *Service) ListCatalog(ctx context.Context, q ListQuery) (ListResult, error) {
	if q.Page < 1 {
		q.Page = 1
	}
	if q.Limit < 1 || q.Limit > 60 {
		q.Limit = 20
	}
	q.Query = strings.TrimSpace(q.Query)
	q.Kind = strings.TrimSpace(q.Kind)
	items, total, err := s.repo.ListCatalog(ctx, q)
	if err != nil {
		return ListResult{}, apperror.Internal()
	}
	out := make([]CatalogPublic, 0, len(items))
	for _, item := range items {
		out = append(out, item.Public())
	}
	return ListResult{Items: out, Total: total, Page: q.Page, Limit: q.Limit}, nil
}

func (s *Service) GetCatalog(ctx context.Context, kind string, id uuid.UUID) (CatalogPublic, error) {
	kind = strings.TrimSpace(kind)
	if kind != KindProduct && kind != KindShop && kind != KindService {
		return CatalogPublic{}, apperror.BadRequest("Tur noto‘g‘ri")
	}
	item, err := s.repo.GetCatalog(ctx, kind, id)
	if err != nil {
		if errors.Is(err, ErrNotFound) {
			return CatalogPublic{}, apperror.NotFound("Mahsulot topilmadi")
		}
		return CatalogPublic{}, apperror.Internal()
	}
	return item.Public(), nil
}

func (s *Service) Categories(ctx context.Context, parent *uuid.UUID, roots bool) ([]CategoryPublic, error) {
	items, err := s.repo.ListCategories(ctx, parent, roots, 80)
	if err != nil {
		return nil, apperror.Internal()
	}
	return items, nil
}

func (s *Service) Cart(ctx context.Context, customerID uuid.UUID) (CartPublic, error) {
	rows, err := s.repo.ListCart(ctx, customerID)
	if err != nil {
		return CartPublic{}, apperror.Internal()
	}
	return s.hydrateCart(ctx, rows)
}

func (s *Service) UpsertCart(ctx context.Context, customerID uuid.UUID, req CartUpsertRequest) (CartPublic, error) {
	kind := strings.TrimSpace(req.Kind)
	if kind != KindProduct && kind != KindShop && kind != KindService {
		return CartPublic{}, apperror.BadRequest("Tur noto‘g‘ri")
	}
	id, err := uuid.Parse(strings.TrimSpace(req.ItemID))
	if err != nil {
		return CartPublic{}, apperror.BadRequest("Mahsulot tanlanishi shart")
	}
	if req.Quantity < 0 {
		return CartPublic{}, apperror.BadRequest("Miqdor noto‘g‘ri")
	}
	if req.Quantity > 0 {
		item, err := s.repo.GetCatalog(ctx, kind, id)
		if err != nil {
			if errors.Is(err, ErrNotFound) {
				return CartPublic{}, apperror.NotFound("Mahsulot topilmadi")
			}
			return CartPublic{}, apperror.Internal()
		}
		if kind != KindService && req.Quantity > item.Quantity {
			return CartPublic{}, apperror.BadRequest("Omborda yetarli mahsulot yo‘q")
		}
	}
	if err = s.repo.UpsertCart(ctx, customerID, kind, id, req.Quantity); err != nil {
		return CartPublic{}, apperror.Internal()
	}
	return s.Cart(ctx, customerID)
}

func (s *Service) Checkout(ctx context.Context, customerID uuid.UUID, req CheckoutRequest) (OrderPublic, error) {
	address := strings.TrimSpace(req.Address)
	if n := utf8.RuneCountInString(address); n < 8 || n > 300 {
		return OrderPublic{}, apperror.BadRequest("Yetkazish manzili 8-300 belgi oralig‘ida bo‘lishi kerak")
	}
	note := strings.TrimSpace(req.Note)
	if utf8.RuneCountInString(note) > 400 {
		return OrderPublic{}, apperror.BadRequest("Izoh 400 belgidan oshmasin")
	}
	rows, err := s.repo.ListCart(ctx, customerID)
	if err != nil {
		return OrderPublic{}, apperror.Internal()
	}
	if len(rows) == 0 {
		return OrderPublic{}, apperror.BadRequest("Savat bo‘sh")
	}
	order := Order{
		ID:         uuid.New(),
		CustomerID: customerID,
		Status:     OrderPending,
		Address:    address,
		Note:       note,
		RegionID:   parseUUID(req.RegionID),
		DistrictID: parseUUID(req.DistrictID),
		MFYID:      parseUUID(req.MFYID),
	}
	var total float64
	for _, row := range rows {
		item, err := s.repo.GetCatalog(ctx, row.Kind, row.ItemID)
		if err != nil {
			return OrderPublic{}, apperror.BadRequest("Savatchada mavjud bo‘lmagan mahsulot bor")
		}
		if row.Kind != KindService && row.Quantity > item.Quantity {
			return OrderPublic{}, apperror.BadRequest(item.Name + " omborda yetarli emas")
		}
		image := ""
		if len(item.Images) > 0 {
			image = item.Images[0]
		}
		order.Items = append(order.Items, OrderItem{
			Kind:       row.Kind,
			ItemID:     item.ID,
			Name:       item.Name,
			Unit:       item.Unit,
			Quantity:   row.Quantity,
			UnitPrice:  item.Price,
			SellerName: item.SellerName,
			Image:      image,
		})
		total += item.Price * row.Quantity
	}
	order.Total = roundMoney(total)
	created, err := s.repo.CreateOrder(ctx, order)
	if err != nil {
		if errors.Is(err, errInsufficient) {
			return OrderPublic{}, apperror.BadRequest("Omborda yetarli mahsulot yo‘q")
		}
		return OrderPublic{}, apperror.Internal()
	}
	_ = s.repo.ClearCart(ctx, customerID)
	return created.Public(), nil
}

func (s *Service) Orders(ctx context.Context, customerID uuid.UUID, page, limit int) (OrderListResult, error) {
	if page < 1 {
		page = 1
	}
	if limit < 1 || limit > 50 {
		limit = 20
	}
	items, total, err := s.repo.ListOrders(ctx, customerID, page, limit)
	if err != nil {
		return OrderListResult{}, apperror.Internal()
	}
	out := make([]OrderPublic, 0, len(items))
	for _, item := range items {
		out = append(out, item.Public())
	}
	return OrderListResult{Items: out, Total: total, Page: page, Limit: limit}, nil
}

func (s *Service) Order(ctx context.Context, customerID, id uuid.UUID) (OrderPublic, error) {
	item, err := s.repo.FindOrder(ctx, id, customerID)
	if err != nil {
		if errors.Is(err, ErrNotFound) {
			return OrderPublic{}, apperror.NotFound("Buyurtma topilmadi")
		}
		return OrderPublic{}, apperror.Internal()
	}
	return item.Public(), nil
}

func (s *Service) hydrateCart(ctx context.Context, rows []CartRow) (CartPublic, error) {
	out := CartPublic{Items: make([]CartLine, 0, len(rows))}
	for _, row := range rows {
		line := CartLine{
			ID:       row.ID.String(),
			Kind:     row.Kind,
			ItemID:   row.ItemID.String(),
			Quantity: row.Quantity,
		}
		item, err := s.repo.GetCatalog(ctx, row.Kind, row.ItemID)
		if err == nil {
			pub := item.Public()
			line.Item = &pub
			line.LineTotal = roundMoney(pub.Price * row.Quantity)
			out.Total += line.LineTotal
		}
		out.Items = append(out.Items, line)
	}
	out.Total = roundMoney(out.Total)
	out.Count = len(out.Items)
	return out, nil
}

func parseUUID(raw string) uuid.UUID {
	id, err := uuid.Parse(strings.TrimSpace(raw))
	if err != nil {
		return uuid.Nil
	}
	return id
}
