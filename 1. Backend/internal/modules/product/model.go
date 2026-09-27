package product

import (
	"time"

	"github.com/google/uuid"

	"qoshida/backend/internal/shared/audit"
)

const (
	StatusActive = "active"

	ApprovalPending  = "pending"
	ApprovalApproved = "approved"
	ApprovalRejected = "rejected"

	SubmittedKontragent = "kontragent"
	SubmittedStaff      = "staff"

	UnitDona = "dona"
	UnitLitr = "litr"
	UnitKg   = "kg"
)

type Product struct {
	ID                uuid.UUID
	KontragentID      uuid.UUID
	CategoryID        uuid.UUID
	SubcategoryID     uuid.UUID
	Name              string
	Description       string
	SalePrice         float64
	CostPrice         float64
	Quantity          float64
	Unit              string
	UnitSize          float64
	CommissionPercent float64
	Images            []string
	ApprovalStatus    string
	RejectionNote     string
	SubmittedBy       string
	Status            string
	ReviewedByName    string
	ReviewedByRole    string
	ReviewedAt        *time.Time
	CreatedAt         time.Time
	UpdatedAt         time.Time
	KontragentName    string
	CategoryName      string
	SubcategoryName   string
	RegionID          uuid.UUID
	DistrictID        uuid.UUID
	MFYID             uuid.UUID
	RegionName        string
	DistrictName      string
	MFYName           string
}

type KontragentRef struct {
	ID         uuid.UUID
	Name       string
	RegionID   uuid.UUID
	DistrictID uuid.UUID
	MFYID      uuid.UUID
	Status     string
}

type CategoryRef struct {
	ID       uuid.UUID
	ParentID *uuid.UUID
	Name     string
	Status   string
}

type Public struct {
	ID                string       `json:"id"`
	KontragentID      string       `json:"kontragent_id"`
	CategoryID        string       `json:"category_id"`
	SubcategoryID     string       `json:"subcategory_id"`
	Name              string       `json:"name"`
	Description       string       `json:"description"`
	SalePrice         float64      `json:"sale_price"`
	CostPrice         float64      `json:"cost_price"`
	Quantity          float64      `json:"quantity"`
	Unit              string       `json:"unit"`
	UnitSize          float64      `json:"unit_size"`
	CommissionPercent float64      `json:"commission_percent"`
	CommissionAmount  float64      `json:"commission_amount"`
	Images            []string     `json:"images"`
	ApprovalStatus    string       `json:"approval_status"`
	RejectionNote     string       `json:"rejection_note,omitempty"`
	SubmittedBy       string       `json:"submitted_by"`
	Status            string       `json:"status"`
	ReviewedByName    string       `json:"reviewed_by_name,omitempty"`
	ReviewedByRole    string       `json:"reviewed_by_role,omitempty"`
	ReviewedAt        string       `json:"reviewed_at,omitempty"`
	KontragentName    string       `json:"kontragent_name,omitempty"`
	CategoryName      string       `json:"category_name,omitempty"`
	SubcategoryName   string       `json:"subcategory_name,omitempty"`
	RegionName        string       `json:"region_name,omitempty"`
	DistrictName      string       `json:"district_name,omitempty"`
	MFYName           string       `json:"mfy_name,omitempty"`
	CreatedAt         string       `json:"created_at"`
	Audit             *audit.Trail `json:"audit,omitempty"`
}

func (p Product) Public() Public {
	images := p.Images
	if images == nil {
		images = []string{}
	}
	margin := p.SalePrice - p.CostPrice
	if margin < 0 {
		margin = 0
	}
	out := Public{
		ID:                p.ID.String(),
		KontragentID:      p.KontragentID.String(),
		CategoryID:        p.CategoryID.String(),
		SubcategoryID:     p.SubcategoryID.String(),
		Name:              p.Name,
		Description:       p.Description,
		SalePrice:         p.SalePrice,
		CostPrice:         p.CostPrice,
		Quantity:          p.Quantity,
		Unit:              p.Unit,
		UnitSize:          p.UnitSize,
		CommissionPercent: p.CommissionPercent,
		CommissionAmount:  roundMoney(margin * p.CommissionPercent / 100),
		Images:            images,
		ApprovalStatus:    p.ApprovalStatus,
		RejectionNote:     p.RejectionNote,
		SubmittedBy:       p.SubmittedBy,
		Status:            p.Status,
		ReviewedByName:    p.ReviewedByName,
		ReviewedByRole:    p.ReviewedByRole,
		KontragentName:    p.KontragentName,
		CategoryName:      p.CategoryName,
		SubcategoryName:   p.SubcategoryName,
		RegionName:        p.RegionName,
		DistrictName:      p.DistrictName,
		MFYName:           p.MFYName,
		CreatedAt:         p.CreatedAt.UTC().Format(time.RFC3339),
	}
	if p.ReviewedAt != nil {
		out.ReviewedAt = p.ReviewedAt.UTC().Format(time.RFC3339)
	}
	return out
}

func roundMoney(v float64) float64 {
	return float64(int(v*100+0.5)) / 100
}

type CreateRequest struct {
	KontragentID      string   `json:"kontragent_id"`
	CategoryID        string   `json:"category_id"`
	SubcategoryID     string   `json:"subcategory_id"`
	Name              string   `json:"name"`
	Description       string   `json:"description"`
	SalePrice         float64  `json:"sale_price"`
	CostPrice         float64  `json:"cost_price"`
	Quantity          float64  `json:"quantity"`
	Unit              string   `json:"unit"`
	UnitSize          float64  `json:"unit_size"`
	CommissionPercent float64  `json:"commission_percent"`
	Images            []string `json:"images"`
	Status            string   `json:"status"`
}

type RejectRequest struct {
	Note string `json:"note"`
}

type Scope struct {
	KontragentID *uuid.UUID
	RegionID     *uuid.UUID
	DistrictID   *uuid.UUID
	MFYID        *uuid.UUID
	Staff        bool
}

type ListQuery struct {
	KontragentID   *uuid.UUID
	CategoryID     *uuid.UUID
	ApprovalStatus string
	Query          string
	Page           int
	Limit          int
	RegionID       *uuid.UUID
	DistrictID     *uuid.UUID
	MFYID          *uuid.UUID
}

type ListResult struct {
	Items []Public `json:"items"`
	Total int      `json:"total"`
	Page  int      `json:"page"`
	Limit int      `json:"limit"`
}

func ScopeFromKontragent(id uuid.UUID) *Scope {
	return &Scope{KontragentID: &id, Staff: false}
}

func ScopeFromManager(regionID uuid.UUID, districtID *uuid.UUID, managerType string) *Scope {
	scope := &Scope{RegionID: &regionID, Staff: true}
	if managerType == "district" && districtID != nil {
		scope.DistrictID = districtID
	}
	return scope
}

func ScopeFromDirector(regionID, districtID, mfyID uuid.UUID) *Scope {
	return &Scope{RegionID: &regionID, DistrictID: &districtID, MFYID: &mfyID, Staff: true}
}

func isStaff(scope *Scope) bool {
	return scope == nil || scope.Staff
}
