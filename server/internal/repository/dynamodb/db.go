package dynamodb

import (
	"context"
	"fmt"
	"log"
	"strings"
	"time"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/credentials"
	"github.com/aws/aws-sdk-go-v2/service/dynamodb"
	"github.com/aws/aws-sdk-go-v2/service/dynamodb/types"
)

// DB wraps the DynamoDB client
type DB struct {
	Client      *dynamodb.Client
	TablePrefix string
}

// DBConfig holds configuration for DynamoDB connection
type DBConfig struct {
	Endpoint    string // For local development (e.g., "http://localhost:8000")
	Region      string
	TablePrefix string
}

// ParseDatabaseURL parses a DynamoDB URL in the format:
// dynamodb://region/prefix or dynamodb://localhost:8000/prefix (for local)
func ParseDatabaseURL(url string) (*DBConfig, error) {
	if !strings.HasPrefix(url, "dynamodb://") {
		return nil, fmt.Errorf("invalid DynamoDB URL: must start with dynamodb://")
	}

	url = strings.TrimPrefix(url, "dynamodb://")
	parts := strings.SplitN(url, "/", 2)

	cfg := &DBConfig{
		Region:      "us-east-1",
		TablePrefix: "agentrace_",
	}

	if len(parts) >= 1 {
		hostOrRegion := parts[0]
		// Check if it's a local endpoint (contains port)
		if strings.Contains(hostOrRegion, ":") || hostOrRegion == "localhost" {
			cfg.Endpoint = "http://" + hostOrRegion
		} else {
			cfg.Region = hostOrRegion
		}
	}

	if len(parts) >= 2 && parts[1] != "" {
		cfg.TablePrefix = parts[1]
		if !strings.HasSuffix(cfg.TablePrefix, "_") {
			cfg.TablePrefix += "_"
		}
	}

	return cfg, nil
}

// Open creates a new DynamoDB client and ensures tables exist
func Open(databaseURL string) (*DB, error) {
	dbConfig, err := ParseDatabaseURL(databaseURL)
	if err != nil {
		return nil, err
	}

	ctx := context.Background()

	// Load AWS config
	var cfg aws.Config
	if dbConfig.Endpoint != "" {
		// Local DynamoDB
		cfg, err = config.LoadDefaultConfig(ctx,
			config.WithRegion(dbConfig.Region),
			config.WithCredentialsProvider(credentials.NewStaticCredentialsProvider("dummy", "dummy", "")),
		)
	} else {
		cfg, err = config.LoadDefaultConfig(ctx,
			config.WithRegion(dbConfig.Region),
		)
	}
	if err != nil {
		return nil, fmt.Errorf("failed to load AWS config: %w", err)
	}

	// Create DynamoDB client
	var client *dynamodb.Client
	if dbConfig.Endpoint != "" {
		client = dynamodb.NewFromConfig(cfg, func(o *dynamodb.Options) {
			o.BaseEndpoint = aws.String(dbConfig.Endpoint)
		})
	} else {
		client = dynamodb.NewFromConfig(cfg)
	}

	db := &DB{
		Client:      client,
		TablePrefix: dbConfig.TablePrefix,
	}

	// Ensure tables exist
	if err := db.ensureTables(ctx); err != nil {
		return nil, fmt.Errorf("failed to ensure tables: %w", err)
	}

	return db, nil
}

// Close closes the DynamoDB client (no-op for DynamoDB)
func (db *DB) Close() error {
	return nil
}

// TableName returns the full table name with prefix
func (db *DB) TableName(name string) string {
	return db.TablePrefix + name
}

// ensureTables creates all required tables if they don't exist
func (db *DB) ensureTables(ctx context.Context) error {
	tables := []tableDefinition{
		db.projectsTable(),
		db.sessionsTable(),
		db.eventsTable(),
		db.usersTable(),
		db.apiKeysTable(),
		db.webSessionsTable(),
		db.passwordCredentialsTable(),
		db.oauthConnectionsTable(),
		db.planDocumentsTable(),
		db.planDocumentEventsTable(),
		db.userFavoritesTable(),
	}

	for _, table := range tables {
		if err := db.createTableIfNotExists(ctx, table); err != nil {
			return fmt.Errorf("failed to create table %s: %w", table.name, err)
		}
	}

	// Create default project if it doesn't exist
	if err := db.ensureDefaultProject(ctx); err != nil {
		return fmt.Errorf("failed to ensure default project: %w", err)
	}

	return nil
}

const projectGSIPK = "PROJECT"

func (db *DB) ensureDefaultProject(ctx context.Context) error {
	defaultProjectID := "00000000-0000-0000-0000-000000000000"

	// Check if default project exists
	result, err := db.Client.GetItem(ctx, &dynamodb.GetItemInput{
		TableName: aws.String(db.TableName("projects")),
		Key: map[string]types.AttributeValue{
			"id": &types.AttributeValueMemberS{Value: defaultProjectID},
		},
	})
	if err != nil {
		return err
	}
	if result.Item != nil {
		return nil // Already exists
	}

	// Create default project
	// Note: We omit canonical_git_repository for the default project to avoid
	// empty string in GSI (DynamoDB doesn't allow empty strings in GSI key attributes)
	_, err = db.Client.PutItem(ctx, &dynamodb.PutItemInput{
		TableName: aws.String(db.TableName("projects")),
		Item: map[string]types.AttributeValue{
			"id":         &types.AttributeValueMemberS{Value: defaultProjectID},
			"_gsi_pk":    &types.AttributeValueMemberS{Value: projectGSIPK},
			"created_at": &types.AttributeValueMemberS{Value: time.Now().Format(time.RFC3339Nano)},
		},
	})
	return err
}

type tableDefinition struct {
	name                   string
	keySchema              []types.KeySchemaElement
	attributeDefinitions   []types.AttributeDefinition
	globalSecondaryIndexes []types.GlobalSecondaryIndex
}

func (db *DB) createTableIfNotExists(ctx context.Context, table tableDefinition) error {
	tableName := db.TableName(table.name)

	// Check if table exists
	_, err := db.Client.DescribeTable(ctx, &dynamodb.DescribeTableInput{
		TableName: aws.String(tableName),
	})
	if err == nil {
		return nil // Table exists
	}

	// Create table
	input := &dynamodb.CreateTableInput{
		TableName:            aws.String(tableName),
		KeySchema:            table.keySchema,
		AttributeDefinitions: table.attributeDefinitions,
		BillingMode:          types.BillingModePayPerRequest,
	}

	if len(table.globalSecondaryIndexes) > 0 {
		input.GlobalSecondaryIndexes = table.globalSecondaryIndexes
	}

	_, err = db.Client.CreateTable(ctx, input)
	if err != nil {
		return err
	}

	// Wait for table to be active
	waiter := dynamodb.NewTableExistsWaiter(db.Client)
	err = waiter.Wait(ctx, &dynamodb.DescribeTableInput{
		TableName: aws.String(tableName),
	}, 2*time.Minute)
	if err != nil {
		return fmt.Errorf("failed waiting for table %s: %w", tableName, err)
	}

	log.Printf("Created DynamoDB table: %s", tableName)
	return nil
}

// Table definitions

func (db *DB) projectsTable() tableDefinition {
	return tableDefinition{
		name: "projects",
		keySchema: []types.KeySchemaElement{
			{AttributeName: aws.String("id"), KeyType: types.KeyTypeHash},
		},
		attributeDefinitions: []types.AttributeDefinition{
			{AttributeName: aws.String("id"), AttributeType: types.ScalarAttributeTypeS},
			{AttributeName: aws.String("canonical_git_repository"), AttributeType: types.ScalarAttributeTypeS},
			{AttributeName: aws.String("created_at"), AttributeType: types.ScalarAttributeTypeS},
			{AttributeName: aws.String("_gsi_pk"), AttributeType: types.ScalarAttributeTypeS},
		},
		globalSecondaryIndexes: []types.GlobalSecondaryIndex{
			{
				IndexName: aws.String("canonical_git_repository-index"),
				KeySchema: []types.KeySchemaElement{
					{AttributeName: aws.String("canonical_git_repository"), KeyType: types.KeyTypeHash},
				},
				Projection: &types.Projection{ProjectionType: types.ProjectionTypeAll},
			},
			{
				IndexName: aws.String("gsi-created_at-index"),
				KeySchema: []types.KeySchemaElement{
					{AttributeName: aws.String("_gsi_pk"), KeyType: types.KeyTypeHash},
					{AttributeName: aws.String("created_at"), KeyType: types.KeyTypeRange},
				},
				Projection: &types.Projection{ProjectionType: types.ProjectionTypeAll},
			},
		},
	}
}

func (db *DB) sessionsTable() tableDefinition {
	return tableDefinition{
		name: "sessions",
		keySchema: []types.KeySchemaElement{
			{AttributeName: aws.String("id"), KeyType: types.KeyTypeHash},
		},
		attributeDefinitions: []types.AttributeDefinition{
			{AttributeName: aws.String("id"), AttributeType: types.ScalarAttributeTypeS},
			{AttributeName: aws.String("claude_session_id"), AttributeType: types.ScalarAttributeTypeS},
			{AttributeName: aws.String("project_id"), AttributeType: types.ScalarAttributeTypeS},
			{AttributeName: aws.String("updated_at"), AttributeType: types.ScalarAttributeTypeS},
			{AttributeName: aws.String("created_at"), AttributeType: types.ScalarAttributeTypeS},
			{AttributeName: aws.String("_gsi_pk"), AttributeType: types.ScalarAttributeTypeS},
		},
		globalSecondaryIndexes: []types.GlobalSecondaryIndex{
			{
				IndexName: aws.String("claude_session_id-index"),
				KeySchema: []types.KeySchemaElement{
					{AttributeName: aws.String("claude_session_id"), KeyType: types.KeyTypeHash},
				},
				Projection: &types.Projection{ProjectionType: types.ProjectionTypeAll},
			},
			{
				IndexName: aws.String("project_id-updated_at-index"),
				KeySchema: []types.KeySchemaElement{
					{AttributeName: aws.String("project_id"), KeyType: types.KeyTypeHash},
					{AttributeName: aws.String("updated_at"), KeyType: types.KeyTypeRange},
				},
				Projection: &types.Projection{ProjectionType: types.ProjectionTypeAll},
			},
			{
				IndexName: aws.String("project_id-created_at-index"),
				KeySchema: []types.KeySchemaElement{
					{AttributeName: aws.String("project_id"), KeyType: types.KeyTypeHash},
					{AttributeName: aws.String("created_at"), KeyType: types.KeyTypeRange},
				},
				Projection: &types.Projection{ProjectionType: types.ProjectionTypeAll},
			},
			{
				IndexName: aws.String("gsi-updated_at-index"),
				KeySchema: []types.KeySchemaElement{
					{AttributeName: aws.String("_gsi_pk"), KeyType: types.KeyTypeHash},
					{AttributeName: aws.String("updated_at"), KeyType: types.KeyTypeRange},
				},
				Projection: &types.Projection{ProjectionType: types.ProjectionTypeAll},
			},
			{
				IndexName: aws.String("gsi-created_at-index"),
				KeySchema: []types.KeySchemaElement{
					{AttributeName: aws.String("_gsi_pk"), KeyType: types.KeyTypeHash},
					{AttributeName: aws.String("created_at"), KeyType: types.KeyTypeRange},
				},
				Projection: &types.Projection{ProjectionType: types.ProjectionTypeAll},
			},
		},
	}
}

func (db *DB) eventsTable() tableDefinition {
	return tableDefinition{
		name: "events",
		keySchema: []types.KeySchemaElement{
			{AttributeName: aws.String("session_id"), KeyType: types.KeyTypeHash},
			{AttributeName: aws.String("sort_key"), KeyType: types.KeyTypeRange}, // created_at#id for chronological ordering
		},
		attributeDefinitions: []types.AttributeDefinition{
			{AttributeName: aws.String("session_id"), AttributeType: types.ScalarAttributeTypeS},
			{AttributeName: aws.String("sort_key"), AttributeType: types.ScalarAttributeTypeS},
		},
	}
}

func (db *DB) usersTable() tableDefinition {
	return tableDefinition{
		name: "users",
		keySchema: []types.KeySchemaElement{
			{AttributeName: aws.String("id"), KeyType: types.KeyTypeHash},
		},
		attributeDefinitions: []types.AttributeDefinition{
			{AttributeName: aws.String("id"), AttributeType: types.ScalarAttributeTypeS},
			{AttributeName: aws.String("email"), AttributeType: types.ScalarAttributeTypeS},
		},
		globalSecondaryIndexes: []types.GlobalSecondaryIndex{
			{
				IndexName: aws.String("email-index"),
				KeySchema: []types.KeySchemaElement{
					{AttributeName: aws.String("email"), KeyType: types.KeyTypeHash},
				},
				Projection: &types.Projection{ProjectionType: types.ProjectionTypeAll},
			},
		},
	}
}

func (db *DB) apiKeysTable() tableDefinition {
	return tableDefinition{
		name: "api_keys",
		keySchema: []types.KeySchemaElement{
			{AttributeName: aws.String("id"), KeyType: types.KeyTypeHash},
		},
		attributeDefinitions: []types.AttributeDefinition{
			{AttributeName: aws.String("id"), AttributeType: types.ScalarAttributeTypeS},
			{AttributeName: aws.String("user_id"), AttributeType: types.ScalarAttributeTypeS},
			{AttributeName: aws.String("key_hash"), AttributeType: types.ScalarAttributeTypeS},
		},
		globalSecondaryIndexes: []types.GlobalSecondaryIndex{
			{
				IndexName: aws.String("user_id-index"),
				KeySchema: []types.KeySchemaElement{
					{AttributeName: aws.String("user_id"), KeyType: types.KeyTypeHash},
				},
				Projection: &types.Projection{ProjectionType: types.ProjectionTypeAll},
			},
			{
				IndexName: aws.String("key_hash-index"),
				KeySchema: []types.KeySchemaElement{
					{AttributeName: aws.String("key_hash"), KeyType: types.KeyTypeHash},
				},
				Projection: &types.Projection{ProjectionType: types.ProjectionTypeAll},
			},
		},
	}
}

func (db *DB) webSessionsTable() tableDefinition {
	return tableDefinition{
		name: "web_sessions",
		keySchema: []types.KeySchemaElement{
			{AttributeName: aws.String("id"), KeyType: types.KeyTypeHash},
		},
		attributeDefinitions: []types.AttributeDefinition{
			{AttributeName: aws.String("id"), AttributeType: types.ScalarAttributeTypeS},
			{AttributeName: aws.String("token"), AttributeType: types.ScalarAttributeTypeS},
		},
		globalSecondaryIndexes: []types.GlobalSecondaryIndex{
			{
				IndexName: aws.String("token-index"),
				KeySchema: []types.KeySchemaElement{
					{AttributeName: aws.String("token"), KeyType: types.KeyTypeHash},
				},
				Projection: &types.Projection{ProjectionType: types.ProjectionTypeAll},
			},
		},
	}
}

func (db *DB) passwordCredentialsTable() tableDefinition {
	return tableDefinition{
		name: "password_credentials",
		keySchema: []types.KeySchemaElement{
			{AttributeName: aws.String("id"), KeyType: types.KeyTypeHash},
		},
		attributeDefinitions: []types.AttributeDefinition{
			{AttributeName: aws.String("id"), AttributeType: types.ScalarAttributeTypeS},
			{AttributeName: aws.String("user_id"), AttributeType: types.ScalarAttributeTypeS},
		},
		globalSecondaryIndexes: []types.GlobalSecondaryIndex{
			{
				IndexName: aws.String("user_id-index"),
				KeySchema: []types.KeySchemaElement{
					{AttributeName: aws.String("user_id"), KeyType: types.KeyTypeHash},
				},
				Projection: &types.Projection{ProjectionType: types.ProjectionTypeAll},
			},
		},
	}
}

func (db *DB) oauthConnectionsTable() tableDefinition {
	return tableDefinition{
		name: "oauth_connections",
		keySchema: []types.KeySchemaElement{
			{AttributeName: aws.String("id"), KeyType: types.KeyTypeHash},
		},
		attributeDefinitions: []types.AttributeDefinition{
			{AttributeName: aws.String("id"), AttributeType: types.ScalarAttributeTypeS},
			{AttributeName: aws.String("user_id"), AttributeType: types.ScalarAttributeTypeS},
			{AttributeName: aws.String("provider_provider_id"), AttributeType: types.ScalarAttributeTypeS},
		},
		globalSecondaryIndexes: []types.GlobalSecondaryIndex{
			{
				IndexName: aws.String("user_id-index"),
				KeySchema: []types.KeySchemaElement{
					{AttributeName: aws.String("user_id"), KeyType: types.KeyTypeHash},
				},
				Projection: &types.Projection{ProjectionType: types.ProjectionTypeAll},
			},
			{
				IndexName: aws.String("provider_provider_id-index"),
				KeySchema: []types.KeySchemaElement{
					{AttributeName: aws.String("provider_provider_id"), KeyType: types.KeyTypeHash},
				},
				Projection: &types.Projection{ProjectionType: types.ProjectionTypeAll},
			},
		},
	}
}

func (db *DB) planDocumentsTable() tableDefinition {
	return tableDefinition{
		name: "plan_documents",
		keySchema: []types.KeySchemaElement{
			{AttributeName: aws.String("id"), KeyType: types.KeyTypeHash},
		},
		attributeDefinitions: []types.AttributeDefinition{
			{AttributeName: aws.String("id"), AttributeType: types.ScalarAttributeTypeS},
			{AttributeName: aws.String("project_id"), AttributeType: types.ScalarAttributeTypeS},
			{AttributeName: aws.String("updated_at"), AttributeType: types.ScalarAttributeTypeS},
			{AttributeName: aws.String("created_at"), AttributeType: types.ScalarAttributeTypeS},
			{AttributeName: aws.String("_gsi_pk"), AttributeType: types.ScalarAttributeTypeS},
		},
		globalSecondaryIndexes: []types.GlobalSecondaryIndex{
			{
				IndexName: aws.String("project_id-updated_at-index"),
				KeySchema: []types.KeySchemaElement{
					{AttributeName: aws.String("project_id"), KeyType: types.KeyTypeHash},
					{AttributeName: aws.String("updated_at"), KeyType: types.KeyTypeRange},
				},
				Projection: &types.Projection{ProjectionType: types.ProjectionTypeAll},
			},
			{
				IndexName: aws.String("project_id-created_at-index"),
				KeySchema: []types.KeySchemaElement{
					{AttributeName: aws.String("project_id"), KeyType: types.KeyTypeHash},
					{AttributeName: aws.String("created_at"), KeyType: types.KeyTypeRange},
				},
				Projection: &types.Projection{ProjectionType: types.ProjectionTypeAll},
			},
			{
				IndexName: aws.String("gsi-updated_at-index"),
				KeySchema: []types.KeySchemaElement{
					{AttributeName: aws.String("_gsi_pk"), KeyType: types.KeyTypeHash},
					{AttributeName: aws.String("updated_at"), KeyType: types.KeyTypeRange},
				},
				Projection: &types.Projection{ProjectionType: types.ProjectionTypeAll},
			},
			{
				IndexName: aws.String("gsi-created_at-index"),
				KeySchema: []types.KeySchemaElement{
					{AttributeName: aws.String("_gsi_pk"), KeyType: types.KeyTypeHash},
					{AttributeName: aws.String("created_at"), KeyType: types.KeyTypeRange},
				},
				Projection: &types.Projection{ProjectionType: types.ProjectionTypeAll},
			},
		},
	}
}

func (db *DB) planDocumentEventsTable() tableDefinition {
	return tableDefinition{
		name: "plan_document_events",
		keySchema: []types.KeySchemaElement{
			{AttributeName: aws.String("plan_document_id"), KeyType: types.KeyTypeHash},
			{AttributeName: aws.String("sort_key"), KeyType: types.KeyTypeRange}, // created_at#id for chronological ordering
		},
		attributeDefinitions: []types.AttributeDefinition{
			{AttributeName: aws.String("plan_document_id"), AttributeType: types.ScalarAttributeTypeS},
			{AttributeName: aws.String("sort_key"), AttributeType: types.ScalarAttributeTypeS},
			{AttributeName: aws.String("claude_session_id"), AttributeType: types.ScalarAttributeTypeS},
			{AttributeName: aws.String("user_id"), AttributeType: types.ScalarAttributeTypeS},
		},
		globalSecondaryIndexes: []types.GlobalSecondaryIndex{
			{
				IndexName: aws.String("claude_session_id-index"),
				KeySchema: []types.KeySchemaElement{
					{AttributeName: aws.String("claude_session_id"), KeyType: types.KeyTypeHash},
				},
				Projection: &types.Projection{ProjectionType: types.ProjectionTypeAll},
			},
			{
				IndexName: aws.String("user_id-index"),
				KeySchema: []types.KeySchemaElement{
					{AttributeName: aws.String("user_id"), KeyType: types.KeyTypeHash},
				},
				Projection: &types.Projection{ProjectionType: types.ProjectionTypeAll},
			},
		},
	}
}

func (db *DB) userFavoritesTable() tableDefinition {
	return tableDefinition{
		name: "user_favorites",
		keySchema: []types.KeySchemaElement{
			{AttributeName: aws.String("id"), KeyType: types.KeyTypeHash},
		},
		attributeDefinitions: []types.AttributeDefinition{
			{AttributeName: aws.String("id"), AttributeType: types.ScalarAttributeTypeS},
			{AttributeName: aws.String("user_id"), AttributeType: types.ScalarAttributeTypeS},
			{AttributeName: aws.String("user_target_type_target_id"), AttributeType: types.ScalarAttributeTypeS},
		},
		globalSecondaryIndexes: []types.GlobalSecondaryIndex{
			{
				IndexName: aws.String("user_id-index"),
				KeySchema: []types.KeySchemaElement{
					{AttributeName: aws.String("user_id"), KeyType: types.KeyTypeHash},
				},
				Projection: &types.Projection{ProjectionType: types.ProjectionTypeAll},
			},
			{
				IndexName: aws.String("user_target_type_target_id-index"),
				KeySchema: []types.KeySchemaElement{
					{AttributeName: aws.String("user_target_type_target_id"), KeyType: types.KeyTypeHash},
				},
				Projection: &types.Projection{ProjectionType: types.ProjectionTypeAll},
			},
		},
	}
}
