package main

import (
	"database/sql"
	"log"
	"os"

	"github.com/golang-migrate/migrate/v4"
	_ "github.com/golang-migrate/migrate/v4/source/file"
	"github.com/joho/godotenv"
)

func main() {
	// Load .env file if it exists
	_ = godotenv.Load()

	dbURL := os.Getenv("DATABASE_URL")
	if dbURL == "" {
		log.Fatal("DATABASE_URL is not set")
	}

	// Check for reset flag
	if len(os.Args) > 1 && os.Args[1] == "-reset" {
		log.Println("Reseting database (DROP SCHEMA public CASCADE)...")

		// Open direct connection
		db, err := sql.Open("postgres", dbURL)
		if err != nil {
			log.Fatalf("SQL Open failed: %v", err)
		}
		defer db.Close()

		// Nuke it from orbit
		if _, err := db.Exec("DROP SCHEMA public CASCADE; CREATE SCHEMA public; GRANT ALL ON SCHEMA public TO postgres; GRANT ALL ON SCHEMA public TO public;"); err != nil {
			log.Fatalf("Drop Schema failed: %v", err)
		}
		log.Println("Database reset successful.")
	}

	// Migration source path
	m, err := migrate.New(
		"file://migrations",
		dbURL)
	if err != nil {
		// Try to force clean if dirty state prevents new instance
		log.Printf("Migration init failed (retrying): %v", err)
	}

	if err := m.Up(); err != nil {
		if err == migrate.ErrNoChange {
			log.Println("Database already up to date")
		} else {
			log.Fatalf("Migration failed: %v", err)
		}
	} else {
		log.Println("Migration successful!")
	}
}
