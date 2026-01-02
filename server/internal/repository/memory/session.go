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
	if session.UpdatedAt.IsZero() {
		session.UpdatedAt = session.StartedAt
	}
	if session.ProjectID == "" {
		session.ProjectID = domain.DefaultProjectID
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

func (r *SessionRepository) FindByClaudeSessionID(ctx context.Context, claudeSessionID string) (*domain.Session, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()

	for _, session := range r.sessions {
		if session.ClaudeSessionID == claudeSessionID {
			return session, nil
		}
	}
	return nil, nil
}

func (r *SessionRepository) FindAll(ctx context.Context, limit int, offset int) ([]*domain.Session, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()

	sessions := make([]*domain.Session, 0, len(r.sessions))
	for _, s := range r.sessions {
		sessions = append(sessions, s)
	}

	// Sort by UpdatedAt descending (newest first)
	sort.Slice(sessions, func(i, j int) bool {
		return sessions[i].UpdatedAt.After(sessions[j].UpdatedAt)
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

func (r *SessionRepository) FindByProjectID(ctx context.Context, projectID string, limit int, offset int) ([]*domain.Session, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()

	sessions := make([]*domain.Session, 0)
	for _, s := range r.sessions {
		if s.ProjectID == projectID {
			sessions = append(sessions, s)
		}
	}

	// Sort by UpdatedAt descending (newest first)
	sort.Slice(sessions, func(i, j int) bool {
		return sessions[i].UpdatedAt.After(sessions[j].UpdatedAt)
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
	now := time.Now()
	session := &domain.Session{
		ID:              uuid.New().String(),
		UserID:          userID,
		ProjectID:       domain.DefaultProjectID,
		ClaudeSessionID: claudeSessionID,
		StartedAt:       now,
		UpdatedAt:       now,
		CreatedAt:       now,
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

func (r *SessionRepository) UpdateProjectID(ctx context.Context, id string, projectID string) error {
	r.mu.Lock()
	defer r.mu.Unlock()

	session, ok := r.sessions[id]
	if !ok {
		return nil
	}
	session.ProjectID = projectID
	return nil
}

func (r *SessionRepository) UpdateGitBranch(ctx context.Context, id string, gitBranch string) error {
	r.mu.Lock()
	defer r.mu.Unlock()

	session, ok := r.sessions[id]
	if !ok {
		return nil
	}
	session.GitBranch = gitBranch
	return nil
}

func (r *SessionRepository) UpdateUpdatedAt(ctx context.Context, id string, updatedAt time.Time) error {
	r.mu.Lock()
	defer r.mu.Unlock()

	session, ok := r.sessions[id]
	if !ok {
		return nil
	}
	session.UpdatedAt = updatedAt
	return nil
}
