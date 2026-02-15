package config

import (
	"os"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestLoadConfig(t *testing.T) {
	// Set some env variables
	os.Setenv("SERVER_PORT", "9999")
	os.Setenv("AUTH_MAX_FAILED_ATTEMPTS", "10")
	defer os.Unsetenv("SERVER_PORT")
	defer os.Unsetenv("AUTH_MAX_FAILED_ATTEMPTS")

	cfg, err := LoadConfig()
	assert.NoError(t, err)
	assert.Equal(t, "9999", cfg.Port)
	assert.Equal(t, 10, cfg.AuthMaxFailedAttempts)
	assert.Equal(t, 15, cfg.AuthLockoutMinutes) // default
}
