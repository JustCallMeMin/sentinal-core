package server

import (
	"net/http/httptest"
	"testing"

	"github.com/gofiber/fiber/v2"
	"github.com/google/uuid"
	"github.com/sentinal/core/internal/api/auth"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

type MockTokenService struct {
	mock.Mock
}

func (m *MockTokenService) GenerateToken(userID, tenantID uuid.UUID, email string, permissions []string) (string, error) {
	args := m.Called(userID, tenantID, email, permissions)
	return args.String(0), args.Error(1)
}

func (m *MockTokenService) GenerateMFAToken(userID, tenantID uuid.UUID) (string, error) {
	args := m.Called(userID, tenantID)
	return args.String(0), args.Error(1)
}

func (m *MockTokenService) ValidateToken(tokenString string) (*auth.UserClaims, error) {
	args := m.Called(tokenString)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*auth.UserClaims), args.Error(1)
}

func TestAuthMiddleware(t *testing.T) {
	app := fiber.New()
	mockSvc := new(MockTokenService)
	middleware := AuthMiddleware(mockSvc)

	app.Get("/protected", middleware, func(c *fiber.Ctx) error {
		return c.Status(fiber.StatusOK).JSON(fiber.Map{
			"user_id":   c.Locals("user_id"),
			"tenant_id": c.Locals("tenant_id"),
		})
	})

	t.Run("Valid Token", func(t *testing.T) {
		userID := uuid.New()
		tenantID := uuid.New()
		claims := &auth.UserClaims{
			TenantID: tenantID,
			Email:    "test@example.com",
		}
		claims.Subject = userID.String()

		mockSvc.On("ValidateToken", "valid-token").Return(claims, nil).Once()

		req := httptest.NewRequest("GET", "/protected", nil)
		req.Header.Set("Authorization", "Bearer valid-token")

		resp, _ := app.Test(req)
		assert.Equal(t, 200, resp.StatusCode)
		mockSvc.AssertExpectations(t)
	})

	t.Run("Missing Header", func(t *testing.T) {
		req := httptest.NewRequest("GET", "/protected", nil)
		resp, _ := app.Test(req)
		assert.Equal(t, 401, resp.StatusCode)
	})

	t.Run("Invalid Format", func(t *testing.T) {
		req := httptest.NewRequest("GET", "/protected", nil)
		req.Header.Set("Authorization", "invalid-format")
		resp, _ := app.Test(req)
		assert.Equal(t, 401, resp.StatusCode)
	})

	t.Run("Invalid Token", func(t *testing.T) {
		mockSvc.On("ValidateToken", "bad-token").Return(nil, assert.AnError).Once()

		req := httptest.NewRequest("GET", "/protected", nil)
		req.Header.Set("Authorization", "Bearer bad-token")

		resp, _ := app.Test(req)
		assert.Equal(t, 401, resp.StatusCode)
		mockSvc.AssertExpectations(t)
	})
}
