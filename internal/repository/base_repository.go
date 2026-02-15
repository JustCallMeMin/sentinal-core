package repository

import (
	"context"
	"fmt"

	"github.com/georgysavva/scany/v2/pgxscan"
	"github.com/jackc/pgx/v5/pgxpool"
)

// BaseRepository defines common DB operations using Generics
type BaseRepository[T any] struct {
	Pool      *pgxpool.Pool
	TableName string
}

// NewBaseRepository creates a new generic repository
func NewBaseRepository[T any](pool *pgxpool.Pool, tableName string) *BaseRepository[T] {
	return &BaseRepository[T]{
		Pool:      pool,
		TableName: tableName,
	}
}

// GetByID fetches a single entity by ID (Note: requires "id" column or override)
func (r *BaseRepository[T]) GetByID(ctx context.Context, id interface{}, idColumn string) (*T, error) {
	var entity T
	query := fmt.Sprintf("SELECT * FROM %s WHERE %s = $1 LIMIT 1", r.TableName, idColumn)

	if err := pgxscan.Get(ctx, r.Pool, &entity, query, id); err != nil {
		return nil, fmt.Errorf("failed to get %s by id: %w", r.TableName, err)
	}

	return &entity, nil
}

// ListAll returns all records from the table (Caution for large tables)
func (r *BaseRepository[T]) ListAll(ctx context.Context, limit, offset int) ([]*T, error) {
	var entities []*T
	query := fmt.Sprintf("SELECT * FROM %s LIMIT $1 OFFSET $2", r.TableName)

	if err := pgxscan.Select(ctx, r.Pool, &entities, query, limit, offset); err != nil {
		return nil, fmt.Errorf("failed to list %s: %w", r.TableName, err)
	}

	return entities, nil
}

// Count returns the total number of records
func (r *BaseRepository[T]) Count(ctx context.Context) (int, error) {
	var count int
	query := fmt.Sprintf("SELECT COUNT(*) FROM %s", r.TableName)

	if err := r.Pool.QueryRow(ctx, query).Scan(&count); err != nil {
		return 0, fmt.Errorf("failed to count records in %s: %w", r.TableName, err)
	}

	return count, nil
}
