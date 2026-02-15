package auth

import (
	"fmt"
	"time"

	"github.com/golang-jwt/jwt/v5"
	"github.com/google/uuid"
)

type TokenService interface {
	GenerateToken(userID uuid.UUID, tenantID uuid.UUID, email string, permissions []string) (string, error)
	ValidateToken(tokenString string) (*UserClaims, error)
}

type jwtTokenService struct {
	secret []byte
	expiry time.Duration
}

type UserClaims struct {
	jwt.RegisteredClaims
	TenantID    uuid.UUID `json:"tenant_id"`
	Email       string    `json:"email"`
	Permissions []string  `json:"permissions"`
}

func NewTokenService(secret string, expiryHours int) TokenService {
	return &jwtTokenService{
		secret: []byte(secret),
		expiry: time.Duration(expiryHours) * time.Hour,
	}
}

func (s *jwtTokenService) ValidateToken(tokenString string) (*UserClaims, error) {
	token, err := jwt.ParseWithClaims(tokenString, &UserClaims{}, func(t *jwt.Token) (interface{}, error) {
		if _, ok := t.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, fmt.Errorf("unexpected signing method: %v", t.Header["alg"])
		}
		return s.secret, nil
	})

	if err != nil {
		return nil, fmt.Errorf("invalid token: %w", err)
	}

	if claims, ok := token.Claims.(*UserClaims); ok && token.Valid {
		return claims, nil
	}

	return nil, fmt.Errorf("invalid token claims")
}

func (s *jwtTokenService) GenerateToken(userID uuid.UUID, tenantID uuid.UUID, email string, permissions []string) (string, error) {
	claims := UserClaims{
		RegisteredClaims: jwt.RegisteredClaims{
			Subject:   userID.String(),
			ExpiresAt: jwt.NewNumericDate(time.Now().Add(s.expiry)),
			IssuedAt:  jwt.NewNumericDate(time.Now()),
		},
		TenantID:    tenantID,
		Email:       email,
		Permissions: permissions,
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	signedToken, err := token.SignedString(s.secret)
	if err != nil {
		return "", fmt.Errorf("failed to sign token: %w", err)
	}

	return signedToken, nil
}
