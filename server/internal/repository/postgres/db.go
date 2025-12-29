package postgres

import (
	"database/sql"
	"fmt"

	_ "github.com/lib/pq"
	"github.com/satetsu888/agentrace/server/migrations"
)

// DB wraps the PostgreSQL database connection
type DB struct {
	*sql.DB
}

// Open opens a PostgreSQL database connection and runs migrations
func Open(databaseURL string) (*DB, error) {
	db, err := sql.Open("postgres", databaseURL)
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
	_, err := db.Exec(migrations.PostgresInitial)
	return err
}
