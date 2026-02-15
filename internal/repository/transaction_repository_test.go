package repository

import (
	"context"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/sentinal/core/internal/domain/models"
	"github.com/sentinal/core/internal/testutil"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestTransactionRepository_Integration(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping integration test")
	}

	pool, cleanup := testutil.SetupTestDB(t)
	defer cleanup()

	tenantRepo := NewTenantRepository(pool)
	txRepo := NewTransactionRepository(pool)
	ctx := context.Background()

	// 1. Setup a tenant first (needed for FK if we had them or for partitioning)
	tenant := &models.Tenant{
		Name:            "Tx Test Tenant",
		IndustrySegment: "Finance",
	}
	err := tenantRepo.Create(ctx, tenant)
	require.NoError(t, err)

	t.Run("Save and List Transactions", func(t *testing.T) {
		tx := &models.Transaction{
			TransactionID: uuid.New(),
			TenantID:      tenant.TenantID,
			Amount:        99.99,
			Currency:      "USD",
			OccurredAt:    time.Now(),
			Payload:       map[string]interface{}{"card_type": "visa"},
		}

		err := txRepo.Save(ctx, tx)
		require.NoError(t, err)

		results, err := txRepo.ListByTenant(ctx, tenant.TenantID.String(), 10, 0)
		require.NoError(t, err)
		assert.Len(t, results, 1)
		assert.Equal(t, tx.Amount, results[0].Amount)
	})

	t.Run("Bulk Save Transactions", func(t *testing.T) {
		txs := []*models.Transaction{
			{
				TransactionID: uuid.New(),
				TenantID:      tenant.TenantID,
				Amount:        10.0,
				Currency:      "USD",
				OccurredAt:    time.Now(),
			},
			{
				TransactionID: uuid.New(),
				TenantID:      tenant.TenantID,
				Amount:        20.0,
				Currency:      "USD",
				OccurredAt:    time.Now(),
			},
		}

		count, err := txRepo.SaveBulk(ctx, txs)
		require.NoError(t, err)
		assert.Equal(t, int64(2), count)

		results, err := txRepo.ListByTenant(ctx, tenant.TenantID.String(), 10, 0)
		require.NoError(t, err)
		// Total should be 3 (1 from previous test + 2 bulk)
		assert.Len(t, results, 3)
	})
}
