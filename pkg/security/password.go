package security

import (
	"fmt"

	"golang.org/x/crypto/bcrypt"
)

// HashPassword generates a bcrypt hash of the password using the default cost.
func HashPassword(password string) (string, error) {
	if len(password) > 72 {
		return "", fmt.Errorf("password too long (max 72 characters for bcrypt)")
	}

	bytes, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	if err != nil {
		return "", fmt.Errorf("failed to hash password: %w", err)
	}
	return string(bytes), nil
}

// CheckPasswordHash compares a bcrypt hashed password with its possible
// plaintext equivalent. Returns true if the password matches the hash.
func CheckPasswordHash(password, hash string) bool {
	err := bcrypt.CompareHashAndPassword([]byte(hash), []byte(password))
	return err == nil
}
