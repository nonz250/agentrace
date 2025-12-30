package mongodb

import (
	"context"
	"time"

	"github.com/google/uuid"
	"github.com/satetsu888/agentrace/server/internal/domain"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
)

type OAuthConnectionRepository struct {
	collection *mongo.Collection
}

func NewOAuthConnectionRepository(db *DB) *OAuthConnectionRepository {
	return &OAuthConnectionRepository{
		collection: db.Collection("oauth_connections"),
	}
}

type oauthConnectionDocument struct {
	ID         string    `bson:"_id"`
	UserID     string    `bson:"user_id"`
	Provider   string    `bson:"provider"`
	ProviderID string    `bson:"provider_id"`
	CreatedAt  time.Time `bson:"created_at"`
}

func (r *OAuthConnectionRepository) Create(ctx context.Context, conn *domain.OAuthConnection) error {
	if conn.ID == "" {
		conn.ID = uuid.New().String()
	}
	if conn.CreatedAt.IsZero() {
		conn.CreatedAt = time.Now()
	}

	doc := oauthConnectionDocument{
		ID:         conn.ID,
		UserID:     conn.UserID,
		Provider:   conn.Provider,
		ProviderID: conn.ProviderID,
		CreatedAt:  conn.CreatedAt,
	}

	_, err := r.collection.InsertOne(ctx, doc)
	return err
}

func (r *OAuthConnectionRepository) FindByProviderAndProviderID(ctx context.Context, provider, providerID string) (*domain.OAuthConnection, error) {
	var doc oauthConnectionDocument
	err := r.collection.FindOne(ctx, bson.M{"provider": provider, "provider_id": providerID}).Decode(&doc)
	if err == mongo.ErrNoDocuments {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}

	return &domain.OAuthConnection{
		ID:         doc.ID,
		UserID:     doc.UserID,
		Provider:   doc.Provider,
		ProviderID: doc.ProviderID,
		CreatedAt:  doc.CreatedAt,
	}, nil
}

func (r *OAuthConnectionRepository) FindByUserID(ctx context.Context, userID string) ([]*domain.OAuthConnection, error) {
	cursor, err := r.collection.Find(ctx, bson.M{"user_id": userID})
	if err != nil {
		return nil, err
	}
	defer cursor.Close(ctx)

	var connections []*domain.OAuthConnection
	for cursor.Next(ctx) {
		var doc oauthConnectionDocument
		if err := cursor.Decode(&doc); err != nil {
			return nil, err
		}
		connections = append(connections, &domain.OAuthConnection{
			ID:         doc.ID,
			UserID:     doc.UserID,
			Provider:   doc.Provider,
			ProviderID: doc.ProviderID,
			CreatedAt:  doc.CreatedAt,
		})
	}

	return connections, cursor.Err()
}

func (r *OAuthConnectionRepository) Delete(ctx context.Context, id string) error {
	_, err := r.collection.DeleteOne(ctx, bson.M{"_id": id})
	return err
}
