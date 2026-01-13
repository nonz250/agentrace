package migrations

import (
	"database/sql"
	"fmt"
	"sort"

	"golang.org/x/mod/semver"
)

// Dialect represents the database dialect
type Dialect string

const (
	DialectSQLite   Dialect = "sqlite"
	DialectPostgres Dialect = "postgres"
)

// Runner handles database migrations
type Runner struct {
	db         *sql.DB
	dialect    Dialect
	migrations []Migration
}

// NewRunner creates a new migration runner
func NewRunner(db *sql.DB, dialect Dialect) *Runner {
	var migrations []Migration
	switch dialect {
	case DialectSQLite:
		migrations = SQLiteMigrations()
	case DialectPostgres:
		migrations = PostgresMigrations()
	}

	// Sort migrations by semantic version
	sort.Slice(migrations, func(i, j int) bool {
		return semver.Compare("v"+migrations[i].Version, "v"+migrations[j].Version) < 0
	})

	return &Runner{
		db:         db,
		dialect:    dialect,
		migrations: migrations,
	}
}

// Run executes all pending migrations
func (r *Runner) Run() error {
	// Step 1: Apply initial schema with existence checks (for v0.0.1-alpha compatibility)
	if err := r.runInitialSchema(); err != nil {
		return fmt.Errorf("failed to run initial schema: %w", err)
	}

	// Step 2: Apply versioned migrations from schema_migrations
	if err := r.runVersionedMigrations(); err != nil {
		return fmt.Errorf("failed to run versioned migrations: %w", err)
	}

	return nil
}

// runInitialSchema applies the initial schema with existence checks
// This handles both new databases and existing v0.0.1-alpha databases
func (r *Runner) runInitialSchema() error {
	// Apply initial schema (all statements use IF NOT EXISTS)
	var schema string
	switch r.dialect {
	case DialectSQLite:
		schema = SQLiteInitialSchema
	case DialectPostgres:
		schema = PostgresInitialSchema
	}

	if _, err := r.db.Exec(schema); err != nil {
		return fmt.Errorf("failed to run initial schema: %w", err)
	}

	// For existing databases: add message column if not exists
	// (new databases already have this column from initial schema)
	if err := r.ensureMessageColumn(); err != nil {
		return fmt.Errorf("failed to ensure message column: %w", err)
	}

	return nil
}

// ensureMessageColumn adds the message column to plan_document_events if it doesn't exist
// This is for backward compatibility with v0.0.1-alpha databases
func (r *Runner) ensureMessageColumn() error {
	var exists bool
	var checkQuery string

	switch r.dialect {
	case DialectSQLite:
		checkQuery = `SELECT COUNT(*) > 0 FROM pragma_table_info('plan_document_events') WHERE name='message'`
	case DialectPostgres:
		checkQuery = `SELECT EXISTS (SELECT 1 FROM information_schema.columns WHERE table_name = 'plan_document_events' AND column_name = 'message')`
	}

	if err := r.db.QueryRow(checkQuery).Scan(&exists); err != nil {
		return err
	}

	if !exists {
		var alterQuery string
		switch r.dialect {
		case DialectSQLite:
			alterQuery = `ALTER TABLE plan_document_events ADD COLUMN message TEXT NOT NULL DEFAULT ''`
		case DialectPostgres:
			alterQuery = `ALTER TABLE plan_document_events ADD COLUMN message TEXT NOT NULL DEFAULT ''`
		}
		if _, err := r.db.Exec(alterQuery); err != nil {
			return err
		}
	}

	return nil
}

// runVersionedMigrations applies migrations that haven't been run yet
func (r *Runner) runVersionedMigrations() error {
	if len(r.migrations) == 0 {
		return nil
	}

	appliedVersions, err := r.getAppliedVersions()
	if err != nil {
		return fmt.Errorf("failed to get applied versions: %w", err)
	}

	for _, m := range r.migrations {
		if appliedVersions[m.Version] {
			continue
		}

		if _, err := r.db.Exec(m.SQL); err != nil {
			return fmt.Errorf("failed to run migration %s: %w", m.Version, err)
		}

		if err := r.recordVersion(m.Version); err != nil {
			return fmt.Errorf("failed to record version %s: %w", m.Version, err)
		}
	}

	return nil
}

func (r *Runner) getAppliedVersions() (map[string]bool, error) {
	rows, err := r.db.Query(`SELECT version FROM schema_migrations`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	versions := make(map[string]bool)
	for rows.Next() {
		var version string
		if err := rows.Scan(&version); err != nil {
			return nil, err
		}
		versions[version] = true
	}

	return versions, rows.Err()
}

func (r *Runner) recordVersion(version string) error {
	var query string
	switch r.dialect {
	case DialectSQLite:
		query = `INSERT INTO schema_migrations (version) VALUES (?)`
	case DialectPostgres:
		query = `INSERT INTO schema_migrations (version) VALUES ($1)`
	}
	_, err := r.db.Exec(query, version)
	return err
}
