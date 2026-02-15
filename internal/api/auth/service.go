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
	uow          repositories.UnitOfWork
	tokenService TokenService
}

func NewService(uow repositories.UnitOfWork, tokenService TokenService) Service {
	return &service{
		uow:          uow,
		tokenService: tokenService,
	}
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

	var permissions []string
	if user.RoleID != nil {
		p, err := s.uow.RBAC().GetPermissionsByRoleID(ctx, *user.RoleID)
		if err != nil {
			return nil, fmt.Errorf("failed to fetch user permissions")
		}
		permissions = p
	}

	accessToken, err := s.tokenService.GenerateToken(user.UserID, user.TenantID, user.Email, permissions)
	if err != nil {
		return nil, fmt.Errorf("failed to generate access token")
	}

	return &LoginResponse{
		AccessToken: accessToken,
		Email:       user.Email,
		Status:      "success",
	}, nil
}
