package admin

import (
	"time"

	"github.com/google/uuid"
)

const (
	RoleGeneral = "general"
	RoleAdmin   = "admin"
)

type Admin struct {
	ID           uuid.UUID
	FirstName    string
	LastName     string
	Phone        string
	Username     string
	Role         string
	PasswordHash string
	IsActive     bool
	CreatedAt    time.Time
}

type Public struct {
	ID        string `json:"id"`
	FirstName string `json:"first_name"`
	LastName  string `json:"last_name"`
	Phone     string `json:"phone"`
	Username  string `json:"username"`
	Role      string `json:"role"`
	IsActive  bool   `json:"is_active"`
	CreatedAt string `json:"created_at"`
}

func (a Admin) Public() Public {
	return Public{
		ID:        a.ID.String(),
		FirstName: a.FirstName,
		LastName:  a.LastName,
		Phone:     a.Phone,
		Username:  a.Username,
		Role:      a.Role,
		IsActive:  a.IsActive,
		CreatedAt: a.CreatedAt.UTC().Format(time.RFC3339),
	}
}

type CreateRequest struct {
	FirstName string `json:"first_name"`
	LastName  string `json:"last_name"`
	Phone     string `json:"phone"`
	Username  string `json:"username"`
	Password  string `json:"password"`
}

type UpdateRequest struct {
	FirstName string `json:"first_name"`
	LastName  string `json:"last_name"`
	Phone     string `json:"phone"`
	Username  string `json:"username"`
	Password  string `json:"password"`
}
