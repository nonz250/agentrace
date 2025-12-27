package memory

import (
	"context"
	"sync"
	"time"

	"github.com/google/uuid"
	"github.com/satetsu888/agentrace/server/internal/domain"
)

type EventRepository struct {
	mu     sync.RWMutex
	events map[string]*domain.Event
}

func NewEventRepository() *EventRepository {
	return &EventRepository{
		events: make(map[string]*domain.Event),
	}
}

func (r *EventRepository) Create(ctx context.Context, event *domain.Event) error {
	r.mu.Lock()
	defer r.mu.Unlock()

	if event.ID == "" {
		event.ID = uuid.New().String()
	}
	if event.CreatedAt.IsZero() {
		event.CreatedAt = time.Now()
	}

	r.events[event.ID] = event
	return nil
}

func (r *EventRepository) FindBySessionID(ctx context.Context, sessionID string) ([]*domain.Event, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()

	events := make([]*domain.Event, 0)
	for _, e := range r.events {
		if e.SessionID == sessionID {
			events = append(events, e)
		}
	}
	return events, nil
}
