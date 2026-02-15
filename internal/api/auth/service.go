package auth

import (
	"context"
	"fmt"

	"github.com/google/uuid"
	"github.com/sentinal/core/internal/domain/repositories"
	"github.com/sentinal/core/pkg/security"
)

type Service interface {
	Login(ctx context.Context, req *LoginRequest) (*LoginResponse, error)
}

type service struct {
	uow repositories.UnitOfWork
}

func NewService(uow repositories.UnitOfWork) Service {
	return &service{uow: uow}
}

func (s *service) Login(ctx context.Context, req *LoginRequest) (*LoginResponse, error) {
	tenantID, err := uuid.Parse(req.TenantID)
	if err != nil {
		return nil, fmt.Errorf("invalid tenant id")
	}

	user, err := s.uow.Users().FindByEmail(ctx, tenantID, req.Email)
	if err != nil {
		return nil, fmt.Errorf("invalid credentials") // Don't leak if user exists
	}

	if !security.CheckPasswordHash(req.Password, user.PasswordHash) {
		return nil, fmt.Errorf("invalid credentials")
	}

	if user.Status != "active" {
		return nil, fmt.Errorf("account is %s", user.Status)
	}

	// For SC-022, we just return a success message or placeholder token
	// JWT implementation is slated for SC-023
	return &LoginResponse{
		AccessToken: "placeholder_token_for_now",
		Email:       user.Email,
		Status:      "success",
	}, nil
}
