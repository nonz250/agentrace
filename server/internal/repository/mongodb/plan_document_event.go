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
	ID              string    `bson:"_id"`
	PlanDocumentID  string    `bson:"plan_document_id"`
	ClaudeSessionID *string   `bson:"claude_session_id,omitempty"`
	ToolUseID       *string   `bson:"tool_use_id,omitempty"`
	UserID          *string   `bson:"user_id,omitempty"`
	EventType       string    `bson:"event_type"`
	Patch           string    `bson:"patch"`
	CreatedAt       time.Time `bson:"created_at"`
}

func (r *PlanDocumentEventRepository) Create(ctx context.Context, event *domain.PlanDocumentEvent) error {
	if event.ID == "" {
		event.ID = uuid.New().String()
	}
	if event.CreatedAt.IsZero() {
		event.CreatedAt = time.Now()
	}
	if event.EventType == "" {
		event.EventType = domain.PlanDocumentEventTypeBodyChange
	}

	doc := planDocumentEventDocument{
		ID:              event.ID,
		PlanDocumentID:  event.PlanDocumentID,
		ClaudeSessionID: event.ClaudeSessionID,
		ToolUseID:       event.ToolUseID,
		UserID:          event.UserID,
		EventType:       string(event.EventType),
		Patch:           event.Patch,
		CreatedAt:       event.CreatedAt,
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

func (r *PlanDocumentEventRepository) FindByClaudeSessionID(ctx context.Context, claudeSessionID string) ([]*domain.PlanDocumentEvent, error) {
	opts := options.Find().SetSort(bson.D{{Key: "created_at", Value: 1}})

	cursor, err := r.collection.Find(ctx, bson.M{"claude_session_id": claudeSessionID}, opts)
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

func (r *PlanDocumentEventRepository) GetPlanDocumentIDsByUserIDs(ctx context.Context, userIDs []string) ([]string, error) {
	if len(userIDs) == 0 {
		return []string{}, nil
	}

	// Use distinct to get unique plan_document_ids for the given user_ids
	results, err := r.collection.Distinct(ctx, "plan_document_id", bson.M{
		"user_id": bson.M{"$in": userIDs},
	})
	if err != nil {
		return nil, err
	}

	var planDocIDs []string
	for _, result := range results {
		if planDocID, ok := result.(string); ok {
			planDocIDs = append(planDocIDs, planDocID)
		}
	}

	return planDocIDs, nil
}

func docToPlanDocumentEvent(doc *planDocumentEventDocument) *domain.PlanDocumentEvent {
	eventType := domain.PlanDocumentEventType(doc.EventType)
	if eventType == "" {
		eventType = domain.PlanDocumentEventTypeBodyChange
	}
	return &domain.PlanDocumentEvent{
		ID:              doc.ID,
		PlanDocumentID:  doc.PlanDocumentID,
		ClaudeSessionID: doc.ClaudeSessionID,
		ToolUseID:       doc.ToolUseID,
		UserID:          doc.UserID,
		EventType:       eventType,
		Patch:           doc.Patch,
		CreatedAt:       doc.CreatedAt,
	}
}
