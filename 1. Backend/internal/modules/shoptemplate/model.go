package shoptemplate

import (
	"time"

	"github.com/google/uuid"

	"qoshida/backend/internal/shared/audit"
)

const (
	UnitDona = "dona"
	UnitLitr = "litr"
	UnitKg   = "kg"
)

type Template struct {
	ID              uuid.UUID
	CategoryID      uuid.UUID
	SubcategoryID   uuid.UUID
	Name            string
	Description     string
	Unit            string
	UnitSize        float64
	Images          []string
	CreatedAt       time.Time
	UpdatedAt       time.Time
	CategoryName    string
	SubcategoryName string
}

type CategoryRef struct {
	ID       uuid.UUID
	ParentID *uuid.UUID
	Name     string
	Status   string
}

type Public struct {
	ID              string       `json:"id"`
	CategoryID      string       `json:"category_id"`
	SubcategoryID   string       `json:"subcategory_id"`
	Name            string       `json:"name"`
	Description     string       `json:"description"`
	Unit            string       `json:"unit"`
	UnitSize        float64      `json:"unit_size"`
	Images          []string     `json:"images"`
	CategoryName    string       `json:"category_name,omitempty"`
	SubcategoryName string       `json:"subcategory_name,omitempty"`
	CreatedAt       string       `json:"created_at"`
	Audit           *audit.Trail `json:"audit,omitempty"`
}

func (t Template) Public() Public {
	images := t.Images
	if images == nil {
		images = []string{}
	}
	return Public{
		ID:              t.ID.String(),
		CategoryID:      t.CategoryID.String(),
		SubcategoryID:   t.SubcategoryID.String(),
		Name:            t.Name,
		Description:     t.Description,
		Unit:            t.Unit,
		UnitSize:        t.UnitSize,
		Images:          images,
		CategoryName:    t.CategoryName,
		SubcategoryName: t.SubcategoryName,
		CreatedAt:       t.CreatedAt.UTC().Format(time.RFC3339),
	}
}

type CreateRequest struct {
	CategoryID    string   `json:"category_id"`
	SubcategoryID string   `json:"subcategory_id"`
	Name          string   `json:"name"`
	Description   string   `json:"description"`
	Unit          string   `json:"unit"`
	UnitSize      float64  `json:"unit_size"`
	Images        []string `json:"images"`
}

type ListQuery struct {
	CategoryID *uuid.UUID
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
