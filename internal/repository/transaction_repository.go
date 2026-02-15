package repository

import (
	"context"
	"fmt"

	"github.com/georgysavva/scany/v2/pgxscan"
	"github.com/jackc/pgx/v5"
	"github.com/sentinal/core/internal/domain/models"
)

type transactionRepository struct {
	*BaseRepository[models.Transaction]
}

func NewTransactionRepository(db Querier) *transactionRepository {
	return &transactionRepository{
		BaseRepository: NewBaseRepository[models.Transaction](db, "transactions", ""),
	}
}

func (r *transactionRepository) Save(ctx context.Context, tx *models.Transaction) error {
	query := `
		INSERT INTO transactions (transaction_id, tenant_id, correlation_id, amount, currency, occurred_at, payload, created_at)
		VALUES ($1, $2, $3, $4, $5, $6, $7, NOW())
	`
	_, err := r.DB.Exec(ctx, query,
		tx.TransactionID,
		tx.TenantID,
		tx.CorrelationID,
		tx.Amount,
		tx.Currency,
		tx.OccurredAt,
		tx.Payload,
	)
	if err != nil {
		return fmt.Errorf("failed to save transaction: %w", err)
	}
	return nil
}

// SaveBulk uses DB.CopyFrom for maximum performance with large datasets
func (r *transactionRepository) SaveBulk(ctx context.Context, transactions []*models.Transaction) (int64, error) {
	rows := [][]interface{}{}
	for _, tx := range transactions {
		rows = append(rows, []interface{}{
			tx.TransactionID,
			tx.TenantID,
			tx.CorrelationID,
			tx.Amount,
			tx.Currency,
			tx.OccurredAt,
			tx.Payload,
		})
	}

	copyCount, err := r.DB.CopyFrom(
		ctx,
		pgx.Identifier{"transactions"},
		[]string{"transaction_id", "tenant_id", "correlation_id", "amount", "currency", "occurred_at", "payload"},
		pgx.CopyFromRows(rows),
	)

	if err != nil {
		return 0, fmt.Errorf("bulk copy failed: %w", err)
	}

	return copyCount, nil
}

func (r *transactionRepository) ListByTenant(ctx context.Context, tenantID string, limit, offset int) ([]*models.Transaction, error) {
	var results []*models.Transaction
	query := `SELECT * FROM transactions WHERE tenant_id = $1 ORDER BY occurred_at DESC LIMIT $2 OFFSET $3`

	if err := pgxscan.Select(ctx, r.DB, &results, query, tenantID, limit, offset); err != nil {
		return nil, fmt.Errorf("failed to list transactions for tenant: %w", err)
	}

	return results, nil
}
