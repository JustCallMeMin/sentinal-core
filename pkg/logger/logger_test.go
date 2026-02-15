package logger

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"go.uber.org/zap"
)

func TestInit(t *testing.T) {
	tests := []struct {
		name  string
		level string
		env   string
	}{
		{
			name:  "Development Environment",
			level: "debug",
			env:   "development",
		},
		{
			name:  "Production Environment",
			level: "info",
			env:   "production",
		},
		{
			name:  "Invalid Level Defaults to Info",
			level: "invalid",
			env:   "production",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			assert.NotPanics(t, func() {
				Init(tt.level, tt.env)
			})
			assert.NotNil(t, Log)
		})
	}
}

func TestLogFunctions(t *testing.T) {
	Init("debug", "development")

	assert.NotPanics(t, func() {
		Debug("test debug", zap.String("key", "value"))
		Info("test info")
		Error("test error", zap.Error(assert.AnError))
	})
}
