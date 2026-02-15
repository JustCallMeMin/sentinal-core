package repository

import (
	"errors"
	"fmt"

	"github.com/jackc/pgx/v5/pgconn"
)

var (
	ErrNotFound          = errors.New("resource not found")
	ErrAlreadyExists     = errors.New("resource already exists")
	ErrForeignKey        = errors.New("foreign key violation")
	ErrInvalidConstraint = errors.New("invalid constraint")
	ErrInternalDB        = errors.New("internal database error")
)

// MapError converts standard Postgres errors to domain errors
func MapError(err error) error {
	if err == nil {
		return nil
	}

	// Handle specific pgconn.PgError
	var pgErr *pgconn.PgError
	if errors.As(err, &pgErr) {
		switch pgErr.Code {
		case "23505": // unique_violation
			return fmt.Errorf("%w: %s", ErrAlreadyExists, pgErr.Detail)
		case "23503": // foreign_key_violation
			return fmt.Errorf("%w: %s", ErrForeignKey, pgErr.Detail)
		case "23502": // not_null_violation
			return fmt.Errorf("%w: %s", ErrInvalidConstraint, pgErr.ColumnName)
		case "23514": // check_violation
			return fmt.Errorf("%w: %s", ErrInvalidConstraint, pgErr.ConstraintName)
		}
	}

	// Default to internal error if not mapped
	return fmt.Errorf("%w: %v", ErrInternalDB, err)
}
