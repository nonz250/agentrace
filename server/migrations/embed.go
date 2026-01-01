package migrations

import (
	_ "embed"
)

//go:embed sqlite/001_initial.sql
var SQLiteSchema string

//go:embed postgres/001_initial.up.sql
var PostgresSchema string
