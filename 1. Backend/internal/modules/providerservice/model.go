package providerservice

import (
	"time"

	"github.com/google/uuid"

	"qoshida/backend/internal/shared/audit"
)

const (
	ApprovalPending  = "pending"
	ApprovalApproved = "approved"
	ApprovalRejected = "rejected"

	SubmittedProvider = "provider"
	SubmittedStaff    = "staff"
)

type ServiceItem struct {
	ID             uuid.UUID
	ProviderID     uuid.UUID
	Name           string
	Price          float64
	Images         []string
	ApprovalStatus string
	RejectionNote  string
	SubmittedBy    string
	ReviewedByName string
	ReviewedByRole string
	ReviewedAt     *time.Time
	CreatedAt      time.Time
	UpdatedAt      time.Time
	ProviderName   string
	RegionID       uuid.UUID
	DistrictID     uuid.UUID
	MFYID          uuid.UUID
}

type ProviderRef struct {
	ID                   uuid.UUID
	Name                 string
	RegionID             uuid.UUID
	DistrictID           uuid.UUID
	MFYID                uuid.UUID
	Status               string
	IdentificationStatus string
}

type Public struct {
	ID             string       `json:"id"`
	ProviderID     string       `json:"provider_id"`
	Name           string       `json:"name"`
	Price          float64      `json:"price"`
	Images         []string     `json:"images"`
	Image          string       `json:"image,omitempty"`
	ApprovalStatus string       `json:"approval_status"`
	RejectionNote  string       `json:"rejection_note,omitempty"`
	SubmittedBy    string       `json:"submitted_by"`
	ReviewedByName string       `json:"reviewed_by_name,omitempty"`
	ReviewedByRole string       `json:"reviewed_by_role,omitempty"`
	ReviewedAt     string       `json:"reviewed_at,omitempty"`
	ProviderName   string       `json:"provider_name,omitempty"`
	CreatedAt      string       `json:"created_at"`
	Audit          *audit.Trail `json:"audit,omitempty"`
}

func (s ServiceItem) Public() Public {
	images := s.Images
	if images == nil {
		images = []string{}
	}
	image := ""
	if len(images) > 0 {
		image = images[0]
	}
	out := Public{
		ID:             s.ID.String(),
		ProviderID:     s.ProviderID.String(),
		Name:           s.Name,
		Price:          s.Price,
		Images:         images,
		Image:          image,
		ApprovalStatus: s.ApprovalStatus,
		RejectionNote:  s.RejectionNote,
		SubmittedBy:    s.SubmittedBy,
		ReviewedByName: s.ReviewedByName,
		ReviewedByRole: s.ReviewedByRole,
		ProviderName:   s.ProviderName,
		CreatedAt:      s.CreatedAt.UTC().Format(time.RFC3339),
	}
	if s.ReviewedAt != nil {
		out.ReviewedAt = s.ReviewedAt.UTC().Format(time.RFC3339)
	}
	return out
}

type ItemRequest struct {
	Name   string   `json:"name"`
	Price  float64  `json:"price"`
	Images []string `json:"images"`
}

type CreateRequest struct {
	ProviderID string        `json:"provider_id"`
	Items      []ItemRequest `json:"items"`
	Name       string        `json:"name"`
	Price      float64       `json:"price"`
	Images     []string      `json:"images"`
}

type UpdateRequest struct {
	Name   string   `json:"name"`
	Price  float64  `json:"price"`
	Images []string `json:"images"`
}

type RejectRequest struct {
	Note string `json:"note"`
}

type Scope struct {
	ProviderID *uuid.UUID
	RegionID   *uuid.UUID
	DistrictID *uuid.UUID
	MFYID      *uuid.UUID
	Staff      bool
}

type ListQuery struct {
	ProviderID     *uuid.UUID
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

type CreateResult struct {
	Items []Public `json:"items"`
}

func ScopeFromProvider(id uuid.UUID) *Scope {
	return &Scope{ProviderID: &id, Staff: false}
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
