package memory

import (
	"context"
	"sort"
	"sync"
	"time"

	"github.com/google/uuid"
	"github.com/satetsu888/agentrace/server/internal/domain"
)

type SessionRepository struct {
	mu       sync.RWMutex
	sessions map[string]*domain.Session
}

func NewSessionRepository() *SessionRepository {
	return &SessionRepository{
		sessions: make(map[string]*domain.Session),
	}
}

func (r *SessionRepository) Create(ctx context.Context, session *domain.Session) error {
	r.mu.Lock()
	defer r.mu.Unlock()

	if session.ID == "" {
		session.ID = uuid.New().String()
	}
	if session.CreatedAt.IsZero() {
		session.CreatedAt = time.Now()
	}
	if session.StartedAt.IsZero() {
		session.StartedAt = time.Now()
	}

	r.sessions[session.ID] = session
	return nil
}

func (r *SessionRepository) FindByID(ctx context.Context, id string) (*domain.Session, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()

	session, ok := r.sessions[id]
	if !ok {
		return nil, nil
	}
	return session, nil
}

func (r *SessionRepository) FindAll(ctx context.Context, limit int, offset int) ([]*domain.Session, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()

	sessions := make([]*domain.Session, 0, len(r.sessions))
	for _, s := range r.sessions {
		sessions = append(sessions, s)
	}

	// Sort by StartedAt descending (newest first)
	sort.Slice(sessions, func(i, j int) bool {
		return sessions[i].StartedAt.After(sessions[j].StartedAt)
	})

	// Apply offset and limit
	if offset >= len(sessions) {
		return []*domain.Session{}, nil
	}
	sessions = sessions[offset:]
	if limit > 0 && limit < len(sessions) {
		sessions = sessions[:limit]
	}

	return sessions, nil
}

func (r *SessionRepository) FindOrCreateByClaudeSessionID(ctx context.Context, claudeSessionID string, userID *string) (*domain.Session, error) {
	r.mu.Lock()
	defer r.mu.Unlock()

	// Find existing session by ClaudeSessionID
	for _, s := range r.sessions {
		if s.ClaudeSessionID == claudeSessionID {
			// Update UserID if provided and not already set
			if userID != nil && s.UserID == nil {
				s.UserID = userID
			}
			return s, nil
		}
	}

	// Create new session
	session := &domain.Session{
		ID:              uuid.New().String(),
		UserID:          userID,
		ClaudeSessionID: claudeSessionID,
		StartedAt:       time.Now(),
		CreatedAt:       time.Now(),
	}
	r.sessions[session.ID] = session
	return session, nil
}

func (r *SessionRepository) UpdateUserID(ctx context.Context, id string, userID string) error {
	r.mu.Lock()
	defer r.mu.Unlock()

	session, ok := r.sessions[id]
	if !ok {
		return nil
	}
	session.UserID = &userID
	return nil
}

func (r *SessionRepository) UpdateProjectPath(ctx context.Context, id string, projectPath string) error {
	r.mu.Lock()
	defer r.mu.Unlock()

	session, ok := r.sessions[id]
	if !ok {
		return nil
	}
	session.ProjectPath = projectPath
	return nil
}
