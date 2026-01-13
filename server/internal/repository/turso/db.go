package turso

import (
	"database/sql"
	"fmt"

	_ "github.com/tursodatabase/libsql-client-go/libsql"
	"github.com/satetsu888/agentrace/server/migrations"
)

// DB wraps the Turso database connection
type DB struct {
	*sql.DB
}

// Open opens a Turso database connection and runs migrations
func Open(databaseURL string) (*DB, error) {
	db, err := sql.Open("libsql", databaseURL)
	if err != nil {
		return nil, fmt.Errorf("failed to open turso database: %w", err)
	}

	// Test connection
	if err := db.Ping(); err != nil {
		db.Close()
		return nil, fmt.Errorf("failed to ping turso database: %w", err)
	}

	// Run migrations
	if err := runMigrations(db); err != nil {
		db.Close()
		return nil, fmt.Errorf("failed to run migrations: %w", err)
	}

	return &DB{db}, nil
}

func runMigrations(db *sql.DB) error {
	runner := migrations.NewRunner(db, migrations.DialectSQLite)
	return runner.Run()
}
