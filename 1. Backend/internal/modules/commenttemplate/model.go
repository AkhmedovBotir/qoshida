package commenttemplate

import (
	"time"

	"github.com/google/uuid"
)

const StatusActive = "active"

type CommentTemplate struct {
	ID        uuid.UUID
	Comment   string
	SortOrder int
	Status    string
	CreatedAt time.Time
	UpdatedAt time.Time
}

type Public struct {
	ID        string `json:"id"`
	Comment   string `json:"comment"`
	SortOrder int    `json:"sort_order"`
	Status    string `json:"status"`
	CreatedAt string `json:"created_at"`
	UpdatedAt string `json:"updated_at"`
}

func (t CommentTemplate) Public() Public {
	return Public{
		ID:        t.ID.String(),
		Comment:   t.Comment,
		SortOrder: t.SortOrder,
		Status:    t.Status,
		CreatedAt: t.CreatedAt.UTC().Format(time.RFC3339),
		UpdatedAt: t.UpdatedAt.UTC().Format(time.RFC3339),
	}
}

type CreateRequest struct {
	Comment string `json:"comment"`
	Status  string `json:"status"`
}

type UpdateRequest = CreateRequest

type ReorderRequest struct {
	FromID string `json:"from_id"`
	ToID   string `json:"to_id"`
}

type ListQuery struct {
	Query string
	Page  int
	Limit int
}

type ListResult struct {
	Items      []Public `json:"items"`
	Total      int      `json:"total"`
	Page       int      `json:"page"`
	Limit      int      `json:"limit"`
	TotalPages int      `json:"total_pages"`
}

type Stats struct {
	Total    int `json:"total"`
	Active   int `json:"active"`
	Inactive int `json:"inactive"`
}
