package config

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestConfig_Validate(t *testing.T) {
	tests := []struct {
		name    string
		config  Config
		wantErr bool
	}{
		{
			name: "Valid Production Config",
			config: Config{
				AppEnv:      "production",
				Port:        "8080",
				DatabaseURL: "postgres://user:pass@localhost:5432/db",
				LogLevel:    "info",
				JWTSecret:   "production-secret-key-at-least-32-chars",
			},
			wantErr: false,
		},
		{
			name: "Valid Development Config",
			config: Config{
				AppEnv:      "development",
				Port:        "3000",
				DatabaseURL: "http://localhost:5432", // is.URL accepts this
				LogLevel:    "debug",
			},
			wantErr: false,
		},
		{
			name: "Missing DatabaseURL",
			config: Config{
				AppEnv: "development",
				Port:   "8080",
			},
			wantErr: true,
		},
		{
			name: "Invalid AppEnv",
			config: Config{
				AppEnv:      "staging", // staging is not in our allowed list
				Port:        "8080",
				DatabaseURL: "postgres://localhost:5432",
			},
			wantErr: true,
		},
		{
			name: "Missing Port",
			config: Config{
				AppEnv:      "development",
				DatabaseURL: "postgres://localhost:5432",
			},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := tt.config.Validate()
			if tt.wantErr {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
			}
		})
	}
}
