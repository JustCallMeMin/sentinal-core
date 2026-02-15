package repository

import (
	"context"
	"fmt"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/sentinal/core/internal/domain/models"
	"github.com/sentinal/core/internal/domain/repositories"
)

type unitOfWork struct {
	pool *pgxpool.Pool
	db   Querier // Can be pool or tx
}

func NewUnitOfWork(pool *pgxpool.Pool) repositories.UnitOfWork {
	return &unitOfWork{
		pool: pool,
		db:   pool,
	}
}

func (u *unitOfWork) Do(ctx context.Context, fn func(repositories.UnitOfWork) error) error {
	// If already in a transaction (db is not the pool), just execute
	if _, ok := u.db.(pgx.Tx); ok {
		return fn(u)
	}

	// Start new transaction
	tx, err := u.pool.Begin(ctx)
	if err != nil {
		return fmt.Errorf("failed to begin transaction: %w", err)
	}

	// Ensure cleanup
	defer func() {
		if p := recover(); p != nil {
			_ = tx.Rollback(ctx)
			panic(p) // re-throw panic after rollback
		}
	}()

	// Create a new UoW tied to this transaction
	txUow := &unitOfWork{
		pool: u.pool,
		db:   tx,
	}

	if err := fn(txUow); err != nil {
		if rbErr := tx.Rollback(ctx); rbErr != nil {
			return fmt.Errorf("rollback failed: %v (original error: %w)", rbErr, err)
		}
		return err
	}

	if err := tx.Commit(ctx); err != nil {
		return fmt.Errorf("commit failed: %w", err)
	}

	return nil
}

func (u *unitOfWork) Tenants() repositories.TenantRepository {
	return &tenantRepository{
		BaseRepository: NewBaseRepository[models.Tenant](u.db, "tenants"),
	}
}

func (u *unitOfWork) Transactions() repositories.TransactionRepository {
	return &transactionRepository{
		BaseRepository: NewBaseRepository[models.Transaction](u.db, "transactions"),
	}
}
