package api_test

import (
	"bytes"
	"context"
	"encoding/json"
	"net/http/httptest"
	"testing"

	"github.com/google/uuid"
	"github.com/sentinal/core/internal/api/transaction"
	"github.com/sentinal/core/internal/domain/models"
	"github.com/sentinal/core/internal/domain/repositories"
	"github.com/sentinal/core/internal/repository"
	"github.com/sentinal/core/internal/server"
	"github.com/sentinal/core/internal/testutil"
	"github.com/sentinal/core/pkg/config"
	"github.com/sentinal/core/pkg/logger"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestTransactionIngestion_Integration(t *testing.T) {
	ctx := context.Background()

	// Initialize logger for test
	logger.Init("error", "test")

	// Setup test database
	pool, cleanup := testutil.SetupTestDB(t)
	defer cleanup()

	// Create test tenant
	uow := repository.NewUnitOfWork(pool)
	testTenant := &models.Tenant{
		Name:            "Test Merchant",
		IndustrySegment: "E-commerce",
	}
	err := uow.Do(ctx, func(u repositories.UnitOfWork) error {
		return u.Tenants().Create(ctx, testTenant)
	})
	require.NoError(t, err)

	// Initialize server with dependencies
	cfg := &config.Config{AppEnv: "test"}
	txService := transaction.NewService(uow)
	txHandler := transaction.NewHandler(txService)
	srv := server.New(cfg, pool, uow, txHandler)

	t.Run("Create Transaction - Success", func(t *testing.T) {
		payload := map[string]interface{}{
			"user_id":    "user_12345",
			"amount":     150.50,
			"currency":   "USD",
			"ip_address": "192.168.1.1",
			"device_id":  "device_xyz",
			"payload": map[string]interface{}{
				"item_category": "electronics",
			},
		}

		body, _ := json.Marshal(payload)
		req := httptest.NewRequest("POST", "/api/v1/transactions", bytes.NewReader(body))
		req.Header.Set("Content-Type", "application/json")
		req.Header.Set("X-Tenant-ID", testTenant.TenantID.String())

		resp, err := srv.App.Test(req)
		require.NoError(t, err)
		defer resp.Body.Close()

		assert.Equal(t, 201, resp.StatusCode)

		var result transaction.CreateTransactionResponse
		err = json.NewDecoder(resp.Body).Decode(&result)
		require.NoError(t, err)

		assert.NotEqual(t, uuid.Nil, result.TransactionID)
		assert.Equal(t, "received", result.Status)

		// Verify DB persistence
		var count int
		err = pool.QueryRow(ctx, "SELECT COUNT(*) FROM transactions WHERE transaction_id = $1", result.TransactionID).Scan(&count)
		require.NoError(t, err)
		assert.Equal(t, 1, count)
	})

	t.Run("Create Transaction - Validation Error", func(t *testing.T) {
		payload := map[string]interface{}{
			"user_id":  "user_123",
			"amount":   -10.0, // Invalid: negative amount
			"currency": "USD",
		}

		body, _ := json.Marshal(payload)
		req := httptest.NewRequest("POST", "/api/v1/transactions", bytes.NewReader(body))
		req.Header.Set("Content-Type", "application/json")
		req.Header.Set("X-Tenant-ID", testTenant.TenantID.String())

		resp, err := srv.App.Test(req)
		require.NoError(t, err)
		defer resp.Body.Close()

		assert.Equal(t, 400, resp.StatusCode)
	})

	t.Run("Create Transaction - Missing Tenant ID", func(t *testing.T) {
		payload := map[string]interface{}{
			"user_id":  "user_123",
			"amount":   100.0,
			"currency": "USD",
		}

		body, _ := json.Marshal(payload)
		req := httptest.NewRequest("POST", "/api/v1/transactions", bytes.NewReader(body))
		req.Header.Set("Content-Type", "application/json")
		// No X-Tenant-ID header

		resp, err := srv.App.Test(req)
		require.NoError(t, err)
		defer resp.Body.Close()

		assert.Equal(t, 401, resp.StatusCode)
	})
}
