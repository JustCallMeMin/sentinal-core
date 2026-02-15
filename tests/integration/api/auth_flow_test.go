package api_test

import (
	"bytes"
	"context"
	"encoding/json"
	"net/http/httptest"
	"testing"

	"github.com/sentinal/core/internal/api/auth"
	"github.com/sentinal/core/internal/api/transaction"
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

func TestAuthFlow_Integration(t *testing.T) {
	ctx := context.Background()
	logger.Init("error", "test")

	pool, cleanup := testutil.SetupTestDB(t)
	defer cleanup()

	uow := repository.NewUnitOfWork(pool)

	// Setup Test Data
	tenant := &models.Tenant{
		Name:            "Auth Test Merchant",
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
		Email:        "test@example.com",
		PasswordHash: hash,
		Status:       "active",
	}
	err = uow.Do(ctx, func(u repositories.UnitOfWork) error {
		return u.Users().Create(ctx, user)
	})
	require.NoError(t, err)

	// Initialize Server
	cfg := &config.Config{
		AppEnv:    "test",
		JWTSecret: "test-secret",
		JWTExpiry: 1,
	}
	tokenService := auth.NewTokenService(cfg.JWTSecret, cfg.JWTExpiry)
	authService := auth.NewService(uow, tokenService)
	authHandler := auth.NewHandler(authService)
	txService := transaction.NewService(uow)
	txHandler := transaction.NewHandler(txService)
	srv := server.New(cfg, pool, uow, txHandler, authHandler)

	t.Run("Login - Success", func(t *testing.T) {
		loginReq := auth.LoginRequest{
			TenantID: tenant.TenantID.String(),
			Email:    user.Email,
			Password: password,
		}
		body, _ := json.Marshal(loginReq)
		req := httptest.NewRequest("POST", "/api/v1/auth/login", bytes.NewReader(body))
		req.Header.Set("Content-Type", "application/json")

		resp, err := srv.App.Test(req)
		require.NoError(t, err)
		defer resp.Body.Close()

		assert.Equal(t, 200, resp.StatusCode)

		var result auth.LoginResponse
		err = json.NewDecoder(resp.Body).Decode(&result)
		require.NoError(t, err)
		assert.Equal(t, "success", result.Status)
		assert.Equal(t, user.Email, result.Email)
		assert.NotEmpty(t, result.AccessToken)
	})

	t.Run("Login - Invalid Password", func(t *testing.T) {
		loginReq := auth.LoginRequest{
			TenantID: tenant.TenantID.String(),
			Email:    user.Email,
			Password: "wrongpassword",
		}
		body, _ := json.Marshal(loginReq)
		req := httptest.NewRequest("POST", "/api/v1/auth/login", bytes.NewReader(body))
		req.Header.Set("Content-Type", "application/json")

		resp, err := srv.App.Test(req)
		require.NoError(t, err)
		assert.Equal(t, 401, resp.StatusCode)
	})

	t.Run("Login - NonExistent User", func(t *testing.T) {
		loginReq := auth.LoginRequest{
			TenantID: tenant.TenantID.String(),
			Email:    "nobody@example.com",
			Password: password,
		}
		body, _ := json.Marshal(loginReq)
		req := httptest.NewRequest("POST", "/api/v1/auth/login", bytes.NewReader(body))
		req.Header.Set("Content-Type", "application/json")

		resp, err := srv.App.Test(req)
		require.NoError(t, err)
		assert.Equal(t, 401, resp.StatusCode)
	})

	t.Run("Login - Validation Error", func(t *testing.T) {
		loginReq := auth.LoginRequest{
			Email: "invalid-email",
		}
		body, _ := json.Marshal(loginReq)
		req := httptest.NewRequest("POST", "/api/v1/auth/login", bytes.NewReader(body))
		req.Header.Set("Content-Type", "application/json")

		resp, err := srv.App.Test(req)
		require.NoError(t, err)
		assert.Equal(t, 400, resp.StatusCode)
	})
}
