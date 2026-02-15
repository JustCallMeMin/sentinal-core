package models

import (
	"time"

	"github.com/google/uuid"
)

type Transaction struct {
	TransactionID uuid.UUID              `json:"transaction_id" db:"transaction_id"`
	TenantID      uuid.UUID              `json:"tenant_id" db:"tenant_id"`
	CorrelationID string                 `json:"correlation_id" db:"correlation_id"`
	Amount        float64                `json:"amount" db:"amount"`
	Currency      string                 `json:"currency" db:"currency"`
	OccurredAt    time.Time              `json:"occurred_at" db:"occurred_at"`
	Payload       map[string]interface{} `json:"payload" db:"payload"`
	CreatedAt     time.Time              `json:"created_at" db:"created_at"`
}
