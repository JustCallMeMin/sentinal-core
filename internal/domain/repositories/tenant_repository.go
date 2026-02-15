package repositories

import (
	"context"

	"github.com/google/uuid"
	"github.com/sentinal/core/internal/domain/models"
)

type TenantRepository interface {
	Create(ctx context.Context, tenant *models.Tenant) error
	GetByID(ctx context.Context, id uuid.UUID) (*models.Tenant, error)
	GetByAPIKeyHash(ctx context.Context, hash string) (*models.Tenant, error)
	List(ctx context.Context, limit, offset int) ([]*models.Tenant, error)
	Update(ctx context.Context, tenant *models.Tenant) error
	Delete(ctx context.Context, id uuid.UUID) error
}
