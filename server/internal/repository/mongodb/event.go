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

type EventRepository struct {
	collection *mongo.Collection
}

func NewEventRepository(db *DB) *EventRepository {
	return &EventRepository{
		collection: db.Collection("events"),
	}
}

type eventDocument struct {
	ID        string                 `bson:"_id"`
	SessionID string                 `bson:"session_id"`
	EventType string                 `bson:"event_type"`
	Payload   map[string]interface{} `bson:"payload"`
	CreatedAt time.Time              `bson:"created_at"`
}

func (r *EventRepository) Create(ctx context.Context, event *domain.Event) error {
	if event.ID == "" {
		event.ID = uuid.New().String()
	}
	if event.CreatedAt.IsZero() {
		event.CreatedAt = time.Now()
	}

	doc := eventDocument{
		ID:        event.ID,
		SessionID: event.SessionID,
		EventType: event.EventType,
		Payload:   event.Payload,
		CreatedAt: event.CreatedAt,
	}

	_, err := r.collection.InsertOne(ctx, doc)
	return err
}

func (r *EventRepository) FindBySessionID(ctx context.Context, sessionID string) ([]*domain.Event, error) {
	// Sort by payload.timestamp if available, otherwise by created_at
	// MongoDB aggregation for sorting by nested field with fallback
	opts := options.Find().SetSort(bson.D{
		{Key: "payload.timestamp", Value: 1},
		{Key: "created_at", Value: 1},
	})

	cursor, err := r.collection.Find(ctx, bson.M{"session_id": sessionID}, opts)
	if err != nil {
		return nil, err
	}
	defer cursor.Close(ctx)

	var events []*domain.Event
	for cursor.Next(ctx) {
		var doc eventDocument
		if err := cursor.Decode(&doc); err != nil {
			return nil, err
		}
		events = append(events, &domain.Event{
			ID:        doc.ID,
			SessionID: doc.SessionID,
			EventType: doc.EventType,
			Payload:   doc.Payload,
			CreatedAt: doc.CreatedAt,
		})
	}

	// Sort by payload.timestamp for correct ordering
	sortEventsByPayloadTimestamp(events)

	return events, cursor.Err()
}

// sortEventsByPayloadTimestamp sorts events by payload.timestamp (ascending)
func sortEventsByPayloadTimestamp(events []*domain.Event) {
	for i := 0; i < len(events)-1; i++ {
		for j := 0; j < len(events)-i-1; j++ {
			t1 := getTimestampFromPayload(events[j])
			t2 := getTimestampFromPayload(events[j+1])
			if t1.After(t2) {
				events[j], events[j+1] = events[j+1], events[j]
			}
		}
	}
}

func getTimestampFromPayload(e *domain.Event) time.Time {
	if ts, ok := e.Payload["timestamp"].(string); ok {
		if parsed, err := time.Parse(time.RFC3339, ts); err == nil {
			return parsed
		}
		if parsed, err := time.Parse("2006-01-02T15:04:05.000Z", ts); err == nil {
			return parsed
		}
	}
	return e.CreatedAt
}
