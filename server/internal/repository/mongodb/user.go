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

type UserRepository struct {
	collection *mongo.Collection
}

func NewUserRepository(db *DB) *UserRepository {
	return &UserRepository{
		collection: db.Collection("users"),
	}
}

type userDocument struct {
	ID          string    `bson:"_id"`
	Email       string    `bson:"email"`
	DisplayName string    `bson:"display_name,omitempty"`
	CreatedAt   time.Time `bson:"created_at"`
}

func (r *UserRepository) Create(ctx context.Context, user *domain.User) error {
	if user.ID == "" {
		user.ID = uuid.New().String()
	}
	if user.CreatedAt.IsZero() {
		user.CreatedAt = time.Now()
	}

	doc := userDocument{
		ID:          user.ID,
		Email:       user.Email,
		DisplayName: user.DisplayName,
		CreatedAt:   user.CreatedAt,
	}

	_, err := r.collection.InsertOne(ctx, doc)
	return err
}

func (r *UserRepository) FindByID(ctx context.Context, id string) (*domain.User, error) {
	var doc userDocument
	err := r.collection.FindOne(ctx, bson.M{"_id": id}).Decode(&doc)
	if err == mongo.ErrNoDocuments {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}

	return &domain.User{
		ID:          doc.ID,
		Email:       doc.Email,
		DisplayName: doc.DisplayName,
		CreatedAt:   doc.CreatedAt,
	}, nil
}

func (r *UserRepository) FindByEmail(ctx context.Context, email string) (*domain.User, error) {
	var doc userDocument
	err := r.collection.FindOne(ctx, bson.M{"email": email}).Decode(&doc)
	if err == mongo.ErrNoDocuments {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}

	return &domain.User{
		ID:          doc.ID,
		Email:       doc.Email,
		DisplayName: doc.DisplayName,
		CreatedAt:   doc.CreatedAt,
	}, nil
}

func (r *UserRepository) FindAll(ctx context.Context) ([]*domain.User, error) {
	opts := options.Find().SetSort(bson.D{{Key: "created_at", Value: -1}})
	cursor, err := r.collection.Find(ctx, bson.M{}, opts)
	if err != nil {
		return nil, err
	}
	defer cursor.Close(ctx)

	var users []*domain.User
	for cursor.Next(ctx) {
		var doc userDocument
		if err := cursor.Decode(&doc); err != nil {
			return nil, err
		}
		users = append(users, &domain.User{
			ID:          doc.ID,
			Email:       doc.Email,
			DisplayName: doc.DisplayName,
			CreatedAt:   doc.CreatedAt,
		})
	}

	return users, cursor.Err()
}
