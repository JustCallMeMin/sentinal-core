package api_test

import (
	"bytes"
	"context"
	"encoding/json"
	"net/http/httptest"
	"testing"

	"github.com/gofiber/fiber/v2"
	"github.com/golang-jwt/jwt/v5"
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

func TestRBACFlow_Integration(t *testing.T) {
	ctx := context.Background()
	logger.Init("error", "test")

	pool, cleanup := testutil.SetupTestDB(t)
	defer cleanup()

	uow := repository.NewUnitOfWork(pool)

	// Setup Tenant
	tenant := &models.Tenant{
		Name:            "RBAC Test Merchant",
		IndustrySegment: "Test",
	}
	err := uow.Do(ctx, func(u repositories.UnitOfWork) error {
		return u.Tenants().Create(ctx, tenant)
	})
	require.NoError(t, err)

	// Fetch Role IDs seeded by migration
	var adminRoleID, viewerRoleID int
	err = pool.QueryRow(ctx, "SELECT role_id FROM roles WHERE name = 'SUPER_ADMIN' AND tenant_id IS NULL").Scan(&adminRoleID)
	require.NoError(t, err)
	err = pool.QueryRow(ctx, "SELECT role_id FROM roles WHERE name = 'VIEWER' AND tenant_id IS NULL").Scan(&viewerRoleID)
	require.NoError(t, err)

	password := "secure123"
	hash, _ := security.HashPassword(password)

	// Create Admin User
	adminUser := &models.User{
		TenantID:     tenant.TenantID,
		Email:        "admin@example.com",
		PasswordHash: hash,
		RoleID:       &adminRoleID,
		Status:       "active",
	}
	// Create Viewer User
	viewerUser := &models.User{
		TenantID:     tenant.TenantID,
		Email:        "viewer@example.com",
		PasswordHash: hash,
		RoleID:       &viewerRoleID,
		Status:       "active",
	}

	err = uow.Do(ctx, func(u repositories.UnitOfWork) error {
		if err := u.Users().Create(ctx, adminUser); err != nil {
			return err
		}
		return u.Users().Create(ctx, viewerUser)
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

	// Register a test route protected by RBAC
	srv.App.Get("/test/admin-only",
		server.AuthMiddleware(tokenService),
		server.RBACMiddleware("users:manage"),
		func(c *fiber.Ctx) error {
			return c.SendStatus(200)
		},
	)

	t.Run("Admin Login & Permissions Verification", func(t *testing.T) {
		loginReq := auth.LoginRequest{
			TenantID: tenant.TenantID.String(),
			Email:    adminUser.Email,
			Password: password,
		}
		body, _ := json.Marshal(loginReq)
		req := httptest.NewRequest("POST", "/api/v1/auth/login", bytes.NewReader(body))
		req.Header.Set("Content-Type", "application/json")

		resp, _ := srv.App.Test(req)
		assert.Equal(t, 200, resp.StatusCode)

		var result auth.LoginResponse
		json.NewDecoder(resp.Body).Decode(&result)

		// Verify JWT Claims
		claims := &auth.UserClaims{}
		_, err := jwt.ParseWithClaims(result.AccessToken, claims, func(t *jwt.Token) (interface{}, error) {
			return []byte(cfg.JWTSecret), nil
		})
		require.NoError(t, err)
		assert.Contains(t, claims.Permissions, "users:manage")
		assert.Contains(t, claims.Permissions, "transactions:read")
		assert.Contains(t, claims.Permissions, "transactions:write")
	})

	t.Run("Access Authorized Route", func(t *testing.T) {
		// Get Admin Token
		loginReq := auth.LoginRequest{
			TenantID: tenant.TenantID.String(),
			Email:    adminUser.Email,
			Password: password,
		}
		body, _ := json.Marshal(loginReq)
		reqLogin := httptest.NewRequest("POST", "/api/v1/auth/login", bytes.NewReader(body))
		reqLogin.Header.Set("Content-Type", "application/json")
		respLogin, _ := srv.App.Test(reqLogin)
		var result auth.LoginResponse
		json.NewDecoder(respLogin.Body).Decode(&result)

		// Request protected route
		req := httptest.NewRequest("GET", "/test/admin-only", nil)
		req.Header.Set("Authorization", "Bearer "+result.AccessToken)
		resp, _ := srv.App.Test(req)
		assert.Equal(t, 200, resp.StatusCode)
	})

	t.Run("Access Unauthorized Route (Viewer tries Admin route)", func(t *testing.T) {
		// Get Viewer Token
		loginReq := auth.LoginRequest{
			TenantID: tenant.TenantID.String(),
			Email:    viewerUser.Email,
			Password: password,
		}
		body, _ := json.Marshal(loginReq)
		reqLogin := httptest.NewRequest("POST", "/api/v1/auth/login", bytes.NewReader(body))
		reqLogin.Header.Set("Content-Type", "application/json")
		respLogin, _ := srv.App.Test(reqLogin)
		var result auth.LoginResponse
		json.NewDecoder(respLogin.Body).Decode(&result)

		// Request protected route (requires users:manage, viewer only has transactions:read)
		req := httptest.NewRequest("GET", "/test/admin-only", nil)
		req.Header.Set("Authorization", "Bearer "+result.AccessToken)
		resp, _ := srv.App.Test(req)
		assert.Equal(t, 403, resp.StatusCode)
	})
}
