package repository

import (
	"context"
	"fmt"

	"github.com/georgysavva/scany/v2/pgxscan"
	"github.com/google/uuid"
	"github.com/sentinal/core/internal/domain/models"
)

type userRepository struct {
	*BaseRepository[models.User]
}

func NewUserRepository(db Querier) *userRepository {
	return &userRepository{
		BaseRepository: NewBaseRepository[models.User](db, "users", ""),
	}
}

func (r *userRepository) Create(ctx context.Context, user *models.User) error {
	if user.UserID == uuid.Nil {
		user.UserID = uuid.New()
	}
	query := `
		INSERT INTO users (user_id, tenant_id, email, password_hash, role_id, status, created_at, updated_at)
		VALUES ($1, $2, $3, $4, $5, $6, NOW(), NOW())
	`
	_, err := r.DB.Exec(ctx, query,
		user.UserID,
		user.TenantID,
		user.Email,
		user.PasswordHash,
		user.RoleID,
		user.Status,
	)
	if err != nil {
		return fmt.Errorf("failed to create user: %w", err)
	}
	return nil
}

func (r *userRepository) FindByEmail(ctx context.Context, tenantID uuid.UUID, email string) (*models.User, error) {
	var user models.User
	query := `SELECT * FROM users WHERE tenant_id = $1 AND email = $2 LIMIT 1`
	err := pgxscan.Get(ctx, r.DB, &user, query, tenantID, email)
	if err != nil {
		if pgxscan.NotFound(err) {
			return nil, fmt.Errorf("user not found")
		}
		return nil, fmt.Errorf("failed to find user by email: %w", err)
	}
	return &user, nil
}

func (r *userRepository) GetByID(ctx context.Context, userID uuid.UUID) (*models.User, error) {
	var user models.User
	query := `SELECT * FROM users WHERE user_id = $1 LIMIT 1`
	err := pgxscan.Get(ctx, r.DB, &user, query, userID)
	if err != nil {
		if pgxscan.NotFound(err) {
			return nil, fmt.Errorf("user not found")
		}
		return nil, fmt.Errorf("failed to get user by id: %w", err)
	}
	return &user, nil
}

func (r *userRepository) UpdateStatus(ctx context.Context, userID uuid.UUID, status string) error {
	query := `UPDATE users SET status = $1, updated_at = NOW() WHERE user_id = $2`
	_, err := r.DB.Exec(ctx, query, status, userID)
	if err != nil {
		return fmt.Errorf("failed to update user status: %w", err)
	}
	return nil
}

func (r *userRepository) IncrementFailedAttempts(ctx context.Context, userID uuid.UUID, maxFailedAttempts int, lockoutDurationMinutes int) error {
	query := `
		UPDATE users 
		SET failed_attempts = failed_attempts + 1,
		    locked_until = CASE 
		        WHEN failed_attempts + 1 >= $1 THEN NOW() + ($2 * INTERVAL '1 minute')
		        ELSE locked_until 
		    END,
		    updated_at = NOW()
		WHERE user_id = $3
	`
	_, err := r.DB.Exec(ctx, query, maxFailedAttempts, lockoutDurationMinutes, userID)
	if err != nil {
		return fmt.Errorf("failed to increment failed attempts: %w", err)
	}
	return nil
}

func (r *userRepository) ResetFailedAttempts(ctx context.Context, userID uuid.UUID) error {
	query := `UPDATE users SET failed_attempts = 0, locked_until = NULL, updated_at = NOW() WHERE user_id = $1`
	_, err := r.DB.Exec(ctx, query, userID)
	if err != nil {
		return fmt.Errorf("failed to reset failed attempts: %w", err)
	}
	return nil
}

func (r *userRepository) UpdateMFASecret(ctx context.Context, userID uuid.UUID, secret string) error {
	query := `UPDATE users SET mfa_secret = $1, updated_at = NOW() WHERE user_id = $2`
	_, err := r.DB.Exec(ctx, query, secret, userID)
	if err != nil {
		return fmt.Errorf("failed to update mfa secret: %w", err)
	}
	return nil
}

func (r *userRepository) SetMFAEnabled(ctx context.Context, userID uuid.UUID, enabled bool) error {
	query := `UPDATE users SET mfa_enabled = $1, updated_at = NOW() WHERE user_id = $2`
	_, err := r.DB.Exec(ctx, query, enabled, userID)
	if err != nil {
		return fmt.Errorf("failed to set mfa enabled: %w", err)
	}
	return nil
}
