package auth

import (
	"fmt"
	"time"

	"github.com/golang-jwt/jwt/v5"
	"github.com/google/uuid"
)

type TokenGenerator interface {
	GenerateToken(userID uuid.UUID, tenantID uuid.UUID, email string) (string, error)
}

type jwtTokenService struct {
	secret []byte
	expiry time.Duration
}

type UserClaims struct {
	jwt.RegisteredClaims
	TenantID uuid.UUID `json:"tenant_id"`
	Email    string    `json:"email"`
}

func NewTokenService(secret string, expiryHours int) TokenGenerator {
	return &jwtTokenService{
		secret: []byte(secret),
		expiry: time.Duration(expiryHours) * time.Hour,
	}
}

func (s *jwtTokenService) GenerateToken(userID uuid.UUID, tenantID uuid.UUID, email string) (string, error) {
	claims := UserClaims{
		RegisteredClaims: jwt.RegisteredClaims{
			Subject:   userID.String(),
			ExpiresAt: jwt.NewNumericDate(time.Now().Add(s.expiry)),
			IssuedAt:  jwt.NewNumericDate(time.Now()),
		},
		TenantID: tenantID,
		Email:    email,
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	signedToken, err := token.SignedString(s.secret)
	if err != nil {
		return "", fmt.Errorf("failed to sign token: %w", err)
	}

	return signedToken, nil
}
