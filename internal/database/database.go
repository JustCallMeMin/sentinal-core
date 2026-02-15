package database

import (
	"context"
	"fmt"
	"sync"
	"time"

	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/sentinal/core/pkg/logger"
	"go.uber.org/zap"
)

var (
	pool *pgxpool.Pool
	once sync.Once
)

// Init initializes the database connection pool
func Init(connString string) (*pgxpool.Pool, error) {
	var err error

	once.Do(func() {
		ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
		defer cancel()

		config, parseErr := pgxpool.ParseConfig(connString)
		if parseErr != nil {
			err = fmt.Errorf("unable to parse connection string: %w", parseErr)
			return
		}

		// Optimization: Set pool settings
		config.MaxConns = 20
		config.MinConns = 5
		config.MaxConnLifetime = 30 * time.Minute

		// Connect
		dbPool, connErr := pgxpool.NewWithConfig(ctx, config)
		if connErr != nil {
			err = fmt.Errorf("unable to connect to database: %w", connErr)
			return
		}

		// Verify connection
		if pingErr := dbPool.Ping(ctx); pingErr != nil {
			err = fmt.Errorf("database ping failed: %w", pingErr)
			return
		}

		pool = dbPool
		logger.Info("Database connection pool established", zap.String("db_host", config.ConnConfig.Host))
	})

	return pool, err
}

// GetPool returns the active database connection pool
func GetPool() *pgxpool.Pool {
	return pool
}

// Close gracefully closes the database pool
func Close() {
	if pool != nil {
		logger.Info("Closing database connection pool...")
		pool.Close()
	}
}
