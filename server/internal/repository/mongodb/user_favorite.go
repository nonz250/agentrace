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

type UserFavoriteRepository struct {
	collection *mongo.Collection
}

func NewUserFavoriteRepository(db *DB) *UserFavoriteRepository {
	return &UserFavoriteRepository{
		collection: db.Collection("user_favorites"),
	}
}

type userFavoriteDocument struct {
	ID         string    `bson:"_id"`
	UserID     string    `bson:"user_id"`
	TargetType string    `bson:"target_type"`
	TargetID   string    `bson:"target_id"`
	CreatedAt  time.Time `bson:"created_at"`
}

func (r *UserFavoriteRepository) Create(ctx context.Context, favorite *domain.UserFavorite) error {
	if favorite.ID == "" {
		favorite.ID = uuid.New().String()
	}
	if favorite.CreatedAt.IsZero() {
		favorite.CreatedAt = time.Now()
	}

	doc := userFavoriteDocument{
		ID:         favorite.ID,
		UserID:     favorite.UserID,
		TargetType: string(favorite.TargetType),
		TargetID:   favorite.TargetID,
		CreatedAt:  favorite.CreatedAt,
	}

	_, err := r.collection.InsertOne(ctx, doc)
	return err
}

func (r *UserFavoriteRepository) Delete(ctx context.Context, id string) error {
	_, err := r.collection.DeleteOne(ctx, bson.M{"_id": id})
	return err
}

func (r *UserFavoriteRepository) DeleteByUserAndTarget(ctx context.Context, userID string, targetType domain.UserFavoriteTargetType, targetID string) error {
	_, err := r.collection.DeleteOne(ctx, bson.M{
		"user_id":     userID,
		"target_type": string(targetType),
		"target_id":   targetID,
	})
	return err
}

func (r *UserFavoriteRepository) FindByUserID(ctx context.Context, userID string) ([]*domain.UserFavorite, error) {
	opts := options.Find().SetSort(bson.D{{Key: "created_at", Value: -1}})
	cursor, err := r.collection.Find(ctx, bson.M{"user_id": userID}, opts)
	if err != nil {
		return nil, err
	}
	defer cursor.Close(ctx)

	var favorites []*domain.UserFavorite
	for cursor.Next(ctx) {
		var doc userFavoriteDocument
		if err := cursor.Decode(&doc); err != nil {
			return nil, err
		}
		favorites = append(favorites, docToUserFavorite(&doc))
	}

	return favorites, cursor.Err()
}

func (r *UserFavoriteRepository) FindByUserAndTargetType(ctx context.Context, userID string, targetType domain.UserFavoriteTargetType) ([]*domain.UserFavorite, error) {
	opts := options.Find().SetSort(bson.D{{Key: "created_at", Value: -1}})
	cursor, err := r.collection.Find(ctx, bson.M{
		"user_id":     userID,
		"target_type": string(targetType),
	}, opts)
	if err != nil {
		return nil, err
	}
	defer cursor.Close(ctx)

	var favorites []*domain.UserFavorite
	for cursor.Next(ctx) {
		var doc userFavoriteDocument
		if err := cursor.Decode(&doc); err != nil {
			return nil, err
		}
		favorites = append(favorites, docToUserFavorite(&doc))
	}

	return favorites, cursor.Err()
}

func (r *UserFavoriteRepository) FindByUserAndTarget(ctx context.Context, userID string, targetType domain.UserFavoriteTargetType, targetID string) (*domain.UserFavorite, error) {
	var doc userFavoriteDocument
	err := r.collection.FindOne(ctx, bson.M{
		"user_id":     userID,
		"target_type": string(targetType),
		"target_id":   targetID,
	}).Decode(&doc)
	if err == mongo.ErrNoDocuments {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}

	return docToUserFavorite(&doc), nil
}

func (r *UserFavoriteRepository) GetTargetIDs(ctx context.Context, userID string, targetType domain.UserFavoriteTargetType) ([]string, error) {
	cursor, err := r.collection.Find(ctx, bson.M{
		"user_id":     userID,
		"target_type": string(targetType),
	})
	if err != nil {
		return nil, err
	}
	defer cursor.Close(ctx)

	var targetIDs []string
	for cursor.Next(ctx) {
		var doc userFavoriteDocument
		if err := cursor.Decode(&doc); err != nil {
			return nil, err
		}
		targetIDs = append(targetIDs, doc.TargetID)
	}

	return targetIDs, cursor.Err()
}

func docToUserFavorite(doc *userFavoriteDocument) *domain.UserFavorite {
	return &domain.UserFavorite{
		ID:         doc.ID,
		UserID:     doc.UserID,
		TargetType: domain.UserFavoriteTargetType(doc.TargetType),
		TargetID:   doc.TargetID,
		CreatedAt:  doc.CreatedAt,
	}
}
