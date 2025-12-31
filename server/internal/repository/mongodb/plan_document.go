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

type PlanDocumentRepository struct {
	collection *mongo.Collection
}

func NewPlanDocumentRepository(db *DB) *PlanDocumentRepository {
	return &PlanDocumentRepository{
		collection: db.Collection("plan_documents"),
	}
}

type planDocumentDocument struct {
	ID           string    `bson:"_id"`
	Description  string    `bson:"description"`
	Body         string    `bson:"body"`
	GitRemoteURL string    `bson:"git_remote_url"`
	CreatedAt    time.Time `bson:"created_at"`
	UpdatedAt    time.Time `bson:"updated_at"`
}

func (r *PlanDocumentRepository) Create(ctx context.Context, doc *domain.PlanDocument) error {
	if doc.ID == "" {
		doc.ID = uuid.New().String()
	}
	now := time.Now()
	if doc.CreatedAt.IsZero() {
		doc.CreatedAt = now
	}
	if doc.UpdatedAt.IsZero() {
		doc.UpdatedAt = now
	}

	mongoDoc := planDocumentDocument{
		ID:           doc.ID,
		Description:  doc.Description,
		Body:         doc.Body,
		GitRemoteURL: doc.GitRemoteURL,
		CreatedAt:    doc.CreatedAt,
		UpdatedAt:    doc.UpdatedAt,
	}

	_, err := r.collection.InsertOne(ctx, mongoDoc)
	return err
}

func (r *PlanDocumentRepository) FindByID(ctx context.Context, id string) (*domain.PlanDocument, error) {
	var doc planDocumentDocument
	err := r.collection.FindOne(ctx, bson.M{"_id": id}).Decode(&doc)
	if err == mongo.ErrNoDocuments {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}

	return docToPlanDocument(&doc), nil
}

func (r *PlanDocumentRepository) FindAll(ctx context.Context, limit int, offset int) ([]*domain.PlanDocument, error) {
	opts := options.Find().SetSort(bson.D{{Key: "updated_at", Value: -1}})
	if limit > 0 {
		opts.SetLimit(int64(limit))
		opts.SetSkip(int64(offset))
	}

	cursor, err := r.collection.Find(ctx, bson.M{}, opts)
	if err != nil {
		return nil, err
	}
	defer cursor.Close(ctx)

	var docs []*domain.PlanDocument
	for cursor.Next(ctx) {
		var doc planDocumentDocument
		if err := cursor.Decode(&doc); err != nil {
			return nil, err
		}
		docs = append(docs, docToPlanDocument(&doc))
	}

	return docs, cursor.Err()
}

func (r *PlanDocumentRepository) FindByGitRemoteURL(ctx context.Context, gitRemoteURL string, limit int, offset int) ([]*domain.PlanDocument, error) {
	opts := options.Find().SetSort(bson.D{{Key: "updated_at", Value: -1}})
	if limit > 0 {
		opts.SetLimit(int64(limit))
		opts.SetSkip(int64(offset))
	}

	cursor, err := r.collection.Find(ctx, bson.M{"git_remote_url": gitRemoteURL}, opts)
	if err != nil {
		return nil, err
	}
	defer cursor.Close(ctx)

	var docs []*domain.PlanDocument
	for cursor.Next(ctx) {
		var doc planDocumentDocument
		if err := cursor.Decode(&doc); err != nil {
			return nil, err
		}
		docs = append(docs, docToPlanDocument(&doc))
	}

	return docs, cursor.Err()
}

func (r *PlanDocumentRepository) Update(ctx context.Context, doc *domain.PlanDocument) error {
	doc.UpdatedAt = time.Now()

	_, err := r.collection.UpdateOne(ctx,
		bson.M{"_id": doc.ID},
		bson.M{"$set": bson.M{
			"description":    doc.Description,
			"body":           doc.Body,
			"git_remote_url": doc.GitRemoteURL,
			"updated_at":     doc.UpdatedAt,
		}},
	)
	return err
}

func (r *PlanDocumentRepository) Delete(ctx context.Context, id string) error {
	_, err := r.collection.DeleteOne(ctx, bson.M{"_id": id})
	return err
}

func docToPlanDocument(doc *planDocumentDocument) *domain.PlanDocument {
	return &domain.PlanDocument{
		ID:           doc.ID,
		Description:  doc.Description,
		Body:         doc.Body,
		GitRemoteURL: doc.GitRemoteURL,
		CreatedAt:    doc.CreatedAt,
		UpdatedAt:    doc.UpdatedAt,
	}
}
