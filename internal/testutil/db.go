package testutil

import (
	"context"
	"path/filepath"
	"runtime"
	"testing"
	"time"

	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/sentinal/core/internal/database"
	"github.com/testcontainers/testcontainers-go"
	"github.com/testcontainers/testcontainers-go/modules/postgres"
	"github.com/testcontainers/testcontainers-go/wait"
)

// SetupTestDB starts a PostgreSQL container, runs migrations, and returns a connection pool
func SetupTestDB(t *testing.T) (*pgxpool.Pool, func()) {
	ctx := context.Background()

	// Find migrations directory relative to this file
	_, b, _, _ := runtime.Caller(0)
	basepath := filepath.Dir(b)
	migrationsDir := filepath.Join(basepath, "..", "..", "migrations")

	pgContainer, err := postgres.Run(ctx,
		"postgres:16-alpine",
		postgres.WithDatabase("sentinal_test"),
		postgres.WithUsername("sentinal"),
		postgres.WithPassword("password"),
		testcontainers.WithWaitStrategy(
			wait.ForLog("database system is ready to accept connections").
				WithOccurrence(2).
				WithStartupTimeout(30*time.Second)),
	)
	if err != nil {
		t.Fatalf("failed to start container: %s", err)
	}

	connStr, err := pgContainer.ConnectionString(ctx, "sslmode=disable")
	if err != nil {
		t.Fatalf("failed to get connection string: %s", err)
	}

	// Run migrations
	if err := database.RunMigrations(migrationsDir, connStr); err != nil {
		t.Fatalf("failed to run migrations: %s", err)
	}

	// Initialize pool
	pool, err := database.NewPool(connStr)
	if err != nil {
		t.Fatalf("failed to initialize database pool: %s", err)
	}

	cleanup := func() {
		pool.Close()
		if err := pgContainer.Terminate(ctx); err != nil {
			t.Logf("failed to terminate container: %s", err)
		}
	}

	return pool, cleanup
}
