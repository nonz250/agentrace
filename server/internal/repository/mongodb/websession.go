package mongodb

import (
	"context"
	"time"

	"github.com/google/uuid"
	"github.com/satetsu888/agentrace/server/internal/domain"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
)

type WebSessionRepository struct {
	collection *mongo.Collection
}

func NewWebSessionRepository(db *DB) *WebSessionRepository {
	return &WebSessionRepository{
		collection: db.Collection("web_sessions"),
	}
}

type webSessionDocument struct {
	ID        string    `bson:"_id"`
	UserID    string    `bson:"user_id"`
	Token     string    `bson:"token"`
	ExpiresAt time.Time `bson:"expires_at"`
	CreatedAt time.Time `bson:"created_at"`
}

func (r *WebSessionRepository) Create(ctx context.Context, session *domain.WebSession) error {
	if session.ID == "" {
		session.ID = uuid.New().String()
	}
	if session.CreatedAt.IsZero() {
		session.CreatedAt = time.Now()
	}

	doc := webSessionDocument{
		ID:        session.ID,
		UserID:    session.UserID,
		Token:     session.Token,
		ExpiresAt: session.ExpiresAt,
		CreatedAt: session.CreatedAt,
	}

	_, err := r.collection.InsertOne(ctx, doc)
	return err
}

func (r *WebSessionRepository) FindByToken(ctx context.Context, token string) (*domain.WebSession, error) {
	var doc webSessionDocument
	err := r.collection.FindOne(ctx, bson.M{"token": token}).Decode(&doc)
	if err == mongo.ErrNoDocuments {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}

	return &domain.WebSession{
		ID:        doc.ID,
		UserID:    doc.UserID,
		Token:     doc.Token,
		ExpiresAt: doc.ExpiresAt,
		CreatedAt: doc.CreatedAt,
	}, nil
}

func (r *WebSessionRepository) Delete(ctx context.Context, id string) error {
	_, err := r.collection.DeleteOne(ctx, bson.M{"_id": id})
	return err
}

func (r *WebSessionRepository) DeleteExpired(ctx context.Context) error {
	// MongoDB TTL index handles this automatically, but we can also manually delete
	_, err := r.collection.DeleteMany(ctx, bson.M{
		"expires_at": bson.M{"$lt": time.Now()},
	})
	return err
}
