package mongodb

import (
	"context"
	"time"

	"github.com/google/uuid"
	"github.com/satetsu888/agentrace/server/internal/domain"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

type APIKeyRepository struct {
	collection *mongo.Collection
}

func NewAPIKeyRepository(db *DB) *APIKeyRepository {
	return &APIKeyRepository{
		collection: db.Collection("api_keys"),
	}
}

type apiKeyDocument struct {
	ID         string     `bson:"_id"`
	UserID     string     `bson:"user_id"`
	Name       string     `bson:"name"`
	KeyHash    string     `bson:"key_hash"`
	KeyPrefix  string     `bson:"key_prefix"`
	LastUsedAt *time.Time `bson:"last_used_at,omitempty"`
	CreatedAt  time.Time  `bson:"created_at"`
}

func (r *APIKeyRepository) Create(ctx context.Context, key *domain.APIKey) error {
	if key.ID == "" {
		key.ID = uuid.New().String()
	}
	if key.CreatedAt.IsZero() {
		key.CreatedAt = time.Now()
	}

	doc := apiKeyDocument{
		ID:         key.ID,
		UserID:     key.UserID,
		Name:       key.Name,
		KeyHash:    key.KeyHash,
		KeyPrefix:  key.KeyPrefix,
		LastUsedAt: key.LastUsedAt,
		CreatedAt:  key.CreatedAt,
	}

	_, err := r.collection.InsertOne(ctx, doc)
	return err
}

func (r *APIKeyRepository) FindByKeyHash(ctx context.Context, keyHash string) (*domain.APIKey, error) {
	var doc apiKeyDocument
	err := r.collection.FindOne(ctx, bson.M{"key_hash": keyHash}).Decode(&doc)
	if err == mongo.ErrNoDocuments {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}

	return docToAPIKey(&doc), nil
}

func (r *APIKeyRepository) FindByUserID(ctx context.Context, userID string) ([]*domain.APIKey, error) {
	opts := options.Find().SetSort(bson.D{{Key: "created_at", Value: -1}})
	cursor, err := r.collection.Find(ctx, bson.M{"user_id": userID}, opts)
	if err != nil {
		return nil, err
	}
	defer cursor.Close(ctx)

	var keys []*domain.APIKey
	for cursor.Next(ctx) {
		var doc apiKeyDocument
		if err := cursor.Decode(&doc); err != nil {
			return nil, err
		}
		keys = append(keys, docToAPIKey(&doc))
	}

	return keys, cursor.Err()
}

func (r *APIKeyRepository) FindByID(ctx context.Context, id string) (*domain.APIKey, error) {
	var doc apiKeyDocument
	err := r.collection.FindOne(ctx, bson.M{"_id": id}).Decode(&doc)
	if err == mongo.ErrNoDocuments {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}

	return docToAPIKey(&doc), nil
}

func (r *APIKeyRepository) Delete(ctx context.Context, id string) error {
	_, err := r.collection.DeleteOne(ctx, bson.M{"_id": id})
	return err
}

func (r *APIKeyRepository) UpdateLastUsedAt(ctx context.Context, id string) error {
	now := time.Now()
	_, err := r.collection.UpdateOne(ctx,
		bson.M{"_id": id},
		bson.M{"$set": bson.M{"last_used_at": now}},
	)
	return err
}

func docToAPIKey(doc *apiKeyDocument) *domain.APIKey {
	return &domain.APIKey{
		ID:         doc.ID,
		UserID:     doc.UserID,
		Name:       doc.Name,
		KeyHash:    doc.KeyHash,
		KeyPrefix:  doc.KeyPrefix,
		LastUsedAt: doc.LastUsedAt,
		CreatedAt:  doc.CreatedAt,
	}
}
