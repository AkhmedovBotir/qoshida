package category

import (
	"time"

	"github.com/google/uuid"

	"qoshida/backend/internal/shared/audit"
)

const StatusActive = "active"

type Category struct {
	ID             uuid.UUID
	SourceID       string
	Name           string
	Slug           string
	ParentID       *uuid.UUID
	ParentSourceID string
	ParentName     string
	ImageURL       string
	Censored       bool
	Status         string
	CreatedAt      time.Time
}

type Public struct {
	ID         string  `json:"id"`
	SourceID   string  `json:"source_id"`
	Name       string  `json:"name"`
	Slug       string  `json:"slug"`
	ParentID   *string `json:"parent_id"`
	ParentName string  `json:"parent_name,omitempty"`
	ImageURL   string  `json:"image_url,omitempty"`
	Censored   bool    `json:"censored"`
	Status     string       `json:"status"`
	CreatedAt  string       `json:"created_at"`
	Audit      *audit.Trail `json:"audit,omitempty"`
}

func (c Category) Public() Public {
	out := Public{
		ID:         c.ID.String(),
		SourceID:   c.SourceID,
		Name:       c.Name,
		Slug:       c.Slug,
		ParentName: c.ParentName,
		ImageURL:   c.ImageURL,
		Censored:   c.Censored,
		Status:     c.Status,
		CreatedAt:  c.CreatedAt.UTC().Format(time.RFC3339),
	}
	if c.ParentID != nil {
		id := c.ParentID.String()
		out.ParentID = &id
	}
	return out
}

type CreateRequest struct {
	Name     string  `json:"name"`
	Slug     string  `json:"slug"`
	ParentID *string `json:"parent_id"`
	Censored   bool   `json:"censored"`
	Status     string `json:"status"`
	Image      string `json:"image"`
	ClearImage bool   `json:"clear_image"`
}

type UpdateRequest = CreateRequest

type ListQuery struct {
	ParentID *uuid.UUID
	Roots    bool
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
	Total int `json:"total"`
	Root  int `json:"root"`
	Child int `json:"child"`
}

type ImportResult struct {
	Inserted int `json:"inserted"`
	Updated  int `json:"updated"`
	Total    int `json:"total"`
	Images   int `json:"images"`
}
