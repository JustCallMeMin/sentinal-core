package repositories

import (
	"context"

	"github.com/google/uuid"
	"github.com/sentinal/core/internal/domain/models"
)

type TransactionRepository interface {
	Save(ctx context.Context, tx *models.Transaction) error
	FindByCorrelation(ctx context.Context, tenantID, correlationID uuid.UUID) (*models.Transaction, error)
	SaveBulk(ctx context.Context, transactions []*models.Transaction) (int64, error)
	ListByTenant(ctx context.Context, tenantID string, limit, offset int) ([]*models.Transaction, error)
}
