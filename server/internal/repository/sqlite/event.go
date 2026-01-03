package sqlite

import (
	"context"
	"database/sql"
	"encoding/json"
	"strings"
	"time"

	"github.com/google/uuid"
	"github.com/satetsu888/agentrace/server/internal/domain"
	"github.com/satetsu888/agentrace/server/internal/repository"
)

type EventRepository struct {
	db *DB
}

func NewEventRepository(db *DB) *EventRepository {
	return &EventRepository{db: db}
}

func (r *EventRepository) Create(ctx context.Context, event *domain.Event) error {
	if event.ID == "" {
		event.ID = uuid.New().String()
	}
	if event.CreatedAt.IsZero() {
		event.CreatedAt = time.Now()
	}

	payloadJSON, err := json.Marshal(event.Payload)
	if err != nil {
		return err
	}

	// Use sql.NullString for optional uuid
	var uuidValue sql.NullString
	if event.UUID != "" {
		uuidValue = sql.NullString{String: event.UUID, Valid: true}
	}

	_, err = r.db.ExecContext(ctx,
		`INSERT INTO events (id, session_id, uuid, event_type, payload, created_at)
		 VALUES (?, ?, ?, ?, ?, ?)`,
		event.ID, event.SessionID, uuidValue, event.EventType, string(payloadJSON), event.CreatedAt.Format(time.RFC3339),
	)
	if err != nil {
		// Check for UNIQUE constraint violation (duplicate uuid)
		if strings.Contains(err.Error(), "UNIQUE constraint failed") {
			return repository.ErrDuplicateEvent
		}
		return err
	}
	return nil
}

func (r *EventRepository) FindBySessionID(ctx context.Context, sessionID string) ([]*domain.Event, error) {
	rows, err := r.db.QueryContext(ctx,
		`SELECT id, session_id, uuid, event_type, payload, created_at
		 FROM events WHERE session_id = ?
		 ORDER BY created_at ASC`,
		sessionID,
	)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var events []*domain.Event
	for rows.Next() {
		event, err := r.scanEvent(rows)
		if err != nil {
			return nil, err
		}
		events = append(events, event)
	}

	// Sort by payload.timestamp if available (like memory implementation)
	sortEventsByPayloadTimestamp(events)

	return events, rows.Err()
}

func (r *EventRepository) scanEvent(rows *sql.Rows) (*domain.Event, error) {
	var event domain.Event
	var uuidValue sql.NullString
	var payloadStr, createdAt string

	err := rows.Scan(&event.ID, &event.SessionID, &uuidValue, &event.EventType, &payloadStr, &createdAt)
	if err != nil {
		return nil, err
	}

	if uuidValue.Valid {
		event.UUID = uuidValue.String
	}

	if err := json.Unmarshal([]byte(payloadStr), &event.Payload); err != nil {
		// If unmarshal fails, use empty map
		event.Payload = make(map[string]interface{})
	}

	event.CreatedAt, _ = time.Parse(time.RFC3339, createdAt)

	return &event, nil
}

// sortEventsByPayloadTimestamp sorts events by payload.timestamp (ascending)
// This mirrors the behavior in memory/event.go
func sortEventsByPayloadTimestamp(events []*domain.Event) {
	// Simple bubble sort to maintain stability
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
		// Try parsing without timezone
		if parsed, err := time.Parse("2006-01-02T15:04:05.000Z", ts); err == nil {
			return parsed
		}
	}
	return e.CreatedAt
}

func (r *EventRepository) CountBySessionID(ctx context.Context, sessionID string) (int, error) {
	var count int
	err := r.db.QueryRowContext(ctx,
		`SELECT COUNT(*) FROM events WHERE session_id = ?`,
		sessionID,
	).Scan(&count)
	if err != nil {
		return 0, err
	}
	return count, nil
}
