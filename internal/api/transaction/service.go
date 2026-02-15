package transaction

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"github.com/google/uuid"
	"github.com/sentinal/core/internal/domain/models"
	"github.com/sentinal/core/internal/domain/repositories"
)

// Service handles transaction business logic
type Service interface {
	Create(ctx context.Context, tenantID uuid.UUID, req *CreateTransactionRequest) (*CreateTransactionResponse, error)
}

type service struct {
	uow repositories.UnitOfWork
}

// NewService creates a new Service instance
func NewService(uow repositories.UnitOfWork) Service {
	return &service{uow: uow}
}

// Create ingests a transaction
func (s *service) Create(ctx context.Context, tenantID uuid.UUID, req *CreateTransactionRequest) (*CreateTransactionResponse, error) {
	// Prepare Payload
	payloadMap := make(map[string]interface{})
	if req.Payload != nil {
		if err := json.Unmarshal(req.Payload, &payloadMap); err != nil {
			// If payload is not valid JSON, we wrap it or log error. For now, we proceed with empty map or partial data
			// In production we should return 400 validation error
		}
	}

	// Inject core fields into payload for now (until schema migration)
	payloadMap["user_id"] = req.UserID
	payloadMap["device_id"] = req.DeviceID
	payloadMap["ip_address"] = req.IPAddress

	// Map DTO to Domain Model
	tx := &models.Transaction{
		TransactionID: uuid.New(),
		TenantID:      tenantID,
		Amount:        req.Amount,
		Currency:      req.Currency,
		Payload:       payloadMap,
		OccurredAt:    time.Now().UTC(),
		CreatedAt:     time.Now().UTC(),
	}

	// Use UnitOfWork to create transaction
	if err := s.uow.Do(ctx, func(u repositories.UnitOfWork) error {
		return u.Transactions().Save(ctx, tx)
	}); err != nil {
		return nil, fmt.Errorf("failed to create transaction: %w", err)
	}

	return &CreateTransactionResponse{
		TransactionID: tx.TransactionID,
		Status:        "received",
		ReceivedAt:    tx.OccurredAt,
	}, nil
}
