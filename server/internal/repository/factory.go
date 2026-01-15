package repository

import (
	"fmt"
)

// DBType represents the type of database to use
type DBType string

const (
	DBTypeMemory   DBType = "memory"
	DBTypeSQLite   DBType = "sqlite"
	DBTypePostgres DBType = "postgres"
	DBTypeTurso    DBType = "turso"
	DBTypeDynamoDB DBType = "dynamodb"
)

// RepositoryFactory creates repositories based on the database type
type RepositoryFactory interface {
	Create() (*Repositories, error)
	Close() error
}

// NewRepositoryFactory creates a new repository factory based on the database type
func NewRepositoryFactory(dbType string, databaseURL string) (RepositoryFactory, error) {
	switch DBType(dbType) {
	case DBTypeMemory:
		return NewMemoryFactory(), nil
	case DBTypeSQLite:
		return NewSQLiteFactory(databaseURL)
	case DBTypePostgres:
		return NewPostgresFactory(databaseURL)
	case DBTypeTurso:
		return NewTursoFactory(databaseURL)
	case DBTypeDynamoDB:
		return NewDynamoDBFactory(databaseURL)
	default:
		return nil, fmt.Errorf("unknown database type: %s", dbType)
	}
}

// MemoryFactory creates in-memory repositories
type MemoryFactory struct{}

func NewMemoryFactory() *MemoryFactory {
	return &MemoryFactory{}
}

func (f *MemoryFactory) Create() (*Repositories, error) {
	// This will be implemented by importing memory package
	// For now, return nil - will be wired up in main.go
	return nil, fmt.Errorf("memory factory not wired up - use memory.NewRepositories() directly")
}

func (f *MemoryFactory) Close() error {
	return nil
}

// SQLiteFactory creates SQLite repositories
type SQLiteFactory struct {
	databaseURL string
	db          interface{} // Will be *sql.DB
}

func NewSQLiteFactory(databaseURL string) (*SQLiteFactory, error) {
	if databaseURL == "" {
		return nil, fmt.Errorf("DATABASE_URL is required for sqlite")
	}
	return &SQLiteFactory{databaseURL: databaseURL}, nil
}

func (f *SQLiteFactory) Create() (*Repositories, error) {
	// Will be implemented in sqlite package
	return nil, fmt.Errorf("sqlite factory not implemented yet")
}

func (f *SQLiteFactory) Close() error {
	return nil
}

// PostgresFactory creates PostgreSQL repositories
type PostgresFactory struct {
	databaseURL string
	db          interface{} // Will be *sql.DB
}

func NewPostgresFactory(databaseURL string) (*PostgresFactory, error) {
	if databaseURL == "" {
		return nil, fmt.Errorf("DATABASE_URL is required for postgres")
	}
	return &PostgresFactory{databaseURL: databaseURL}, nil
}

func (f *PostgresFactory) Create() (*Repositories, error) {
	// Will be implemented in postgres package
	return nil, fmt.Errorf("postgres factory not implemented yet")
}

func (f *PostgresFactory) Close() error {
	return nil
}

// DynamoDBFactory creates DynamoDB repositories
type DynamoDBFactory struct {
	endpoint string // Optional endpoint URL for local development
}

func NewDynamoDBFactory(endpoint string) (*DynamoDBFactory, error) {
	// endpoint is optional (empty string means use AWS defaults)
	return &DynamoDBFactory{endpoint: endpoint}, nil
}

func (f *DynamoDBFactory) Create() (*Repositories, error) {
	// Will be implemented in dynamodb package
	return nil, fmt.Errorf("dynamodb factory not implemented yet")
}

func (f *DynamoDBFactory) Close() error {
	return nil
}

// TursoFactory creates Turso repositories
type TursoFactory struct {
	databaseURL string
	db          interface{} // Will be *sql.DB
}

func NewTursoFactory(databaseURL string) (*TursoFactory, error) {
	if databaseURL == "" {
		return nil, fmt.Errorf("DATABASE_URL is required for turso")
	}
	return &TursoFactory{databaseURL: databaseURL}, nil
}

func (f *TursoFactory) Create() (*Repositories, error) {
	// Will be implemented in turso package
	return nil, fmt.Errorf("turso factory not implemented yet")
}

func (f *TursoFactory) Close() error {
	return nil
}
