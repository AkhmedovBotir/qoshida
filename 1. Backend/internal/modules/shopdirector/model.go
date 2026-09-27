package shopdirector

import (
	"time"

	"github.com/google/uuid"

	"qoshida/backend/internal/shared/audit"
)

const (
	StatusActive = "active"
	PurposeSetup = "password_setup"
)

type Director struct {
	ID           uuid.UUID
	RegionID     uuid.UUID
	DistrictID   uuid.UUID
	MFYID        uuid.UUID
	FirstName    string
	LastName     string
	Phone        string
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
	RegionID     string `json:"region_id"`
	DistrictID   string `json:"district_id"`
	MFYID        string `json:"mfy_id"`
	FirstName    string `json:"first_name"`
	LastName     string `json:"last_name"`
	Phone        string `json:"phone"`
	Status       string `json:"status"`
	HasPassword  bool   `json:"has_password"`
	RegionName   string `json:"region_name,omitempty"`
	DistrictName string `json:"district_name,omitempty"`
	MFYName      string `json:"mfy_name,omitempty"`
	CreatedAt    string `json:"created_at"`
	Audit        *audit.Trail `json:"audit,omitempty"`
}

func (d Director) Public() Public {
	return Public{
		ID:           d.ID.String(),
		RegionID:     d.RegionID.String(),
		DistrictID:   d.DistrictID.String(),
		MFYID:        d.MFYID.String(),
		FirstName:    d.FirstName,
		LastName:     d.LastName,
		Phone:        d.Phone,
		Status:       d.Status,
		HasPassword:  d.PasswordHash != "",
		RegionName:   d.RegionName,
		DistrictName: d.DistrictName,
		MFYName:      d.MFYName,
		CreatedAt:    d.CreatedAt.UTC().Format(time.RFC3339),
	}
}

type CreateRequest struct {
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
	RegionID   uuid.UUID
	DistrictID *uuid.UUID
}

type ListQuery struct {
	RegionID   *uuid.UUID
	DistrictID *uuid.UUID
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
	Director Public `json:"director"`
}
