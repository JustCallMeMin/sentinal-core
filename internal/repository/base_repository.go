package repository

import (
	"context"
	"fmt"

	"github.com/georgysavva/scany/v2/pgxscan"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"
)

// Querier is a common interface for pgxpool.Pool and pgx.Tx
type Querier interface {
	pgxscan.Querier
	Exec(ctx context.Context, sql string, arguments ...interface{}) (pgconn.CommandTag, error)
	QueryRow(ctx context.Context, sql string, args ...interface{}) pgx.Row
	CopyFrom(ctx context.Context, tableName pgx.Identifier, columnNames []string, rowSrc pgx.CopyFromSource) (int64, error)
}

// BaseRepository defines common DB operations using Generics
type BaseRepository[T any] struct {
	DB        Querier
	TableName string
}

// NewBaseRepository creates a new generic repository
func NewBaseRepository[T any](db Querier, tableName string) *BaseRepository[T] {
	return &BaseRepository[T]{
		DB:        db,
		TableName: tableName,
	}
}

// GetByID fetches a single entity by ID
func (r *BaseRepository[T]) GetByID(ctx context.Context, id interface{}, idColumn string) (*T, error) {
	var entity T
	query := fmt.Sprintf("SELECT * FROM %s WHERE %s = $1 LIMIT 1", r.TableName, idColumn)

	if err := pgxscan.Get(ctx, r.DB, &entity, query, id); err != nil {
		return nil, fmt.Errorf("failed to get %s by id: %w", r.TableName, err)
	}

	return &entity, nil
}

// ListAll returns records with limit/offset
func (r *BaseRepository[T]) ListAll(ctx context.Context, limit, offset int) ([]*T, error) {
	var entities []*T
	query := fmt.Sprintf("SELECT * FROM %s LIMIT $1 OFFSET $2", r.TableName)

	if err := pgxscan.Select(ctx, r.DB, &entities, query, limit, offset); err != nil {
		return nil, fmt.Errorf("failed to list %s: %w", r.TableName, err)
	}

	return entities, nil
}

// Count returns the total number of records
func (r *BaseRepository[T]) Count(ctx context.Context) (int, error) {
	var count int
	query := fmt.Sprintf("SELECT COUNT(*) FROM %s", r.TableName)

	if err := r.DB.QueryRow(ctx, query).Scan(&count); err != nil {
		return 0, fmt.Errorf("failed to count records in %s: %w", r.TableName, err)
	}

	return count, nil
}
