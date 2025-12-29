package mongodb

import (
	"context"
	"time"

	"github.com/google/uuid"
	"github.com/satetsu888/agentrace/server/internal/domain"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
)

type PasswordCredentialRepository struct {
	collection *mongo.Collection
}

func NewPasswordCredentialRepository(db *DB) *PasswordCredentialRepository {
	return &PasswordCredentialRepository{
		collection: db.Collection("password_credentials"),
	}
}

type passwordCredentialDocument struct {
	ID           string    `bson:"_id"`
	UserID       string    `bson:"user_id"`
	PasswordHash string    `bson:"password_hash"`
	CreatedAt    time.Time `bson:"created_at"`
	UpdatedAt    time.Time `bson:"updated_at"`
}

func (r *PasswordCredentialRepository) Create(ctx context.Context, cred *domain.PasswordCredential) error {
	if cred.ID == "" {
		cred.ID = uuid.New().String()
	}
	now := time.Now()
	if cred.CreatedAt.IsZero() {
		cred.CreatedAt = now
	}
	if cred.UpdatedAt.IsZero() {
		cred.UpdatedAt = now
	}

	doc := passwordCredentialDocument{
		ID:           cred.ID,
		UserID:       cred.UserID,
		PasswordHash: cred.PasswordHash,
		CreatedAt:    cred.CreatedAt,
		UpdatedAt:    cred.UpdatedAt,
	}

	_, err := r.collection.InsertOne(ctx, doc)
	return err
}

func (r *PasswordCredentialRepository) FindByUserID(ctx context.Context, userID string) (*domain.PasswordCredential, error) {
	var doc passwordCredentialDocument
	err := r.collection.FindOne(ctx, bson.M{"user_id": userID}).Decode(&doc)
	if err == mongo.ErrNoDocuments {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}

	return &domain.PasswordCredential{
		ID:           doc.ID,
		UserID:       doc.UserID,
		PasswordHash: doc.PasswordHash,
		CreatedAt:    doc.CreatedAt,
		UpdatedAt:    doc.UpdatedAt,
	}, nil
}

func (r *PasswordCredentialRepository) Update(ctx context.Context, cred *domain.PasswordCredential) error {
	cred.UpdatedAt = time.Now()
	_, err := r.collection.UpdateOne(ctx,
		bson.M{"_id": cred.ID},
		bson.M{"$set": bson.M{
			"password_hash": cred.PasswordHash,
			"updated_at":    cred.UpdatedAt,
		}},
	)
	return err
}

func (r *PasswordCredentialRepository) Delete(ctx context.Context, id string) error {
	_, err := r.collection.DeleteOne(ctx, bson.M{"_id": id})
	return err
}
