package repository

import (
	"context"
	"os"
	"testing"

	"github.com/google/uuid"
	"github.com/sentinal/core/internal/domain/models"
	"github.com/sentinal/core/internal/testutil"
	"github.com/sentinal/core/pkg/logger"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestMain(m *testing.M) {
	logger.Init("debug", "test")
	os.Exit(m.Run())
}

func TestTenantRepository_Integration(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping integration test")
	}

	pool, cleanup := testutil.SetupTestDB(t)
	defer cleanup()

	repo := NewTenantRepository(pool)
	ctx := context.Background()

	t.Run("Create and Get Tenant", func(t *testing.T) {
		tenant := &models.Tenant{
			Name:            "Test Tenant",
			IndustrySegment: "Technology",
			Settings:        map[string]interface{}{"max_users": 100},
		}

		err := repo.Create(ctx, tenant)
		require.NoError(t, err)
		require.NotEqual(t, uuid.Nil, tenant.TenantID)

		found, err := repo.GetByID(ctx, tenant.TenantID)
		require.NoError(t, err)
		require.NotNil(t, found)
		assert.Equal(t, tenant.Name, found.Name)
		assert.Equal(t, tenant.IndustrySegment, found.IndustrySegment)
	})

	t.Run("List Tenants", func(t *testing.T) {
		tenants, err := repo.List(ctx, 10, 0)
		require.NoError(t, err)
		assert.NotEmpty(t, tenants)
	})
}
