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
			continue
		}

		log.Printf("Running DynamoDB migration %s: %s", m.Version, m.Description)

		if err := m.Up(ctx, r.db); err != nil {
			return fmt.Errorf("migration %s failed: %w", m.Version, err)
		}

		if err := r.recordVersion(ctx, m.Version); err != nil {
			return fmt.Errorf("failed to record version %s: %w", m.Version, err)
		}

		log.Printf("Completed DynamoDB migration %s", m.Version)
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

	log.Printf("Created DynamoDB table: %s", tableName)
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
		"applied_at": time.Now().Format(time.RFC3339),
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
		// Migrations will be added here as they are created
		// Example:
		// {
		//     Version:     "0.0.1",
		//     Description: "Add parent_session_id GSI for subagent support",
		//     Up:          migration_0_0_1_AddSubagentSupport,
		// },
	}
}
