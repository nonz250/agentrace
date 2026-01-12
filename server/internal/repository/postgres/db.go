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
	_, err := db.Exec(migrations.PostgresSchema)
	if err != nil {
		return err
	}

	// Add uuid column to events if not exists
	_, err = db.Exec(`
		DO $$
		BEGIN
			IF NOT EXISTS (
				SELECT 1 FROM information_schema.columns
				WHERE table_name = 'events' AND column_name = 'uuid'
			) THEN
				ALTER TABLE events ADD COLUMN uuid VARCHAR(255);
			END IF;
		END $$;
	`)
	if err != nil {
		return err
	}

	// Add unique index for (session_id, uuid) to prevent duplicate events
	_, err = db.Exec(`CREATE UNIQUE INDEX IF NOT EXISTS idx_events_session_uuid ON events(session_id, uuid) WHERE uuid IS NOT NULL`)
	if err != nil {
		return err
	}

	// Add tool_use_id column to plan_document_events if not exists
	_, err = db.Exec(`
		DO $$
		BEGIN
			IF NOT EXISTS (
				SELECT 1 FROM information_schema.columns
				WHERE table_name = 'plan_document_events' AND column_name = 'tool_use_id'
			) THEN
				ALTER TABLE plan_document_events ADD COLUMN tool_use_id VARCHAR(255);
			END IF;
		END $$;
	`)
	if err != nil {
		return err
	}

	// Add message column to plan_document_events if not exists
	_, err = db.Exec(`
		DO $$
		BEGIN
			IF NOT EXISTS (
				SELECT 1 FROM information_schema.columns
				WHERE table_name = 'plan_document_events' AND column_name = 'message'
			) THEN
				ALTER TABLE plan_document_events ADD COLUMN message TEXT NOT NULL DEFAULT '';
			END IF;
		END $$;
	`)
	return err
}
