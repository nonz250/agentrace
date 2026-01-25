package memory

import (
	"context"
	"sort"
	"sync"
	"time"

	"github.com/google/uuid"
	"github.com/satetsu888/agentrace/server/internal/domain"
)

type PlanCommentMessageRepository struct {
	mu       sync.RWMutex
	messages map[string]*domain.PlanCommentMessage
}

func NewPlanCommentMessageRepository() *PlanCommentMessageRepository {
	return &PlanCommentMessageRepository{
		messages: make(map[string]*domain.PlanCommentMessage),
	}
}

func (r *PlanCommentMessageRepository) Create(ctx context.Context, message *domain.PlanCommentMessage) error {
	r.mu.Lock()
	defer r.mu.Unlock()

	if message.ID == "" {
		message.ID = uuid.New().String()
	}
	now := time.Now()
	if message.CreatedAt.IsZero() {
		message.CreatedAt = now
	}
	if message.UpdatedAt.IsZero() {
		message.UpdatedAt = now
	}

	r.messages[message.ID] = message
	return nil
}

func (r *PlanCommentMessageRepository) FindByID(ctx context.Context, id string) (*domain.PlanCommentMessage, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()

	message, ok := r.messages[id]
	if !ok {
		return nil, nil
	}
	return message, nil
}

func (r *PlanCommentMessageRepository) FindByThreadID(ctx context.Context, threadID string) ([]*domain.PlanCommentMessage, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()

	var messages []*domain.PlanCommentMessage
	for _, m := range r.messages {
		if m.ThreadID == threadID {
			messages = append(messages, m)
		}
	}

	// Sort by created_at ascending
	sort.Slice(messages, func(i, j int) bool {
		return messages[i].CreatedAt.Before(messages[j].CreatedAt)
	})

	return messages, nil
}

func (r *PlanCommentMessageRepository) Update(ctx context.Context, message *domain.PlanCommentMessage) error {
	r.mu.Lock()
	defer r.mu.Unlock()

	if _, ok := r.messages[message.ID]; !ok {
		return nil
	}

	message.UpdatedAt = time.Now()
	r.messages[message.ID] = message
	return nil
}

func (r *PlanCommentMessageRepository) Delete(ctx context.Context, id string) error {
	r.mu.Lock()
	defer r.mu.Unlock()

	delete(r.messages, id)
	return nil
}

func (r *PlanCommentMessageRepository) DeleteByThreadID(ctx context.Context, threadID string) error {
	r.mu.Lock()
	defer r.mu.Unlock()

	for id, m := range r.messages {
		if m.ThreadID == threadID {
			delete(r.messages, id)
		}
	}
	return nil
}
