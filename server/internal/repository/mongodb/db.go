package mongodb

import (
	"context"
	"fmt"
	"time"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

// DB wraps the MongoDB client and database
type DB struct {
	client   *mongo.Client
	database *mongo.Database
}

// Open opens a MongoDB connection and creates indexes
func Open(databaseURL string) (*DB, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	client, err := mongo.Connect(ctx, options.Client().ApplyURI(databaseURL))
	if err != nil {
		return nil, fmt.Errorf("failed to connect to MongoDB: %w", err)
	}

	// Ping to verify connection
	if err := client.Ping(ctx, nil); err != nil {
		return nil, fmt.Errorf("failed to ping MongoDB: %w", err)
	}

	// Extract database name from URL or use default
	dbName := "agentrace"
	if parsedDB := client.Database("agentrace"); parsedDB != nil {
		dbName = "agentrace"
	}

	db := &DB{
		client:   client,
		database: client.Database(dbName),
	}

	// Create indexes
	if err := db.createIndexes(ctx); err != nil {
		client.Disconnect(ctx)
		return nil, fmt.Errorf("failed to create indexes: %w", err)
	}

	return db, nil
}

func (db *DB) Close() error {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	return db.client.Disconnect(ctx)
}

func (db *DB) Collection(name string) *mongo.Collection {
	return db.database.Collection(name)
}

func (db *DB) createIndexes(ctx context.Context) error {
	// Users collection - no special indexes needed

	// API Keys collection
	apiKeysIndexes := []mongo.IndexModel{
		{
			Keys:    bson.D{{Key: "key_hash", Value: 1}},
			Options: options.Index().SetUnique(false),
		},
		{
			Keys: bson.D{{Key: "user_id", Value: 1}},
		},
	}
	if _, err := db.Collection("api_keys").Indexes().CreateMany(ctx, apiKeysIndexes); err != nil {
		return err
	}

	// Web Sessions collection
	webSessionsIndexes := []mongo.IndexModel{
		{
			Keys:    bson.D{{Key: "token", Value: 1}},
			Options: options.Index().SetUnique(true),
		},
		{
			Keys:    bson.D{{Key: "expires_at", Value: 1}},
			Options: options.Index().SetExpireAfterSeconds(0),
		},
	}
	if _, err := db.Collection("web_sessions").Indexes().CreateMany(ctx, webSessionsIndexes); err != nil {
		return err
	}

	// Sessions collection
	sessionsIndexes := []mongo.IndexModel{
		{
			Keys: bson.D{{Key: "claude_session_id", Value: 1}},
		},
		{
			Keys: bson.D{{Key: "user_id", Value: 1}},
		},
		{
			Keys: bson.D{{Key: "started_at", Value: -1}},
		},
	}
	if _, err := db.Collection("sessions").Indexes().CreateMany(ctx, sessionsIndexes); err != nil {
		return err
	}

	// Events collection
	eventsIndexes := []mongo.IndexModel{
		{
			Keys: bson.D{{Key: "session_id", Value: 1}},
		},
		{
			Keys: bson.D{{Key: "created_at", Value: 1}},
		},
	}
	if _, err := db.Collection("events").Indexes().CreateMany(ctx, eventsIndexes); err != nil {
		return err
	}

	return nil
}
