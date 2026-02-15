package server

import (
	"encoding/json"
	"io"
	"net/http"
	"net/http/httptest"
	"os"
	"testing"

	"github.com/sentinal/core/pkg/config"
	"github.com/sentinal/core/pkg/logger"
	"github.com/stretchr/testify/assert"
)

func TestMain(m *testing.M) {
	// Initialize logger once for all tests in this package
	logger.Init("debug", "test")
	os.Exit(m.Run())
}

func TestHealthCheck(t *testing.T) {
	// Setup
	cfg := &config.Config{
		AppEnv: "test",
		Port:   "8080",
	}
	srv := New(cfg, nil, nil, nil, nil, nil) // passing nil db/uow for basic health check

	// Create request
	req := httptest.NewRequest("GET", "/healthz", nil)

	// Perform request
	resp, err := srv.App.Test(req)

	// Assertions
	assert.NoError(t, err)
	assert.Equal(t, http.StatusOK, resp.StatusCode)

	// Validate JSON Body
	body, _ := io.ReadAll(resp.Body)
	var data map[string]interface{}
	err = json.Unmarshal(body, &data)

	assert.NoError(t, err)
	assert.Equal(t, "ok", data["status"])
}

func TestRootRoute(t *testing.T) {
	cfg := &config.Config{AppEnv: "test"}
	srv := New(cfg, nil, nil, nil, nil, nil)

	req := httptest.NewRequest("GET", "/", nil)
	resp, err := srv.App.Test(req)

	assert.NoError(t, err)
	assert.Equal(t, http.StatusOK, resp.StatusCode)
}
