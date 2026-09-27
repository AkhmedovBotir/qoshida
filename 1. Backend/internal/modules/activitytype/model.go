package activitytype

import (
	"time"

	"github.com/google/uuid"
)

const StatusActive = "active"

type ActivityType struct {
	ID        uuid.UUID
	SourceID  string
	Name      string
	Icon      string
	Status    string
	CreatedAt time.Time
}

type Public struct {
	ID        string `json:"id"`
	SourceID  string `json:"source_id"`
	Name      string `json:"name"`
	Icon      string `json:"icon"`
	Status    string `json:"status"`
	CreatedAt string `json:"created_at"`
}

func (a ActivityType) Public() Public {
	return Public{
		ID:        a.ID.String(),
		SourceID:  a.SourceID,
		Name:      a.Name,
		Icon:      a.Icon,
		Status:    a.Status,
		CreatedAt: a.CreatedAt.UTC().Format(time.RFC3339),
	}
}

type CreateRequest struct {
	Name   string `json:"name"`
	Icon   string `json:"icon"`
	Status string `json:"status"`
}

type UpdateRequest = CreateRequest

type ListQuery struct {
	Query  string
	Status string
	Page   int
	Limit  int
}

type ListResult struct {
	Items []Public `json:"items"`
	Total int      `json:"total"`
	Page  int      `json:"page"`
	Limit int      `json:"limit"`
}

type Stats struct {
	Total    int `json:"total"`
	Active   int `json:"active"`
	Inactive int `json:"inactive"`
}

type ImportResult struct {
	Inserted int `json:"inserted"`
	Updated  int `json:"updated"`
	Total    int `json:"total"`
}
