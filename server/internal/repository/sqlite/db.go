package sqlite

import (
	"database/sql"
	"fmt"
	"os"
	"path/filepath"

	_ "github.com/mattn/go-sqlite3"
	"github.com/satetsu888/agentrace/server/migrations"
)

// DB wraps the SQLite database connection
type DB struct {
	*sql.DB
}

// Open opens a SQLite database connection and runs migrations
func Open(databaseURL string) (*DB, error) {
	// Ensure directory exists
	dir := filepath.Dir(databaseURL)
	if dir != "." && dir != "" {
		if err := os.MkdirAll(dir, 0755); err != nil {
			return nil, fmt.Errorf("failed to create database directory: %w", err)
		}
	}

	db, err := sql.Open("sqlite3", databaseURL+"?_foreign_keys=on")
	if err != nil {
		return nil, fmt.Errorf("failed to open database: %w", err)
	}

	// Test connection
	if err := db.Ping(); err != nil {
		db.Close()
		return nil, fmt.Errorf("failed to ping database: %w", err)
	}

	// Run migrations
	if err := runMigrations(db); err != nil {
		db.Close()
		return nil, fmt.Errorf("failed to run migrations: %w", err)
	}

	return &DB{db}, nil
}

func runMigrations(db *sql.DB) error {
	_, err := db.Exec(migrations.SQLiteSchema)
	return err
}
