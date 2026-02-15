package auth

import (
	"context"
	"fmt"
	"time"

	"github.com/google/uuid"
	"github.com/pquerna/otp/totp"
	"github.com/sentinal/core/internal/domain/repositories"
	"github.com/sentinal/core/pkg/config"
	"github.com/sentinal/core/pkg/security"
)

type Service interface {
	Login(ctx context.Context, req *LoginRequest) (*LoginResponse, error)
	GenerateMFASetup(ctx context.Context, userID uuid.UUID) (*MFASetupResponse, error)
	ActivateMFA(ctx context.Context, userID uuid.UUID, code string) error
	VerifyMFA(ctx context.Context, req *MFAVerifyRequest) (*LoginResponse, error)
}

type service struct {
	uow          repositories.UnitOfWork
	tokenService TokenService
	cfg          *config.Config
}

func NewService(uow repositories.UnitOfWork, tokenService TokenService, cfg *config.Config) Service {
	return &service{
		uow:          uow,
		tokenService: tokenService,
		cfg:          cfg,
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

	if user.LockedUntil != nil && time.Now().Before(*user.LockedUntil) {
		return nil, fmt.Errorf("account is locked until %v", user.LockedUntil.Format(time.RFC3339))
	}

	if !security.CheckPasswordHash(req.Password, user.PasswordHash) {
		_ = s.uow.Users().IncrementFailedAttempts(ctx, user.UserID, s.cfg.AuthMaxFailedAttempts, s.cfg.AuthLockoutMinutes)
		return nil, fmt.Errorf("invalid credentials")
	}

	if user.Status != "active" {
		return nil, fmt.Errorf("account is %s", user.Status)
	}

	// Reset counter on success
	if user.FailedAttempts > 0 || user.LockedUntil != nil {
		_ = s.uow.Users().ResetFailedAttempts(ctx, user.UserID)
	}

	var permissions []string
	if user.RoleID != nil {
		p, err := s.uow.RBAC().GetPermissionsByRoleID(ctx, *user.RoleID)
		if err != nil {
			return nil, fmt.Errorf("failed to fetch user permissions")
		}
		permissions = p
	}

	if user.MFAEnabled {
		mfaToken, err := s.tokenService.GenerateMFAToken(user.UserID, user.TenantID)
		if err != nil {
			return nil, fmt.Errorf("failed to generate mfa token")
		}
		return &LoginResponse{
			MFAToken: mfaToken,
			Email:    user.Email,
			Status:   "mfa_required",
		}, nil
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

func (s *service) GenerateMFASetup(ctx context.Context, userID uuid.UUID) (*MFASetupResponse, error) {
	user, err := s.uow.Users().GetByID(ctx, userID)
	if err != nil {
		return nil, fmt.Errorf("user not found")
	}

	key, err := totp.Generate(totp.GenerateOpts{
		Issuer:      "Sentinal",
		AccountName: user.Email,
	})
	if err != nil {
		return nil, fmt.Errorf("failed to generate MFA secret")
	}

	// Store secret (initially disabled)
	if err := s.uow.Users().UpdateMFASecret(ctx, userID, key.Secret()); err != nil {
		return nil, err
	}

	return &MFASetupResponse{
		Secret: key.Secret(),
		QRCode: key.String(),
	}, nil
}

func (s *service) ActivateMFA(ctx context.Context, userID uuid.UUID, code string) error {
	user, err := s.uow.Users().GetByID(ctx, userID)
	if err != nil {
		return fmt.Errorf("user not found")
	}

	if user.MFASecret == nil {
		return fmt.Errorf("MFA not set up")
	}

	valid := totp.Validate(code, *user.MFASecret)
	if !valid {
		return fmt.Errorf("invalid MFA code")
	}

	return s.uow.Users().SetMFAEnabled(ctx, userID, true)
}

func (s *service) VerifyMFA(ctx context.Context, req *MFAVerifyRequest) (*LoginResponse, error) {
	claims, err := s.tokenService.ValidateToken(req.MFAToken)
	if err != nil || !claims.IsMFA {
		return nil, fmt.Errorf("invalid or expired MFA token")
	}

	userID, _ := uuid.Parse(claims.Subject)
	user, err := s.uow.Users().GetByID(ctx, userID)
	if err != nil {
		return nil, fmt.Errorf("user not found")
	}

	if !user.MFAEnabled || user.MFASecret == nil {
		return nil, fmt.Errorf("MFA not enabled")
	}

	valid := totp.Validate(req.Code, *user.MFASecret)
	if !valid {
		return nil, fmt.Errorf("invalid MFA code")
	}

	// Fetch permissions
	var permissions []string
	if user.RoleID != nil {
		p, err := s.uow.RBAC().GetPermissionsByRoleID(ctx, *user.RoleID)
		if err != nil {
			return nil, fmt.Errorf("failed to fetch user permissions")
		}
		permissions = p
	}

	// Generate full access token
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
