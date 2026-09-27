package market

import (
	"time"

	"github.com/google/uuid"
)

const (
	KindProduct = "product"
	KindShop    = "shop"
	KindService = "service"

	OrderPending    = "pending"
	OrderConfirmed  = "confirmed"
	OrderDelivering = "delivering"
	OrderDone       = "done"
	OrderCancelled  = "cancelled"
)

type CatalogItem struct {
	Kind         string
	ID           uuid.UUID
	Name         string
	Description  string
	Price        float64
	Quantity     float64
	Unit         string
	UnitSize     float64
	Images       []string
	SellerName   string
	SellerID     uuid.UUID
	CategoryID   uuid.UUID
	CategoryName string
	RegionID     uuid.UUID
	DistrictID   uuid.UUID
	MFYID        uuid.UUID
	RegionName   string
	DistrictName string
	MFYName      string
}

type CatalogPublic struct {
	Kind         string   `json:"kind"`
	ID           string   `json:"id"`
	Name         string   `json:"name"`
	Description  string   `json:"description,omitempty"`
	Price        float64  `json:"price"`
	Quantity     float64  `json:"quantity"`
	Unit         string   `json:"unit"`
	UnitSize     float64  `json:"unit_size,omitempty"`
	Images       []string `json:"images"`
	Image        string   `json:"image,omitempty"`
	SellerName   string   `json:"seller_name,omitempty"`
	SellerID     string   `json:"seller_id,omitempty"`
	CategoryID   string   `json:"category_id,omitempty"`
	CategoryName string   `json:"category_name,omitempty"`
	RegionName   string   `json:"region_name,omitempty"`
	DistrictName string   `json:"district_name,omitempty"`
	MFYName      string   `json:"mfy_name,omitempty"`
}

func (item CatalogItem) Public() CatalogPublic {
	images := item.Images
	if images == nil {
		images = []string{}
	}
	image := ""
	if len(images) > 0 {
		image = images[0]
	}
	out := CatalogPublic{
		Kind:         item.Kind,
		ID:           item.ID.String(),
		Name:         item.Name,
		Description:  item.Description,
		Price:        item.Price,
		Quantity:     item.Quantity,
		Unit:         item.Unit,
		UnitSize:     item.UnitSize,
		Images:       images,
		Image:        image,
		SellerName:   item.SellerName,
		CategoryName: item.CategoryName,
		RegionName:   item.RegionName,
		DistrictName: item.DistrictName,
		MFYName:      item.MFYName,
	}
	if item.SellerID != uuid.Nil {
		out.SellerID = item.SellerID.String()
	}
	if item.CategoryID != uuid.Nil {
		out.CategoryID = item.CategoryID.String()
	}
	return out
}

type CategoryPublic struct {
	ID       string `json:"id"`
	Name     string `json:"name"`
	Slug     string `json:"slug,omitempty"`
	ImageURL string `json:"image_url,omitempty"`
	ParentID string `json:"parent_id,omitempty"`
}

type ListQuery struct {
	Kind       string
	Query      string
	CategoryID *uuid.UUID
	RegionID   *uuid.UUID
	DistrictID *uuid.UUID
	MFYID      *uuid.UUID
	Page       int
	Limit      int
}

type ListResult struct {
	Items []CatalogPublic `json:"items"`
	Total int             `json:"total"`
	Page  int             `json:"page"`
	Limit int             `json:"limit"`
}

type CartRow struct {
	ID        uuid.UUID
	Kind      string
	ItemID    uuid.UUID
	Quantity  float64
	CreatedAt time.Time
}

type CartLine struct {
	ID        string         `json:"id"`
	Kind      string         `json:"kind"`
	ItemID    string         `json:"item_id"`
	Quantity  float64        `json:"quantity"`
	Item      *CatalogPublic `json:"item,omitempty"`
	LineTotal float64        `json:"line_total"`
}

type CartPublic struct {
	Items []CartLine `json:"items"`
	Total float64    `json:"total"`
	Count int        `json:"count"`
}

type CartUpsertRequest struct {
	Kind     string  `json:"kind"`
	ItemID   string  `json:"item_id"`
	Quantity float64 `json:"quantity"`
}

type CheckoutRequest struct {
	Address    string `json:"address"`
	Note       string `json:"note"`
	RegionID   string `json:"region_id"`
	DistrictID string `json:"district_id"`
	MFYID      string `json:"mfy_id"`
}

type Order struct {
	ID         uuid.UUID
	CustomerID uuid.UUID
	Status     string
	Total      float64
	Address    string
	Note       string
	RegionID   uuid.UUID
	DistrictID uuid.UUID
	MFYID      uuid.UUID
	CreatedAt  time.Time
	Items      []OrderItem
}

type OrderItem struct {
	Kind       string
	ItemID     uuid.UUID
	Name       string
	Unit       string
	Quantity   float64
	UnitPrice  float64
	SellerName string
	Image      string
}

type OrderPublic struct {
	ID        string            `json:"id"`
	Status    string            `json:"status"`
	Total     float64           `json:"total"`
	Address   string            `json:"address"`
	Note      string            `json:"note,omitempty"`
	CreatedAt string            `json:"created_at"`
	Items     []OrderItemPublic `json:"items"`
}

type OrderItemPublic struct {
	Kind       string  `json:"kind"`
	ItemID     string  `json:"item_id"`
	Name       string  `json:"name"`
	Unit       string  `json:"unit"`
	Quantity   float64 `json:"quantity"`
	UnitPrice  float64 `json:"unit_price"`
	SellerName string  `json:"seller_name,omitempty"`
	Image      string  `json:"image,omitempty"`
	LineTotal  float64 `json:"line_total"`
}

func (o Order) Public() OrderPublic {
	items := make([]OrderItemPublic, 0, len(o.Items))
	for _, row := range o.Items {
		items = append(items, OrderItemPublic{
			Kind:       row.Kind,
			ItemID:     row.ItemID.String(),
			Name:       row.Name,
			Unit:       row.Unit,
			Quantity:   row.Quantity,
			UnitPrice:  row.UnitPrice,
			SellerName: row.SellerName,
			Image:      row.Image,
			LineTotal:  roundMoney(row.UnitPrice * row.Quantity),
		})
	}
	return OrderPublic{
		ID:        o.ID.String(),
		Status:    o.Status,
		Total:     o.Total,
		Address:   o.Address,
		Note:      o.Note,
		CreatedAt: o.CreatedAt.UTC().Format(time.RFC3339),
		Items:     items,
	}
}

type OrderListResult struct {
	Items []OrderPublic `json:"items"`
	Total int           `json:"total"`
	Page  int           `json:"page"`
	Limit int           `json:"limit"`
}

func roundMoney(v float64) float64 {
	return float64(int(v*100+0.5)) / 100
}
