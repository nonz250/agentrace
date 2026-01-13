package migrations

import (
	_ "embed"
)

// Initial schema (v0.0.1-alpha)
// Applied with existence checks for backward compatibility with existing databases

//go:embed sqlite/initial.sql
var SQLiteInitialSchema string

//go:embed postgres/initial.up.sql
var PostgresInitialSchema string

// Migration represents a single versioned migration
type Migration struct {
	Version string // Semantic version (e.g., "0.0.1", "0.1.0")
	SQL     string
}

// SQLiteMigrations returns all SQLite migrations with semantic versions
// Add new migrations here as they are created
func SQLiteMigrations() []Migration {
	return []Migration{
		// Example: {Version: "0.0.1", SQL: SQLiteMigration_0_0_1},
	}
}

// PostgresMigrations returns all PostgreSQL migrations with semantic versions
// Add new migrations here as they are created
func PostgresMigrations() []Migration {
	return []Migration{
		// Example: {Version: "0.0.1", SQL: PostgresMigration_0_0_1},
	}
}
