package models

import (
	"time"

	"github.com/google/uuid"
)

type Tenant struct {
	TenantID        uuid.UUID              `json:"tenant_id" db:"tenant_id"`
	Name            string                 `json:"name" db:"name"`
	IndustrySegment string                 `json:"industry_segment" db:"industry_segment"`
	APIKeyHash      *string                `json:"-" db:"api_key_hash"`
	Settings        map[string]interface{} `json:"settings" db:"settings"`
	Status          string                 `json:"status" db:"status"`
	CreatedAt       time.Time              `json:"created_at" db:"created_at"`
	UpdatedAt       time.Time              `json:"updated_at" db:"updated_at"`
	DeletedAt       *time.Time             `json:"deleted_at,omitempty" db:"deleted_at"`
}
