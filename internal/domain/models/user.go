package models

import (
	"time"

	"github.com/google/uuid"
)

type User struct {
	UserID       uuid.UUID `json:"user_id" db:"user_id"`
	TenantID     uuid.UUID `json:"tenant_id" db:"tenant_id"`
	Email        string    `json:"email" db:"email"`
	PasswordHash string    `json:"-" db:"password_hash"` // Hide password hash from JSON
	RoleID       *int      `json:"role_id" db:"role_id"`
	Status       string    `json:"status" db:"status"`
	CreatedAt    time.Time `json:"created_at" db:"created_at"`
	UpdatedAt    time.Time `json:"updated_at" db:"updated_at"`
}
