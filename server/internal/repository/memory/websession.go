package memory

import (
	"context"
	"sync"
	"time"

	"github.com/google/uuid"
	"github.com/satetsu888/agentrace/server/internal/domain"
)

type WebSessionRepository struct {
	mu       sync.RWMutex
	sessions map[string]*domain.WebSession
}

func NewWebSessionRepository() *WebSessionRepository {
	return &WebSessionRepository{
		sessions: make(map[string]*domain.WebSession),
	}
}

func (r *WebSessionRepository) Create(ctx context.Context, session *domain.WebSession) error {
	r.mu.Lock()
	defer r.mu.Unlock()

	if session.ID == "" {
		session.ID = uuid.New().String()
	}
	if session.CreatedAt.IsZero() {
		session.CreatedAt = time.Now()
	}

	r.sessions[session.ID] = session
	return nil
}

func (r *WebSessionRepository) FindByToken(ctx context.Context, token string) (*domain.WebSession, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()

	for _, s := range r.sessions {
		if s.Token == token {
			return s, nil
		}
	}
	return nil, nil
}

func (r *WebSessionRepository) Delete(ctx context.Context, id string) error {
	r.mu.Lock()
	defer r.mu.Unlock()

	delete(r.sessions, id)
	return nil
}

func (r *WebSessionRepository) DeleteExpired(ctx context.Context) error {
	r.mu.Lock()
	defer r.mu.Unlock()

	now := time.Now()
	for id, s := range r.sessions {
		if now.After(s.ExpiresAt) {
			delete(r.sessions, id)
		}
	}
	return nil
}
