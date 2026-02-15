package auth

import (
	"testing"
	"time"

	"github.com/golang-jwt/jwt/v5"
	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestTokenService(t *testing.T) {
	secret := "my-auth-secret-key"
	expiryHours := 24
	service := NewTokenService(secret, expiryHours)

	userID := uuid.New()
	tenantID := uuid.New()
	email := "test-jwt@example.com"
	permissions := []string{"transactions:read"}

	t.Run("Generate Valid Token", func(t *testing.T) {
		token, err := service.GenerateToken(userID, tenantID, email, permissions)
		require.NoError(t, err)
		assert.NotEmpty(t, token)

		// Parse back to verify claims
		parsedToken, err := jwt.ParseWithClaims(token, &UserClaims{}, func(t *jwt.Token) (interface{}, error) {
			return []byte(secret), nil
		})

		require.NoError(t, err)
		assert.True(t, parsedToken.Valid)

		claims, ok := parsedToken.Claims.(*UserClaims)
		assert.True(t, ok)
		assert.Equal(t, userID.String(), claims.Subject)
		assert.Equal(t, tenantID, claims.TenantID)
		assert.Equal(t, email, claims.Email)
		assert.Equal(t, permissions, claims.Permissions)
		assert.NotEmpty(t, claims.ID) // Verify JTI (jwt.RegisteredClaims.ID)

		// Verify expiry (within reasonable range)
		expiry := claims.ExpiresAt.Time
		assert.WithinDuration(t, time.Now().Add(time.Duration(expiryHours)*time.Hour), expiry, 5*time.Second)
	})

	t.Run("Invalid Secret Fails Verification", func(t *testing.T) {
		token, err := service.GenerateToken(userID, tenantID, email, nil)
		require.NoError(t, err)

		_, err = jwt.Parse(token, func(t *jwt.Token) (interface{}, error) {
			return []byte("wrong-secret"), nil
		})
		assert.Error(t, err)
	})
}
