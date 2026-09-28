package database

import (
	"database/sql"
	"fmt"
	"io/fs"
	"log"
	"os"
	"path/filepath"
	"sort"
	"strings"
	"time"

	_ "github.com/lib/pq"
)

// Connect opens a PostgreSQL connection pool and verifies connectivity
// with a short retry loop, which is useful when the DB container is
// still starting up alongside the app container in Docker Compose.
func Connect(databaseURL string) (*sql.DB, error) {
	db, err := sql.Open("postgres", databaseURL)
	if err != nil {
		return nil, fmt.Errorf("opening database: %w", err)
	}

	db.SetMaxOpenConns(20)
	db.SetMaxIdleConns(5)
	db.SetConnMaxLifetime(30 * time.Minute)

	var pingErr error
	for i := 0; i < 15; i++ {
		pingErr = db.Ping()
		if pingErr == nil {
			return db, nil
		}
		log.Printf("waiting for database... (%d/15): %v", i+1, pingErr)
		time.Sleep(2 * time.Second)
	}

	return nil, fmt.Errorf("could not connect to database: %w", pingErr)
}

// RunMigrations applies every .up.sql file in migrationsDir, in filename
// order, that has not already been recorded in the schema_migrations
// table. This is a minimal, dependency-free migration runner suitable
// for this project's needs.
func RunMigrations(db *sql.DB, migrationsDir string) error {
	if _, err := db.Exec(`
		CREATE TABLE IF NOT EXISTS schema_migrations (
			version TEXT PRIMARY KEY,
			applied_at TIMESTAMPTZ NOT NULL DEFAULT now()
		)
	`); err != nil {
		return fmt.Errorf("creating schema_migrations table: %w", err)
	}

	entries, err := os.ReadDir(migrationsDir)
	if err != nil {
		return fmt.Errorf("reading migrations dir: %w", err)
	}

	var files []string
	for _, e := range entries {
		if e.IsDir() {
			continue
		}
		if strings.HasSuffix(e.Name(), ".up.sql") {
			files = append(files, e.Name())
		}
	}
	sort.Strings(files)

	for _, name := range files {
		version := strings.TrimSuffix(name, ".up.sql")

		var count int
		if err := db.QueryRow(`SELECT COUNT(*) FROM schema_migrations WHERE version = $1`, version).Scan(&count); err != nil {
			return fmt.Errorf("checking migration %s: %w", version, err)
		}
		if count > 0 {
			continue
		}

		content, err := os.ReadFile(filepath.Join(migrationsDir, name))
		if err != nil {
			return fmt.Errorf("reading migration %s: %w", name, err)
		}

		tx, err := db.Begin()
		if err != nil {
			return fmt.Errorf("starting transaction for %s: %w", name, err)
		}

		if _, err := tx.Exec(string(content)); err != nil {
			tx.Rollback()
			return fmt.Errorf("applying migration %s: %w", name, err)
		}

		if _, err := tx.Exec(`INSERT INTO schema_migrations (version) VALUES ($1)`, version); err != nil {
			tx.Rollback()
			return fmt.Errorf("recording migration %s: %w", name, err)
		}

		if err := tx.Commit(); err != nil {
			return fmt.Errorf("committing migration %s: %w", name, err)
		}

		log.Printf("applied migration: %s", name)
	}

	return nil
}

// Seed runs a SQL file (idempotent, uses ON CONFLICT DO NOTHING) to
// populate sample data. Safe to call multiple times.
func Seed(db *sql.DB, seedFile string) error {
	content, err := os.ReadFile(seedFile)
	if err != nil {
		if os.IsNotExist(err) {
			return nil
		}
		return fmt.Errorf("reading seed file: %w", err)
	}
	if _, err := db.Exec(string(content)); err != nil {
		return fmt.Errorf("running seed file: %w", err)
	}
	log.Println("seed data applied")
	return nil
}

// EnsureUploadDir makes sure the configured upload directory exists.
func EnsureUploadDir(dir string) error {
	return os.MkdirAll(dir, fs.FileMode(0o755))
}
