package kontragent

import (
	"time"

	"github.com/google/uuid"

	"qoshida/backend/internal/shared/audit"
)

const (
	StatusActive = "active"
	PurposeSetup = "password_setup"
)

type Kontragent struct {
	ID             uuid.UUID
	ActivityTypeID uuid.UUID
	Name           string
	INN            string
	RegionID       uuid.UUID
	DistrictID     uuid.UUID
	MFYID          uuid.UUID
	Phone          string
	Logo           string
	PasswordHash   string
	Status         string
	CreatedAt      time.Time
	UpdatedAt      time.Time
	ActivityName   string
	RegionName     string
	DistrictName   string
	MFYName        string
}

type Public struct {
	ID             string `json:"id"`
	ActivityTypeID string `json:"activity_type_id"`
	Name           string `json:"name"`
	INN            string `json:"inn"`
	RegionID       string `json:"region_id"`
	DistrictID     string `json:"district_id"`
	MFYID          string `json:"mfy_id"`
	Phone          string `json:"phone"`
	Logo           string `json:"logo"`
	Status         string `json:"status"`
	HasPassword    bool   `json:"has_password"`
	ActivityName   string `json:"activity_name,omitempty"`
	RegionName     string `json:"region_name,omitempty"`
	DistrictName   string `json:"district_name,omitempty"`
	MFYName        string `json:"mfy_name,omitempty"`
	CreatedAt      string `json:"created_at"`
	Audit          *audit.Trail `json:"audit,omitempty"`
}

func (k Kontragent) Public() Public {
	return Public{
		ID:             k.ID.String(),
		ActivityTypeID: k.ActivityTypeID.String(),
		Name:           k.Name,
		INN:            k.INN,
		RegionID:       k.RegionID.String(),
		DistrictID:     k.DistrictID.String(),
		MFYID:          k.MFYID.String(),
		Phone:          k.Phone,
		Logo:           k.Logo,
		Status:         k.Status,
		HasPassword:    k.PasswordHash != "",
		ActivityName:   k.ActivityName,
		RegionName:     k.RegionName,
		DistrictName:   k.DistrictName,
		MFYName:        k.MFYName,
		CreatedAt:      k.CreatedAt.UTC().Format(time.RFC3339),
	}
}

type CreateRequest struct {
	ActivityTypeID string `json:"activity_type_id"`
	Name           string `json:"name"`
	INN            string `json:"inn"`
	RegionID       string `json:"region_id"`
	DistrictID     string `json:"district_id"`
	MFYID          string `json:"mfy_id"`
	Phone          string `json:"phone"`
	Logo           string `json:"logo"`
	ClearLogo      bool   `json:"clear_logo"`
	Status         string `json:"status"`
	Password       string `json:"password"`
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
	Kontragent Public `json:"kontragent"`
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
