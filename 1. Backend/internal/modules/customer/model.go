package customer

import (
	"time"

	"github.com/google/uuid"
)

const (
	StatusActive = "active"

	PurposeRegister = "register"
	PurposeSetup    = "password_setup"
	PurposeReset    = "password_reset"
	PurposeLogin    = "login"
)

type Customer struct {
	ID           uuid.UUID
	Name         string
	FirstName    string
	LastName     string
	BirthDate    *time.Time
	Phone        string
	PasswordHash string
	RegionID     *uuid.UUID
	DistrictID   *uuid.UUID
	MFYID        *uuid.UUID
	Address      string
	Avatar       string
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
	FirstName    string `json:"first_name"`
	LastName     string `json:"last_name"`
	BirthDate    string `json:"birth_date,omitempty"`
	Phone        string `json:"phone"`
	RegionID     string `json:"region_id,omitempty"`
	DistrictID   string `json:"district_id,omitempty"`
	MFYID        string `json:"mfy_id,omitempty"`
	Address      string `json:"address,omitempty"`
	Avatar       string `json:"avatar,omitempty"`
	Status       string `json:"status"`
	RegionName   string `json:"region_name,omitempty"`
	DistrictName string `json:"district_name,omitempty"`
	MFYName      string `json:"mfy_name,omitempty"`
	CreatedAt    string `json:"created_at"`
}

func (c Customer) Public() Public {
	out := Public{
		ID:           c.ID.String(),
		Name:         c.Name,
		FirstName:    c.FirstName,
		LastName:     c.LastName,
		Phone:        c.Phone,
		Address:      c.Address,
		Avatar:       c.Avatar,
		Status:       c.Status,
		RegionName:   c.RegionName,
		DistrictName: c.DistrictName,
		MFYName:      c.MFYName,
		CreatedAt:    c.CreatedAt.UTC().Format(time.RFC3339),
	}
	if c.BirthDate != nil && !c.BirthDate.IsZero() {
		out.BirthDate = c.BirthDate.Format("2006-01-02")
	}
	if c.RegionID != nil {
		out.RegionID = c.RegionID.String()
	}
	if c.DistrictID != nil {
		out.DistrictID = c.DistrictID.String()
	}
	if c.MFYID != nil {
		out.MFYID = c.MFYID.String()
	}
	return out
}

type PhoneRequest struct {
	Phone   string `json:"phone"`
	Purpose string `json:"purpose"`
}

type VerifyRequest struct {
	Phone   string `json:"phone"`
	Code    string `json:"code"`
	Purpose string `json:"purpose"`
}

type RegisterRequest struct {
	Phone      string `json:"phone"`
	FirstName  string `json:"first_name"`
	LastName   string `json:"last_name"`
	BirthDate  string `json:"birth_date"`
	RegionID   string `json:"region_id"`
	DistrictID string `json:"district_id"`
	MFYID      string `json:"mfy_id"`
	Name       string `json:"name"`
	Password   string `json:"password"`
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
	CanRegister bool   `json:"can_register"`
	Message     string `json:"message"`
}

type SendCodeResult struct {
	Status  string `json:"status"`
	Message string `json:"message"`
	Purpose string `json:"purpose"`
	Exists  bool   `json:"exists"`
}

type VerifyResult struct {
	Status   string  `json:"status"`
	Exists   bool    `json:"exists"`
	Purpose  string  `json:"purpose"`
	Customer *Public `json:"customer,omitempty"`
}

type AuthResult struct {
	Customer Public `json:"customer"`
}

type ProfileRequest struct {
	FirstName  string `json:"first_name"`
	LastName   string `json:"last_name"`
	BirthDate  string `json:"birth_date"`
	Name       string `json:"name"`
	RegionID   string `json:"region_id"`
	DistrictID string `json:"district_id"`
	MFYID      string `json:"mfy_id"`
	Address    string `json:"address"`
	Avatar     string `json:"avatar"`
	ClearAvatar bool  `json:"clear_avatar"`
}

type ProfileFields struct {
	FirstName  string
	LastName   string
	Name       string
	BirthDate  *time.Time
	RegionID   *uuid.UUID
	DistrictID *uuid.UUID
	MFYID      *uuid.UUID
	Address    string
}
