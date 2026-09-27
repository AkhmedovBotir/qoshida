package manager

import (
	"time"

	"github.com/google/uuid"

	"qoshida/backend/internal/shared/audit"
)

const (
	TypeRegion   = "region"
	TypeDistrict = "district"
	StatusActive = "active"
	PurposeSetup = "password_setup"
)

type Manager struct {
	ID           uuid.UUID
	Type         string
	RegionID     uuid.UUID
	DistrictID   *uuid.UUID
	FirstName    string
	LastName     string
	Phone        string
	Username     string
	PasswordHash string
	Status       string
	CreatedAt    time.Time
	UpdatedAt    time.Time
	RegionName   string
	DistrictName string
}

type Public struct {
	ID           string  `json:"id"`
	Type         string  `json:"type"`
	RegionID     string  `json:"region_id"`
	DistrictID   *string `json:"district_id"`
	FirstName    string  `json:"first_name"`
	LastName     string  `json:"last_name"`
	Phone        string  `json:"phone"`
	Username     string  `json:"username"`
	Status       string  `json:"status"`
	HasPassword  bool    `json:"has_password"`
	RegionName   string       `json:"region_name,omitempty"`
	DistrictName string       `json:"district_name,omitempty"`
	CreatedAt    string       `json:"created_at"`
	Audit        *audit.Trail `json:"audit,omitempty"`
}

func (m Manager) Public() Public {
	out := Public{
		ID:           m.ID.String(),
		Type:         m.Type,
		RegionID:     m.RegionID.String(),
		FirstName:    m.FirstName,
		LastName:     m.LastName,
		Phone:        m.Phone,
		Username:     m.Username,
		Status:       m.Status,
		HasPassword:  m.PasswordHash != "",
		RegionName:   m.RegionName,
		DistrictName: m.DistrictName,
		CreatedAt:    m.CreatedAt.UTC().Format(time.RFC3339),
	}
	if m.DistrictID != nil {
		id := m.DistrictID.String()
		out.DistrictID = &id
	}
	return out
}

type CreateRequest struct {
	Type       string `json:"type"`
	RegionID   string `json:"region_id"`
	DistrictID string `json:"district_id"`
	FirstName  string `json:"first_name"`
	LastName   string `json:"last_name"`
	Phone      string `json:"phone"`
	Username   string `json:"username"`
	Status     string `json:"status"`
	Password   string `json:"password"`
}

type UpdateRequest = CreateRequest

type ListQuery struct {
	Type     string
	RegionID *uuid.UUID
	Query    string
	Page     int
	Limit    int
}

type ListResult struct {
	Items []Public `json:"items"`
	Total int      `json:"total"`
	Page  int      `json:"page"`
	Limit int      `json:"limit"`
}

type Stats struct {
	Total    int `json:"total"`
	Region   int `json:"region"`
	District int `json:"district"`
	Active   int `json:"active"`
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
	Manager Public `json:"manager"`
}
