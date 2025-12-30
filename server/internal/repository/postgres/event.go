package postgres

import (
	"context"
	"database/sql"
	"encoding/json"
	"time"

	"github.com/google/uuid"
	"github.com/satetsu888/agentrace/server/internal/domain"
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

	_, err = r.db.ExecContext(ctx,
		`INSERT INTO events (id, session_id, event_type, payload, created_at)
		 VALUES ($1, $2, $3, $4, $5)`,
		event.ID, event.SessionID, event.EventType, payloadJSON, event.CreatedAt,
	)
	return err
}

func (r *EventRepository) FindBySessionID(ctx context.Context, sessionID string) ([]*domain.Event, error) {
	// Order by payload->>'timestamp' if available, otherwise by created_at
	rows, err := r.db.QueryContext(ctx,
		`SELECT id, session_id, event_type, payload, created_at
		 FROM events WHERE session_id = $1
		 ORDER BY COALESCE(payload->>'timestamp', created_at::text) ASC`,
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

	return events, rows.Err()
}

func (r *EventRepository) scanEvent(rows *sql.Rows) (*domain.Event, error) {
	var event domain.Event
	var payloadBytes []byte

	err := rows.Scan(&event.ID, &event.SessionID, &event.EventType, &payloadBytes, &event.CreatedAt)
	if err != nil {
		return nil, err
	}

	if err := json.Unmarshal(payloadBytes, &event.Payload); err != nil {
		// If unmarshal fails, use empty map
		event.Payload = make(map[string]interface{})
	}

	return &event, nil
}

func (r *EventRepository) CountBySessionID(ctx context.Context, sessionID string) (int, error) {
	var count int
	err := r.db.QueryRowContext(ctx,
		`SELECT COUNT(*) FROM events WHERE session_id = $1`,
		sessionID,
	).Scan(&count)
	if err != nil {
		return 0, err
	}
	return count, nil
}
