package main

import (
	"flag"
	"fmt"
	"log"
	"os"
	"strings"

	"github.com/golang-migrate/migrate/v4"
	_ "github.com/golang-migrate/migrate/v4/database/pgx/v5"
	_ "github.com/golang-migrate/migrate/v4/source/file"
	"github.com/joho/godotenv"
)

func main() {
	// Load .env if exists (for local runs)
	_ = godotenv.Load()

	var migrationDir string
	var databaseURL string
	var reset bool
	var up bool
	var down bool

	flag.StringVar(&migrationDir, "dir", "migrations", "Directory containing migration files")
	flag.StringVar(&databaseURL, "url", os.Getenv("DATABASE_URL"), "Database URL")
	flag.BoolVar(&reset, "reset", false, "Reset database (down + up)")
	flag.BoolVar(&up, "up", false, "Run up migrations")
	flag.BoolVar(&down, "down", false, "Run down migrations")
	flag.Parse()

	if databaseURL == "" {
		log.Fatal("DATABASE_URL must be set via flag or environment variable")
	}

	// golang-migrate with pgx/v5 requires pgx5:// prefix
	if strings.HasPrefix(databaseURL, "postgres://") {
		databaseURL = strings.Replace(databaseURL, "postgres://", "pgx5://", 1)
	}

	m, err := migrate.New(
		fmt.Sprintf("file://%s", migrationDir),
		databaseURL,
	)
	if err != nil {
		log.Fatalf("Could not create migrate instance: %v", err)
	}

	defer m.Close()

	if reset {
		log.Println("Resetting database...")
		if err := m.Down(); err != nil && err != migrate.ErrNoChange {
			log.Fatalf("Down migration failed during reset: %v", err)
		}
		if err := m.Up(); err != nil && err != migrate.ErrNoChange {
			log.Fatalf("Up migration failed during reset: %v", err)
		}
		log.Println("Database reset successful")
		return
	}

	if down {
		log.Println("Running down migrations...")
		if err := m.Down(); err != nil {
			if err == migrate.ErrNoChange {
				log.Println("No migrations to rollback")
			} else {
				log.Fatalf("Down migration failed: %v", err)
			}
		} else {
			log.Println("Down migration successful")
		}
	}

	if up || (!up && !down && !reset) {
		log.Println("Running up migrations...")
		if err := m.Up(); err != nil {
			if err == migrate.ErrNoChange {
				log.Println("Database schema is up to date")
			} else {
				log.Fatalf("Up migration failed: %v", err)
			}
		} else {
			log.Println("Up migration successful")
		}
	}
}
