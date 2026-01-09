package memory

import (
	"context"
	"sort"
	"sync"
	"time"

	"github.com/google/uuid"
	"github.com/satetsu888/agentrace/server/internal/domain"
)

type PlanDocumentEventRepository struct {
	mu     sync.RWMutex
	events map[string]*domain.PlanDocumentEvent
}

func NewPlanDocumentEventRepository() *PlanDocumentEventRepository {
	return &PlanDocumentEventRepository{
		events: make(map[string]*domain.PlanDocumentEvent),
	}
}

func (r *PlanDocumentEventRepository) Create(ctx context.Context, event *domain.PlanDocumentEvent) error {
	r.mu.Lock()
	defer r.mu.Unlock()

	if event.ID == "" {
		event.ID = uuid.New().String()
	}
	if event.CreatedAt.IsZero() {
		event.CreatedAt = time.Now()
	}
	if event.EventType == "" {
		event.EventType = domain.PlanDocumentEventTypeBodyChange
	}

	r.events[event.ID] = event
	return nil
}

func (r *PlanDocumentEventRepository) FindByPlanDocumentID(ctx context.Context, planDocumentID string) ([]*domain.PlanDocumentEvent, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()

	events := make([]*domain.PlanDocumentEvent, 0)
	for _, e := range r.events {
		if e.PlanDocumentID == planDocumentID {
			events = append(events, e)
		}
	}

	// Sort by CreatedAt ascending (oldest first)
	sort.Slice(events, func(i, j int) bool {
		return events[i].CreatedAt.Before(events[j].CreatedAt)
	})

	return events, nil
}

func (r *PlanDocumentEventRepository) FindByClaudeSessionID(ctx context.Context, claudeSessionID string) ([]*domain.PlanDocumentEvent, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()

	events := make([]*domain.PlanDocumentEvent, 0)
	for _, e := range r.events {
		if e.ClaudeSessionID != nil && *e.ClaudeSessionID == claudeSessionID {
			events = append(events, e)
		}
	}

	// Sort by CreatedAt ascending (oldest first)
	sort.Slice(events, func(i, j int) bool {
		return events[i].CreatedAt.Before(events[j].CreatedAt)
	})

	return events, nil
}

func (r *PlanDocumentEventRepository) GetCollaboratorUserIDs(ctx context.Context, planDocumentID string) ([]string, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()

	userIDSet := make(map[string]struct{})
	for _, e := range r.events {
		if e.PlanDocumentID == planDocumentID && e.UserID != nil {
			userIDSet[*e.UserID] = struct{}{}
		}
	}

	userIDs := make([]string, 0, len(userIDSet))
	for userID := range userIDSet {
		userIDs = append(userIDs, userID)
	}

	return userIDs, nil
}

func (r *PlanDocumentEventRepository) GetPlanDocumentIDsByUserIDs(ctx context.Context, userIDs []string) ([]string, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()

	if len(userIDs) == 0 {
		return []string{}, nil
	}

	// Build user ID set for fast lookup
	userIDSet := make(map[string]struct{})
	for _, id := range userIDs {
		userIDSet[id] = struct{}{}
	}

	// Find plan document IDs where any of the specified users have events
	planDocIDSet := make(map[string]struct{})
	for _, e := range r.events {
		if e.UserID != nil {
			if _, ok := userIDSet[*e.UserID]; ok {
				planDocIDSet[e.PlanDocumentID] = struct{}{}
			}
		}
	}

	planDocIDs := make([]string, 0, len(planDocIDSet))
	for id := range planDocIDSet {
		planDocIDs = append(planDocIDs, id)
	}

	return planDocIDs, nil
}
