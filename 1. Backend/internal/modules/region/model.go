package region

import (
	"time"

	"github.com/google/uuid"
)

const (
	TypeRegion   = "region"
	TypeDistrict = "district"
	TypeMFY      = "mfy"
	StatusActive = "active"
)

type Region struct {
	ID             uuid.UUID
	SourceID       string
	Name           string
	Type           string
	ParentID       *uuid.UUID
	ParentSourceID string
	ParentName     string
	Code           string
	Status         string
	CreatedAt      time.Time
}

type Public struct {
	ID         string  `json:"id"`
	SourceID   string  `json:"source_id"`
	Name       string  `json:"name"`
	Type       string  `json:"type"`
	ParentID   *string `json:"parent_id"`
	ParentName string  `json:"parent_name,omitempty"`
	Code       string  `json:"code"`
	Status     string  `json:"status"`
	CreatedAt  string  `json:"created_at"`
}

func (r Region) Public() Public {
	out := Public{
		ID:         r.ID.String(),
		SourceID:   r.SourceID,
		Name:       r.Name,
		Type:       r.Type,
		ParentName: r.ParentName,
		Code:       r.Code,
		Status:     r.Status,
		CreatedAt:  r.CreatedAt.UTC().Format(time.RFC3339),
	}
	if r.ParentID != nil {
		id := r.ParentID.String()
		out.ParentID = &id
	}
	return out
}

type CreateRequest struct {
	Name     string  `json:"name"`
	Type     string  `json:"type"`
	ParentID *string `json:"parent_id"`
	Code     string  `json:"code"`
	Status   string  `json:"status"`
}

type UpdateRequest = CreateRequest

type ListQuery struct {
	Type     string
	ParentID *uuid.UUID
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
	MFY      int `json:"mfy"`
}

type ImportResult struct {
	Inserted int `json:"inserted"`
	Updated  int `json:"updated"`
	Total    int `json:"total"`
}
