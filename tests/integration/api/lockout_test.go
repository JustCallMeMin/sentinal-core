package api_test

import (
	"bytes"
	"context"
	"encoding/json"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/google/uuid"
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

func TestAccountLockout_Integration(t *testing.T) {
	ctx := context.Background()
	logger.Init("error", "test")

	pool, cleanup := testutil.SetupTestDB(t)
	defer cleanup()

	uow := repository.NewUnitOfWork(pool)

	// Setup Tenant
	tenant := &models.Tenant{
		Name:            "Lockout Test Merchant",
		IndustrySegment: "Test",
	}
	err := uow.Do(ctx, func(u repositories.UnitOfWork) error {
		return u.Tenants().Create(ctx, tenant)
	})
	require.NoError(t, err)

	password := "correct_password"
	hash, _ := security.HashPassword(password)
	email := "lockout@example.com"
	userID := uuid.New()

	user := &models.User{
		UserID:       userID,
		TenantID:     tenant.TenantID,
		Email:        email,
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

	loginReq := auth.LoginRequest{
		TenantID: tenant.TenantID.String(),
		Email:    email,
		Password: "wrong_password",
	}

	t.Run("Lockout after 5 failures", func(t *testing.T) {
		// 1-4 failures: 401 Unauthorized
		for i := 0; i < 4; i++ {
			body, _ := json.Marshal(loginReq)
			req := httptest.NewRequest("POST", "/api/v1/auth/login", bytes.NewReader(body))
			req.Header.Set("Content-Type", "application/json")
			resp, _ := srv.App.Test(req)
			assert.Equal(t, 401, resp.StatusCode, "Failure %d should be 401", i+1)
		}

		// 5th failure: Still 401 (increments to 5 and sets lock)
		body, _ := json.Marshal(loginReq)
		req5 := httptest.NewRequest("POST", "/api/v1/auth/login", bytes.NewReader(body))
		req5.Header.Set("Content-Type", "application/json")
		resp5, _ := srv.App.Test(req5)
		assert.Equal(t, 401, resp5.StatusCode, "5th failure should be 401")

		// 6th attempt: 401 Unauthorized but with "account is locked" error in body (fiber returns error message)
		body, _ = json.Marshal(loginReq)
		req6 := httptest.NewRequest("POST", "/api/v1/auth/login", bytes.NewReader(body))
		req6.Header.Set("Content-Type", "application/json")
		resp6, _ := srv.App.Test(req6)

		// In our service logic, we return error if locked.
		// Handler usually maps service errors.
		assert.Equal(t, 403, resp6.StatusCode, "Should be 403 Forbidden on lockout")

		// Let's check internal state
		u, _ := uow.Users().GetByID(ctx, userID)
		assert.Equal(t, 5, u.FailedAttempts)
		assert.NotNil(t, u.LockedUntil)
		assert.True(t, u.LockedUntil.After(time.Now()))
	})

	t.Run("Successful login resets counter after lock expires", func(t *testing.T) {
		// Manually expire lock in DB
		_, err = pool.Exec(ctx, "UPDATE users SET locked_until = NOW() - INTERVAL '1 minute' WHERE user_id = $1", userID)
		require.NoError(t, err)

		// Login with correct password
		loginReq.Password = password
		body, _ := json.Marshal(loginReq)
		req := httptest.NewRequest("POST", "/api/v1/auth/login", bytes.NewReader(body))
		req.Header.Set("Content-Type", "application/json")
		resp, _ := srv.App.Test(req)
		assert.Equal(t, 200, resp.StatusCode)

		// Verify reset
		u, _ := uow.Users().GetByID(ctx, userID)
		assert.Equal(t, 0, u.FailedAttempts)
		assert.Nil(t, u.LockedUntil)
	})
}
