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

type PlanDocumentEventRepository struct {
	collection *mongo.Collection
}

func NewPlanDocumentEventRepository(db *DB) *PlanDocumentEventRepository {
	return &PlanDocumentEventRepository{
		collection: db.Collection("plan_document_events"),
	}
}

type planDocumentEventDocument struct {
	ID             string    `bson:"_id"`
	PlanDocumentID string    `bson:"plan_document_id"`
	SessionID      *string   `bson:"session_id,omitempty"`
	UserID         *string   `bson:"user_id,omitempty"`
	Patch          string    `bson:"patch"`
	CreatedAt      time.Time `bson:"created_at"`
}

func (r *PlanDocumentEventRepository) Create(ctx context.Context, event *domain.PlanDocumentEvent) error {
	if event.ID == "" {
		event.ID = uuid.New().String()
	}
	if event.CreatedAt.IsZero() {
		event.CreatedAt = time.Now()
	}

	doc := planDocumentEventDocument{
		ID:             event.ID,
		PlanDocumentID: event.PlanDocumentID,
		SessionID:      event.SessionID,
		UserID:         event.UserID,
		Patch:          event.Patch,
		CreatedAt:      event.CreatedAt,
	}

	_, err := r.collection.InsertOne(ctx, doc)
	return err
}

func (r *PlanDocumentEventRepository) FindByPlanDocumentID(ctx context.Context, planDocumentID string) ([]*domain.PlanDocumentEvent, error) {
	opts := options.Find().SetSort(bson.D{{Key: "created_at", Value: 1}})

	cursor, err := r.collection.Find(ctx, bson.M{"plan_document_id": planDocumentID}, opts)
	if err != nil {
		return nil, err
	}
	defer cursor.Close(ctx)

	var events []*domain.PlanDocumentEvent
	for cursor.Next(ctx) {
		var doc planDocumentEventDocument
		if err := cursor.Decode(&doc); err != nil {
			return nil, err
		}
		events = append(events, docToPlanDocumentEvent(&doc))
	}

	return events, cursor.Err()
}

func (r *PlanDocumentEventRepository) FindBySessionID(ctx context.Context, sessionID string) ([]*domain.PlanDocumentEvent, error) {
	opts := options.Find().SetSort(bson.D{{Key: "created_at", Value: 1}})

	cursor, err := r.collection.Find(ctx, bson.M{"session_id": sessionID}, opts)
	if err != nil {
		return nil, err
	}
	defer cursor.Close(ctx)

	var events []*domain.PlanDocumentEvent
	for cursor.Next(ctx) {
		var doc planDocumentEventDocument
		if err := cursor.Decode(&doc); err != nil {
			return nil, err
		}
		events = append(events, docToPlanDocumentEvent(&doc))
	}

	return events, cursor.Err()
}

func (r *PlanDocumentEventRepository) GetCollaboratorUserIDs(ctx context.Context, planDocumentID string) ([]string, error) {
	// Use distinct to get unique user_ids
	results, err := r.collection.Distinct(ctx, "user_id", bson.M{
		"plan_document_id": planDocumentID,
		"user_id":          bson.M{"$ne": nil},
	})
	if err != nil {
		return nil, err
	}

	var userIDs []string
	for _, result := range results {
		if userID, ok := result.(string); ok {
			userIDs = append(userIDs, userID)
		}
	}

	return userIDs, nil
}

func docToPlanDocumentEvent(doc *planDocumentEventDocument) *domain.PlanDocumentEvent {
	return &domain.PlanDocumentEvent{
		ID:             doc.ID,
		PlanDocumentID: doc.PlanDocumentID,
		SessionID:      doc.SessionID,
		UserID:         doc.UserID,
		Patch:          doc.Patch,
		CreatedAt:      doc.CreatedAt,
	}
}
