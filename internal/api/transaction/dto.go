package transaction

import (
	"encoding/json"
	"time"

	validation "github.com/go-ozzo/ozzo-validation/v4"
	"github.com/google/uuid"
)

// CreateTransactionRequest defines the JSON input for transaction ingestion
type CreateTransactionRequest struct {
	CorrelationID *uuid.UUID      `json:"correlation_id,omitempty"` // Optional idempotency key
	UserID        string          `json:"user_id"`
	Amount        float64         `json:"amount"`
	Currency      string          `json:"currency"`
	DeviceID      string          `json:"device_id"`
	IPAddress     string          `json:"ip_address"`
	Payload       json.RawMessage `json:"payload"` // Arbitrary JSON for merchant-specific data
}

func (r CreateTransactionRequest) Validate() error {
	return validation.ValidateStruct(&r,
		validation.Field(&r.UserID, validation.Required, validation.Length(1, 0)),
		validation.Field(&r.Amount, validation.Required, validation.Min(0.01)),
		validation.Field(&r.Currency, validation.Required, validation.Length(3, 3)),
	)
}

// CreateTransactionResponse defines the success output
type CreateTransactionResponse struct {
	TransactionID uuid.UUID `json:"transaction_id"`
	Status        string    `json:"status"`
	ReceivedAt    time.Time `json:"received_at"`
}
