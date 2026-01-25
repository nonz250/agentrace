package dynamodb

import (
	"context"
	"fmt"
	"log"
	"sort"
	"time"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/feature/dynamodb/attributevalue"
	"github.com/aws/aws-sdk-go-v2/service/dynamodb"
	"github.com/aws/aws-sdk-go-v2/service/dynamodb/types"
	"golang.org/x/mod/semver"
)

// Migration represents a single DynamoDB migration
type Migration struct {
	Version     string
	Description string
	Up          func(ctx context.Context, db *DB) error
}

// MigrationRunner handles DynamoDB migrations
type MigrationRunner struct {
	db         *DB
	migrations []Migration
}

// NewMigrationRunner creates a new DynamoDB migration runner
func NewMigrationRunner(db *DB) *MigrationRunner {
	migrations := registeredMigrations()

	// Sort by semantic version
	sort.Slice(migrations, func(i, j int) bool {
		return semver.Compare("v"+migrations[i].Version, "v"+migrations[j].Version) < 0
	})

	return &MigrationRunner{
		db:         db,
		migrations: migrations,
	}
}

// Run executes all pending migrations
func (r *MigrationRunner) Run(ctx context.Context) error {
	// Ensure schema_migrations table exists
	if err := r.ensureMigrationsTable(ctx); err != nil {
		return fmt.Errorf("failed to ensure migrations table: %w", err)
	}

	// Get applied versions
	appliedVersions, err := r.getAppliedVersions(ctx)
	if err != nil {
		return fmt.Errorf("failed to get applied versions: %w", err)
	}

	// Run pending migrations
	for _, m := range r.migrations {
		if appliedVersions[m.Version] {
			log.Printf("[migration] Skipping dynamodb migration v%s (already applied)", m.Version)
			continue
		}

		log.Printf("[migration] Running dynamodb migration v%s: %s", m.Version, m.Description)

		if err := m.Up(ctx, r.db); err != nil {
			return fmt.Errorf("migration %s failed: %w", m.Version, err)
		}

		if err := r.recordVersion(ctx, m.Version); err != nil {
			return fmt.Errorf("failed to record version %s: %w", m.Version, err)
		}

		log.Printf("[migration] Completed dynamodb migration v%s", m.Version)
	}

	return nil
}

func (r *MigrationRunner) ensureMigrationsTable(ctx context.Context) error {
	tableName := r.db.TableName("schema_migrations")

	// Check if table exists
	_, err := r.db.Client.DescribeTable(ctx, &dynamodb.DescribeTableInput{
		TableName: aws.String(tableName),
	})
	if err == nil {
		return nil // Table exists
	}

	// Create table
	_, err = r.db.Client.CreateTable(ctx, &dynamodb.CreateTableInput{
		TableName: aws.String(tableName),
		KeySchema: []types.KeySchemaElement{
			{AttributeName: aws.String("version"), KeyType: types.KeyTypeHash},
		},
		AttributeDefinitions: []types.AttributeDefinition{
			{AttributeName: aws.String("version"), AttributeType: types.ScalarAttributeTypeS},
		},
		BillingMode: types.BillingModePayPerRequest,
	})
	if err != nil {
		return err
	}

	// Wait for table to be active
	waiter := dynamodb.NewTableExistsWaiter(r.db.Client)
	err = waiter.Wait(ctx, &dynamodb.DescribeTableInput{
		TableName: aws.String(tableName),
	}, 2*time.Minute)
	if err != nil {
		return fmt.Errorf("failed waiting for schema_migrations table: %w", err)
	}

	log.Printf("[migration] Created dynamodb table: %s", tableName)
	return nil
}

func (r *MigrationRunner) getAppliedVersions(ctx context.Context) (map[string]bool, error) {
	tableName := r.db.TableName("schema_migrations")

	result, err := r.db.Client.Scan(ctx, &dynamodb.ScanInput{
		TableName: aws.String(tableName),
	})
	if err != nil {
		return nil, err
	}

	versions := make(map[string]bool)
	for _, item := range result.Items {
		var record struct {
			Version string `dynamodbav:"version"`
		}
		if err := attributevalue.UnmarshalMap(item, &record); err != nil {
			return nil, err
		}
		versions[record.Version] = true
	}

	return versions, nil
}

func (r *MigrationRunner) recordVersion(ctx context.Context, version string) error {
	tableName := r.db.TableName("schema_migrations")

	item, err := attributevalue.MarshalMap(map[string]interface{}{
		"version":    version,
		"applied_at": time.Now().Format(time.RFC3339Nano),
	})
	if err != nil {
		return err
	}

	_, err = r.db.Client.PutItem(ctx, &dynamodb.PutItemInput{
		TableName: aws.String(tableName),
		Item:      item,
	})
	return err
}

// registeredMigrations returns all registered DynamoDB migrations.
// Add new migrations here as they are created.
func registeredMigrations() []Migration {
	return []Migration{
		{
			Version:     "0.0.1",
			Description: "v0.0.1 schema changes",
			Up:          migration_0_0_1,
		},
		{
			Version:     "0.1.0",
			Description: "Add plan comment tables (tables created in ensureTables)",
			Up:          migration_0_1_0,
		},
	}
}

// migration_0_1_0 applies v0.1.0 schema changes
// Note: plan_comment_threads and plan_comment_messages tables are already
// created in ensureTables(), so this migration is a no-op for consistency.
func migration_0_1_0(ctx context.Context, db *DB) error {
	// Tables are created in ensureTables() with createTableIfNotExists
	// This migration exists for version tracking consistency with SQLite/PostgreSQL
	return nil
}

// migration_0_0_1 applies v0.0.1 schema changes
func migration_0_0_1(ctx context.Context, db *DB) error {
	// サブエージェント関連: parent_session_id GSI を追加
	if err := addParentSessionIDGSI(ctx, db); err != nil {
		return err
	}

	return nil
}

// addParentSessionIDGSI adds a GSI for querying subagents by parent session ID
func addParentSessionIDGSI(ctx context.Context, db *DB) error {
	tableName := db.TableName("sessions")

	// 1. Get existing attribute definitions
	desc, err := db.Client.DescribeTable(ctx, &dynamodb.DescribeTableInput{
		TableName: aws.String(tableName),
	})
	if err != nil {
		return fmt.Errorf("failed to describe table: %w", err)
	}

	// 2. Check if GSI already exists
	for _, gsi := range desc.Table.GlobalSecondaryIndexes {
		if *gsi.IndexName == "parent_session_id-created_at-index" {
			log.Printf("[migration] Skipping GSI parent_session_id-created_at-index (already exists)")
			return nil
		}
	}

	// 3. Build attribute definitions (existing + new)
	attrDefs := desc.Table.AttributeDefinitions

	// Check if parent_session_id is already defined
	hasParentSessionID := false
	for _, attr := range attrDefs {
		if *attr.AttributeName == "parent_session_id" {
			hasParentSessionID = true
			break
		}
	}

	if !hasParentSessionID {
		attrDefs = append(attrDefs, types.AttributeDefinition{
			AttributeName: aws.String("parent_session_id"),
			AttributeType: types.ScalarAttributeTypeS,
		})
	}

	// 4. UpdateTable to add GSI
	_, err = db.Client.UpdateTable(ctx, &dynamodb.UpdateTableInput{
		TableName:            aws.String(tableName),
		AttributeDefinitions: attrDefs,
		GlobalSecondaryIndexUpdates: []types.GlobalSecondaryIndexUpdate{
			{
				Create: &types.CreateGlobalSecondaryIndexAction{
					IndexName: aws.String("parent_session_id-created_at-index"),
					KeySchema: []types.KeySchemaElement{
						{AttributeName: aws.String("parent_session_id"), KeyType: types.KeyTypeHash},
						{AttributeName: aws.String("created_at"), KeyType: types.KeyTypeRange},
					},
					Projection: &types.Projection{ProjectionType: types.ProjectionTypeAll},
				},
			},
		},
	})
	if err != nil {
		return fmt.Errorf("failed to update table: %w", err)
	}

	// 5. Wait for GSI to become ACTIVE
	if err := db.WaitForGSIActive(ctx, tableName, "parent_session_id-created_at-index"); err != nil {
		return fmt.Errorf("failed waiting for GSI: %w", err)
	}

	log.Printf("[migration] Added GSI parent_session_id-created_at-index")
	return nil
}
