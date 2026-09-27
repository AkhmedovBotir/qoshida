package shopstock

import (
	"time"

	"github.com/google/uuid"

	"qoshida/backend/internal/shared/audit"
)

type ShopProduct struct {
	ID              uuid.UUID
	ShopID          uuid.UUID
	TemplateID      uuid.UUID
	Quantity        float64
	SalePrice       float64
	CostPrice       float64
	CreatedAt       time.Time
	UpdatedAt       time.Time
	Name            string
	Description     string
	CategoryID      uuid.UUID
	SubcategoryID   uuid.UUID
	CategoryName    string
	SubcategoryName string
	Unit            string
	UnitSize        float64
	Images          []string
	ShopName        string
	RegionID        uuid.UUID
	DistrictID      uuid.UUID
	MFYID           uuid.UUID
	RegionName      string
	DistrictName    string
	MFYName         string
}

type Incoming struct {
	ID            uuid.UUID
	ShopProductID uuid.UUID
	Quantity      float64
	CreatedAt     time.Time
	ProductName   string
	ShopID        uuid.UUID
	ShopName      string
	Unit          string
	UnitSize      float64
	RegionID      uuid.UUID
	DistrictID    uuid.UUID
	MFYID         uuid.UUID
}

type ShopRef struct {
	ID         uuid.UUID
	Name       string
	RegionID   uuid.UUID
	DistrictID uuid.UUID
	MFYID      uuid.UUID
	Status     string
}

type TemplateRef struct {
	ID   uuid.UUID
	Name string
}

type ProductPublic struct {
	ID              string       `json:"id"`
	ShopID          string       `json:"shop_id"`
	TemplateID      string       `json:"template_id"`
	Quantity        float64      `json:"quantity"`
	SalePrice       float64      `json:"sale_price"`
	CostPrice       float64      `json:"cost_price"`
	Name            string       `json:"name"`
	Description     string       `json:"description"`
	CategoryID      string       `json:"category_id"`
	SubcategoryID   string       `json:"subcategory_id"`
	CategoryName    string       `json:"category_name,omitempty"`
	SubcategoryName string       `json:"subcategory_name,omitempty"`
	Unit            string       `json:"unit"`
	UnitSize        float64      `json:"unit_size"`
	Images          []string     `json:"images"`
	ShopName        string       `json:"shop_name,omitempty"`
	RegionName      string       `json:"region_name,omitempty"`
	DistrictName    string       `json:"district_name,omitempty"`
	MFYName         string       `json:"mfy_name,omitempty"`
	CreatedAt       string       `json:"created_at"`
	Audit           *audit.Trail `json:"audit,omitempty"`
}

func (p ShopProduct) Public() ProductPublic {
	images := p.Images
	if images == nil {
		images = []string{}
	}
	return ProductPublic{
		ID:              p.ID.String(),
		ShopID:          p.ShopID.String(),
		TemplateID:      p.TemplateID.String(),
		Quantity:        p.Quantity,
		SalePrice:       p.SalePrice,
		CostPrice:       p.CostPrice,
		Name:            p.Name,
		Description:     p.Description,
		CategoryID:      p.CategoryID.String(),
		SubcategoryID:   p.SubcategoryID.String(),
		CategoryName:    p.CategoryName,
		SubcategoryName: p.SubcategoryName,
		Unit:            p.Unit,
		UnitSize:        p.UnitSize,
		Images:          images,
		ShopName:        p.ShopName,
		RegionName:      p.RegionName,
		DistrictName:    p.DistrictName,
		MFYName:         p.MFYName,
		CreatedAt:       p.CreatedAt.UTC().Format(time.RFC3339),
	}
}

type IncomingPublic struct {
	ID            string       `json:"id"`
	ShopProductID string       `json:"shop_product_id"`
	Quantity      float64      `json:"quantity"`
	ProductName   string       `json:"product_name,omitempty"`
	ShopID        string       `json:"shop_id,omitempty"`
	ShopName      string       `json:"shop_name,omitempty"`
	Unit          string       `json:"unit,omitempty"`
	UnitSize      float64      `json:"unit_size,omitempty"`
	CreatedAt     string       `json:"created_at"`
	Audit         *audit.Trail `json:"audit,omitempty"`
}

func (i Incoming) Public() IncomingPublic {
	return IncomingPublic{
		ID:            i.ID.String(),
		ShopProductID: i.ShopProductID.String(),
		Quantity:      i.Quantity,
		ProductName:   i.ProductName,
		ShopID:        i.ShopID.String(),
		ShopName:      i.ShopName,
		Unit:          i.Unit,
		UnitSize:      i.UnitSize,
		CreatedAt:     i.CreatedAt.UTC().Format(time.RFC3339),
	}
}

type AttachRequest struct {
	ShopID     string  `json:"shop_id"`
	TemplateID string  `json:"template_id"`
	Quantity   float64 `json:"quantity"`
	SalePrice  float64 `json:"sale_price"`
	CostPrice  float64 `json:"cost_price"`
}

type UpdateRequest struct {
	SalePrice float64 `json:"sale_price"`
	CostPrice float64 `json:"cost_price"`
}

type IncomingRequest struct {
	ShopProductID string  `json:"shop_product_id"`
	Quantity      float64 `json:"quantity"`
}

type Scope struct {
	ShopID     *uuid.UUID
	RegionID   *uuid.UUID
	DistrictID *uuid.UUID
	MFYID      *uuid.UUID
}

type ProductListQuery struct {
	ShopID     *uuid.UUID
	TemplateID *uuid.UUID
	Query      string
	Page       int
	Limit      int
	RegionID   *uuid.UUID
	DistrictID *uuid.UUID
	MFYID      *uuid.UUID
}

type IncomingListQuery struct {
	ShopID        *uuid.UUID
	ShopProductID *uuid.UUID
	Query         string
	Page          int
	Limit         int
	RegionID      *uuid.UUID
	DistrictID    *uuid.UUID
	MFYID         *uuid.UUID
}

type ProductListResult struct {
	Items []ProductPublic `json:"items"`
	Total int             `json:"total"`
	Page  int             `json:"page"`
	Limit int             `json:"limit"`
}

type IncomingListResult struct {
	Items []IncomingPublic `json:"items"`
	Total int              `json:"total"`
	Page  int              `json:"page"`
	Limit int              `json:"limit"`
}

func ScopeFromShop(shopID uuid.UUID) *Scope {
	return &Scope{ShopID: &shopID}
}

func ScopeFromDirector(regionID, districtID, mfyID uuid.UUID) *Scope {
	return &Scope{RegionID: &regionID, DistrictID: &districtID, MFYID: &mfyID}
}

func ScopeFromManager(regionID uuid.UUID, districtID *uuid.UUID, managerType string) *Scope {
	scope := &Scope{RegionID: &regionID}
	if managerType == "district" && districtID != nil {
		scope.DistrictID = districtID
	}
	return scope
}
