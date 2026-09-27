package localshop

import (
	"time"

	"github.com/google/uuid"

	"qoshida/backend/internal/shared/audit"
)

const (
	StatusActive = "active"
	PurposeSetup = "password_setup"
)

type Shop struct {
	ID           uuid.UUID
	Name         string
	Phone        string
	Image        string
	RegionID     uuid.UUID
	DistrictID   uuid.UUID
	MFYID        uuid.UUID
	PasswordHash string
	Status       string
	CreatedAt    time.Time
	UpdatedAt    time.Time
	RegionName   string
	DistrictName string
	MFYName      string
}

type Public struct {
	ID           string `json:"id"`
	Name         string `json:"name"`
	Phone        string `json:"phone"`
	Image        string `json:"image"`
	RegionID     string `json:"region_id"`
	DistrictID   string `json:"district_id"`
	MFYID        string `json:"mfy_id"`
	Status       string `json:"status"`
	HasPassword  bool   `json:"has_password"`
	RegionName   string `json:"region_name,omitempty"`
	DistrictName string `json:"district_name,omitempty"`
	MFYName      string `json:"mfy_name,omitempty"`
	CreatedAt    string `json:"created_at"`
	Audit        *audit.Trail `json:"audit,omitempty"`
}

func (s Shop) Public() Public {
	return Public{
		ID:           s.ID.String(),
		Name:         s.Name,
		Phone:        s.Phone,
		Image:        s.Image,
		RegionID:     s.RegionID.String(),
		DistrictID:   s.DistrictID.String(),
		MFYID:        s.MFYID.String(),
		Status:       s.Status,
		HasPassword:  s.PasswordHash != "",
		RegionName:   s.RegionName,
		DistrictName: s.DistrictName,
		MFYName:      s.MFYName,
		CreatedAt:    s.CreatedAt.UTC().Format(time.RFC3339),
	}
}

type CreateRequest struct {
	Name        string `json:"name"`
	Phone       string `json:"phone"`
	Image       string `json:"image"`
	ClearImage  bool   `json:"clear_image"`
	RegionID    string `json:"region_id"`
	DistrictID  string `json:"district_id"`
	MFYID       string `json:"mfy_id"`
	Status      string `json:"status"`
	Password    string `json:"password"`
}

type Scope struct {
	RegionID   uuid.UUID
	DistrictID *uuid.UUID
	MFYID      *uuid.UUID
}

type ListQuery struct {
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
	Shop Public `json:"shop"`
}

func ScopeFromManager(regionID uuid.UUID, districtID *uuid.UUID, managerType string) *Scope {
	scope := &Scope{RegionID: regionID}
	if managerType == "district" && districtID != nil {
		scope.DistrictID = districtID
	}
	return scope
}

func ScopeFromDirector(regionID, districtID, mfyID uuid.UUID) *Scope {
	return &Scope{RegionID: regionID, DistrictID: &districtID, MFYID: &mfyID}
}
