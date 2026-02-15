package api_test

import (
	"bytes"
	"context"
	"encoding/json"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/pquerna/otp/totp"
	"github.com/sentinal/core/internal/api/auth"
	"github.com/sentinal/core/internal/domain/models"
	"github.com/sentinal/core/internal/domain/repositories"
	"github.com/sentinal/core/internal/repository"
	"github.com/sentinal/core/internal/server"
	"github.com/sentinal/core/internal/testutil"
	"github.com/sentinal/core/pkg/config"
	"github.com/sentinal/core/pkg/logger"
	"github.com/sentinal/core/pkg/security"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestMFAFlow_Integration(t *testing.T) {
	ctx := context.Background()
	logger.Init("error", "test")

	pool, cleanup := testutil.SetupTestDB(t)
	defer cleanup()

	uow := repository.NewUnitOfWork(pool)

	// Setup Tenant
	tenant := &models.Tenant{
		Name:            "MFA Test Merchant",
		IndustrySegment: "Test",
	}
	err := uow.Do(ctx, func(u repositories.UnitOfWork) error {
		return u.Tenants().Create(ctx, tenant)
	})
	require.NoError(t, err)

	password := "secure123"
	hash, _ := security.HashPassword(password)
	user := &models.User{
		TenantID:     tenant.TenantID,
		Email:        "mfa@example.com",
		PasswordHash: hash,
		Status:       "active",
	}
	err = uow.Do(ctx, func(u repositories.UnitOfWork) error {
		return u.Users().Create(ctx, user)
	})
	require.NoError(t, err)

	// Initialize Server
	cfg := &config.Config{
		AppEnv:                 "test",
		JWTSecret:              "test-secret",
		JWTExpiry:              1,
		AuthMaxFailedAttempts:  5,
		AuthLockoutMinutes:     15,
		RateLimitMax:           1000,
		RateLimitWindowMinutes: 1,
	}
	tokenService := auth.NewTokenService(cfg.JWTSecret, cfg.JWTExpiry)
	authService := auth.NewService(uow, tokenService, cfg)
	authHandler := auth.NewHandler(authService)
	srv := server.New(cfg, pool, nil, uow, nil, authHandler, tokenService)

	var accessToken string

	t.Run("Step 1: Get Access Token (Pre-MFA login)", func(t *testing.T) {
		loginReq := auth.LoginRequest{
			TenantID: tenant.TenantID.String(),
			Email:    user.Email,
			Password: password,
		}
		body, _ := json.Marshal(loginReq)
		req := httptest.NewRequest("POST", "/api/v1/auth/login", bytes.NewReader(body))
		req.Header.Set("Content-Type", "application/json")
		resp, _ := srv.App.Test(req)
		assert.Equal(t, 200, resp.StatusCode)

		var res auth.LoginResponse
		json.NewDecoder(resp.Body).Decode(&res)
		accessToken = res.AccessToken
		assert.Equal(t, "success", res.Status)
	})

	var mfaSecret string

	t.Run("Step 2: Setup MFA", func(t *testing.T) {
		req := httptest.NewRequest("POST", "/api/v1/auth/mfa/setup", nil)
		req.Header.Set("Authorization", "Bearer "+accessToken)
		resp, _ := srv.App.Test(req)
		assert.Equal(t, 200, resp.StatusCode)

		var res auth.MFASetupResponse
		json.NewDecoder(resp.Body).Decode(&res)
		assert.NotEmpty(t, res.Secret)
		mfaSecret = res.Secret
	})

	t.Run("Step 3: Activate MFA", func(t *testing.T) {
		code, _ := totp.GenerateCode(mfaSecret, time.Now())

		activateReq := auth.MFAActivateRequest{Code: code}
		body, _ := json.Marshal(activateReq)
		req := httptest.NewRequest("POST", "/api/v1/auth/mfa/activate", bytes.NewReader(body))
		req.Header.Set("Authorization", "Bearer "+accessToken)
		req.Header.Set("Content-Type", "application/json")
		resp, _ := srv.App.Test(req)
		assert.Equal(t, 200, resp.StatusCode)

		u, _ := uow.Users().GetByID(ctx, user.UserID)
		assert.True(t, u.MFAEnabled)
	})

	var mfaToken string

	t.Run("Step 4: Login with MFA Enabled", func(t *testing.T) {
		loginReq := auth.LoginRequest{
			TenantID: tenant.TenantID.String(),
			Email:    user.Email,
			Password: password,
		}
		body, _ := json.Marshal(loginReq)
		req := httptest.NewRequest("POST", "/api/v1/auth/login", bytes.NewReader(body))
		req.Header.Set("Content-Type", "application/json")
		resp, _ := srv.App.Test(req)
		assert.Equal(t, 200, resp.StatusCode)

		var res auth.LoginResponse
		json.NewDecoder(resp.Body).Decode(&res)
		assert.Equal(t, "mfa_required", res.Status)
		assert.NotEmpty(t, res.MFAToken)
		assert.Empty(t, res.AccessToken)
		mfaToken = res.MFAToken
	})

	t.Run("Step 5: Verify MFA and get Access Token", func(t *testing.T) {
		code, _ := totp.GenerateCode(mfaSecret, time.Now())
		verifyReq := auth.MFAVerifyRequest{
			MFAToken: mfaToken,
			Code:     code,
		}
		body, _ := json.Marshal(verifyReq)
		req := httptest.NewRequest("POST", "/api/v1/auth/mfa/verify", bytes.NewReader(body))
		req.Header.Set("Content-Type", "application/json")
		resp, _ := srv.App.Test(req)
		assert.Equal(t, 200, resp.StatusCode)

		var res auth.LoginResponse
		json.NewDecoder(resp.Body).Decode(&res)
		assert.Equal(t, "success", res.Status)
		assert.NotEmpty(t, res.AccessToken)
	})
}
