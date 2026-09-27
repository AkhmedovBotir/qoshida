package seller

import (
	"time"

	"github.com/google/uuid"

	"qoshida/backend/internal/shared/audit"
)

const (
	StatusActive = "active"
	PurposeSetup = "password_setup"
)

type Seller struct {
	ID           uuid.UUID
	ShopID       uuid.UUID
	FirstName    string
	LastName     string
	Phone        string
	PasswordHash string
	Status       string
	CreatedAt    time.Time
	UpdatedAt    time.Time
	ShopName     string
	RegionID     uuid.UUID
	DistrictID   uuid.UUID
	MFYID        uuid.UUID
	RegionName   string
	DistrictName string
	MFYName      string
}

type ShopRef struct {
	ID         uuid.UUID
	Name       string
	RegionID   uuid.UUID
	DistrictID uuid.UUID
	MFYID      uuid.UUID
	Status     string
}

type Public struct {
	ID           string `json:"id"`
	ShopID       string `json:"shop_id"`
	FirstName    string `json:"first_name"`
	LastName     string `json:"last_name"`
	Phone        string `json:"phone"`
	Status       string `json:"status"`
	HasPassword  bool   `json:"has_password"`
	ShopName     string `json:"shop_name,omitempty"`
	RegionID     string `json:"region_id,omitempty"`
	DistrictID   string `json:"district_id,omitempty"`
	MFYID        string `json:"mfy_id,omitempty"`
	RegionName   string `json:"region_name,omitempty"`
	DistrictName string `json:"district_name,omitempty"`
	MFYName      string `json:"mfy_name,omitempty"`
	CreatedAt    string `json:"created_at"`
	Audit        *audit.Trail `json:"audit,omitempty"`
}

func (s Seller) Public() Public {
	return Public{
		ID:           s.ID.String(),
		ShopID:       s.ShopID.String(),
		FirstName:    s.FirstName,
		LastName:     s.LastName,
		Phone:        s.Phone,
		Status:       s.Status,
		HasPassword:  s.PasswordHash != "",
		ShopName:     s.ShopName,
		RegionID:     s.RegionID.String(),
		DistrictID:   s.DistrictID.String(),
		MFYID:        s.MFYID.String(),
		RegionName:   s.RegionName,
		DistrictName: s.DistrictName,
		MFYName:      s.MFYName,
		CreatedAt:    s.CreatedAt.UTC().Format(time.RFC3339),
	}
}

type CreateRequest struct {
	ShopID     string `json:"shop_id"`
	RegionID   string `json:"region_id"`
	DistrictID string `json:"district_id"`
	MFYID      string `json:"mfy_id"`
	FirstName  string `json:"first_name"`
	LastName   string `json:"last_name"`
	Phone      string `json:"phone"`
	Status     string `json:"status"`
	Password   string `json:"password"`
}

type Scope struct {
	ShopID     *uuid.UUID
	RegionID   *uuid.UUID
	DistrictID *uuid.UUID
	MFYID      *uuid.UUID
}

type ListQuery struct {
	ShopID     *uuid.UUID
	RegionID   *uuid.UUID
	DistrictID *uuid.UUID
	MFYID      *uuid.UUID
	Query      string
	Page       int
	Limit      int
}

type ListResult struct {
	Items []Public `json:"items"`
	Total int      `json:"total"`
	Page  int      `json:"page"`
	Limit int      `json:"limit"`
}

type Stats struct {
	Total  int `json:"total"`
	Active int `json:"active"`
}

type PhoneRequest struct {
	Phone string `json:"phone"`
}

type VerifyRequest struct {
	Phone string `json:"phone"`
	Code  string `json:"code"`
}

type SetPasswordRequest struct {
	Phone    string `json:"phone"`
	Password string `json:"password"`
}

type LoginRequest struct {
	Phone    string `json:"phone"`
	Password string `json:"password"`
}

type PhoneCheck struct {
	Exists      bool   `json:"exists"`
	Active      bool   `json:"active"`
	HasPassword bool   `json:"has_password"`
	NeedsSetup  bool   `json:"needs_setup"`
	Message     string `json:"message"`
}

type AuthResult struct {
	Seller Public `json:"seller"`
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
