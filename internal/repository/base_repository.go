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
	DB               Querier
	TableName        string
	SoftDeleteColumn string
}

// NewBaseRepository creates a new generic repository
func NewBaseRepository[T any](db Querier, tableName string, softDeleteCol string) *BaseRepository[T] {
	return &BaseRepository[T]{
		DB:               db,
		TableName:        tableName,
		SoftDeleteColumn: softDeleteCol,
	}
}

// GetByID fetches a single entity by ID
func (r *BaseRepository[T]) GetByID(ctx context.Context, id interface{}, idColumn string) (*T, error) {
	var entity T
	query := fmt.Sprintf("SELECT * FROM %s WHERE %s = $1", r.TableName, idColumn)

	if r.SoftDeleteColumn != "" {
		query += fmt.Sprintf(" AND %s IS NULL", r.SoftDeleteColumn)
	}
	query += " LIMIT 1"

	if err := pgxscan.Get(ctx, r.DB, &entity, query, id); err != nil {
		if pgxscan.NotFound(err) {
			return nil, ErrNotFound
		}
		return nil, MapError(err)
	}

	return &entity, nil
}

// ListAll returns records with limit/offset
func (r *BaseRepository[T]) ListAll(ctx context.Context, limit, offset int) ([]*T, error) {
	var entities []*T
	query := fmt.Sprintf("SELECT * FROM %s", r.TableName)

	if r.SoftDeleteColumn != "" {
		query += fmt.Sprintf(" WHERE %s IS NULL", r.SoftDeleteColumn)
	}

	// Always order by something deterministic if possible, but for generic base just limit/offset
	query += " LIMIT $1 OFFSET $2"

	if err := pgxscan.Select(ctx, r.DB, &entities, query, limit, offset); err != nil {
		return nil, MapError(err)
	}

	return entities, nil
}

// Count returns the total number of records
func (r *BaseRepository[T]) Count(ctx context.Context) (int, error) {
	var count int
	query := fmt.Sprintf("SELECT COUNT(*) FROM %s", r.TableName)

	if r.SoftDeleteColumn != "" {
		query += fmt.Sprintf(" WHERE %s IS NULL", r.SoftDeleteColumn)
	}

	if err := r.DB.QueryRow(ctx, query).Scan(&count); err != nil {
		return 0, MapError(err)
	}

	return count, nil
}

// Delete performs a soft delete if configured, otherwise hard delete
func (r *BaseRepository[T]) Delete(ctx context.Context, id interface{}, idColumn string) error {
	var query string
	if r.SoftDeleteColumn != "" {
		query = fmt.Sprintf("UPDATE %s SET %s = NOW() WHERE %s = $1", r.TableName, r.SoftDeleteColumn, idColumn)
	} else {
		query = fmt.Sprintf("DELETE FROM %s WHERE %s = $1", r.TableName, idColumn)
	}

	if _, err := r.DB.Exec(ctx, query, id); err != nil {
		return MapError(err)
	}
	return nil
}
