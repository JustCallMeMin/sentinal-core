package database

import (
	"fmt"
	"strings"

	"github.com/golang-migrate/migrate/v4"
	_ "github.com/golang-migrate/migrate/v4/database/pgx/v5"
	_ "github.com/golang-migrate/migrate/v4/source/file"
)

// RunMigrations executes migrations from the given directory on the given DB URL
func RunMigrations(migrationDir, databaseURL string) error {
	// golang-migrate with pgx/v5 requires pgx5:// prefix
	if strings.HasPrefix(databaseURL, "postgres://") {
		databaseURL = strings.Replace(databaseURL, "postgres://", "pgx5://", 1)
	}

	m, err := migrate.New(
		fmt.Sprintf("file://%s", migrationDir),
		databaseURL,
	)
	if err != nil {
		return fmt.Errorf("could not create migrate instance: %w", err)
	}
	defer m.Close()

	if err := m.Up(); err != nil && err != migrate.ErrNoChange {
		return fmt.Errorf("migration failed: %w", err)
	}

	return nil
}
