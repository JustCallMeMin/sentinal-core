package repositories

import (
	"context"

	"github.com/google/uuid"
	"github.com/sentinal/core/internal/domain/models"
)

type UserRepository interface {
	Create(ctx context.Context, user *models.User) error
	FindByEmail(ctx context.Context, tenantID uuid.UUID, email string) (*models.User, error)
	GetByID(ctx context.Context, userID uuid.UUID) (*models.User, error)
	UpdateStatus(ctx context.Context, userID uuid.UUID, status string) error
}
