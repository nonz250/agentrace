package migrations

import (
	_ "embed"
)

//go:embed sqlite/001_initial.sql
var SQLiteInitial string

//go:embed sqlite/002_plan_documents.sql
var SQLitePlanDocuments string

//go:embed postgres/001_initial.up.sql
var PostgresInitial string

//go:embed postgres/002_plan_documents.up.sql
var PostgresPlanDocuments string
