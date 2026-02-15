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
	srv := server.New(cfg, pool, nil, uow, txHandler, nil, nil)

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

	t.Run("Idempotency - First Request with Correlation ID", func(t *testing.T) {
		correlationID := uuid.New()
		payload := map[string]interface{}{
			"correlation_id": correlationID.String(),
			"user_id":        "user_idempotent",
			"amount":         200.00,
			"currency":       "USD",
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

		// Store for next test
		firstTransactionID := result.TransactionID
		assert.NotEqual(t, uuid.Nil, firstTransactionID)

		// Duplicate request with same correlation_id
		t.Run("Duplicate Request Returns Same Transaction", func(t *testing.T) {
			// Same payload, same correlation_id
			body2, _ := json.Marshal(payload)
			req2 := httptest.NewRequest("POST", "/api/v1/transactions", bytes.NewReader(body2))
			req2.Header.Set("Content-Type", "application/json")
			req2.Header.Set("X-Tenant-ID", testTenant.TenantID.String())

			resp2, err := srv.App.Test(req2)
			require.NoError(t, err)
			defer resp2.Body.Close()

			assert.Equal(t, 201, resp2.StatusCode)

			var result2 transaction.CreateTransactionResponse
			err = json.NewDecoder(resp2.Body).Decode(&result2)
			require.NoError(t, err)

			// Should return SAME transaction ID (idempotent)
			assert.Equal(t, firstTransactionID, result2.TransactionID)

			// Verify only ONE transaction in DB
			var count int
			err = pool.QueryRow(ctx, "SELECT COUNT(*) FROM transactions WHERE correlation_id = $1", correlationID).Scan(&count)
			require.NoError(t, err)
			assert.Equal(t, 1, count, "Should only have 1 transaction with this correlation_id")
		})
	})

	t.Run("Idempotency - Different Tenant Same Correlation ID", func(t *testing.T) {
		// Create second tenant
		tenant2 := &models.Tenant{
			Name:            "Tenant 2",
			IndustrySegment: "Retail",
		}
		err := uow.Do(ctx, func(u repositories.UnitOfWork) error {
			return u.Tenants().Create(ctx, tenant2)
		})
		require.NoError(t, err)

		// Use same correlation_id but different tenant
		correlationID := uuid.New()
		payload := map[string]interface{}{
			"correlation_id": correlationID.String(),
			"user_id":        "user_tenant2",
			"amount":         300.00,
			"currency":       "USD",
		}

		// First request (Tenant 1)
		body1, _ := json.Marshal(payload)
		req1 := httptest.NewRequest("POST", "/api/v1/transactions", bytes.NewReader(body1))
		req1.Header.Set("Content-Type", "application/json")
		req1.Header.Set("X-Tenant-ID", testTenant.TenantID.String())

		resp1, err := srv.App.Test(req1)
		require.NoError(t, err)
		defer resp1.Body.Close()

		var result1 transaction.CreateTransactionResponse
		err = json.NewDecoder(resp1.Body).Decode(&result1)
		require.NoError(t, err)

		// Second request (Tenant 2, same correlation_id)
		body2, _ := json.Marshal(payload)
		req2 := httptest.NewRequest("POST", "/api/v1/transactions", bytes.NewReader(body2))
		req2.Header.Set("Content-Type", "application/json")
		req2.Header.Set("X-Tenant-ID", tenant2.TenantID.String())

		resp2, err := srv.App.Test(req2)
		require.NoError(t, err)
		defer resp2.Body.Close()

		var result2 transaction.CreateTransactionResponse
		err = json.NewDecoder(resp2.Body).Decode(&result2)
		require.NoError(t, err)

		// Should create NEW transaction (different tenant)
		assert.NotEqual(t, result1.TransactionID, result2.TransactionID)

		// Verify 2 transactions exist (one per tenant)
		var count int
		err = pool.QueryRow(ctx, "SELECT COUNT(*) FROM transactions WHERE correlation_id = $1", correlationID).Scan(&count)
		require.NoError(t, err)
		assert.Equal(t, 2, count, "Should have 2 transactions (one per tenant)")
	})

	t.Run("Backward Compatibility - No Correlation ID", func(t *testing.T) {
		payload := map[string]interface{}{
			"user_id":  "user_no_correlation",
			"amount":   400.00,
			"currency": "USD",
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
	})
}
