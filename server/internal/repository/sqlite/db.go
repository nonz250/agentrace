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
	if err != nil {
		return err
	}

	// Add updated_at column to sessions if not exists
	var colExists int
	row := db.QueryRow(`SELECT COUNT(*) FROM pragma_table_info('sessions') WHERE name='updated_at'`)
	if err := row.Scan(&colExists); err != nil {
		return err
	}
	if colExists == 0 {
		_, err = db.Exec(`ALTER TABLE sessions ADD COLUMN updated_at TEXT`)
		if err != nil {
			return err
		}
		// Set default value for existing rows (use started_at as initial updated_at)
		_, err = db.Exec(`UPDATE sessions SET updated_at = started_at WHERE updated_at IS NULL`)
		if err != nil {
			return err
		}
	}

	// Add index for updated_at if not exists
	_, err = db.Exec(`CREATE INDEX IF NOT EXISTS idx_sessions_updated ON sessions(updated_at)`)
	if err != nil {
		return err
	}

	// Add uuid column to events if not exists
	var uuidColExists int
	row = db.QueryRow(`SELECT COUNT(*) FROM pragma_table_info('events') WHERE name='uuid'`)
	if err := row.Scan(&uuidColExists); err != nil {
		return err
	}
	if uuidColExists == 0 {
		_, err = db.Exec(`ALTER TABLE events ADD COLUMN uuid TEXT`)
		if err != nil {
			return err
		}
	}

	// Add unique index for (session_id, uuid) to prevent duplicate events
	_, err = db.Exec(`CREATE UNIQUE INDEX IF NOT EXISTS idx_events_session_uuid ON events(session_id, uuid) WHERE uuid IS NOT NULL`)
	if err != nil {
		return err
	}

	// Add tool_use_id column to plan_document_events if not exists
	var toolUseIdColExists int
	row = db.QueryRow(`SELECT COUNT(*) FROM pragma_table_info('plan_document_events') WHERE name='tool_use_id'`)
	if err := row.Scan(&toolUseIdColExists); err != nil {
		return err
	}
	if toolUseIdColExists == 0 {
		_, err = db.Exec(`ALTER TABLE plan_document_events ADD COLUMN tool_use_id TEXT`)
		if err != nil {
			return err
		}
	}

	return nil
}
