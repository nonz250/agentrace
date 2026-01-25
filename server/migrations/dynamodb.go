package migrations

// DynamoDBMigration represents a single DynamoDB migration
// Unlike SQL migrations, DynamoDB migrations are Go functions because:
// - GSI additions cannot be expressed in SQL
// - Waiting for GSI to become ACTIVE is required
// - Table existence checks and conditional logic are needed
// - Error handling needs to be flexible
type DynamoDBMigration struct {
	Version     string // Semantic version (e.g., "0.0.1", "0.1.0")
	Description string
}

// DynamoDBMigrations returns the list of DynamoDB migrations to be applied.
// The actual migration functions are defined in the dynamodb package
// to avoid circular dependencies.
// This is a registry of versions for documentation purposes.
func DynamoDBMigrations() []DynamoDBMigration {
	return []DynamoDBMigration{
		{Version: "0.0.1", Description: "Add parent_session_id GSI for subagent support"},
		{Version: "0.1.0", Description: "Add plan_comment_threads and plan_comment_messages tables (handled in ensureTables)"},
	}
}
