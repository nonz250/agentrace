package memory

import (
	"context"
	"sort"
	"sync"
	"time"

	"github.com/google/uuid"
	"github.com/satetsu888/agentrace/server/internal/domain"
	"github.com/satetsu888/agentrace/server/internal/repository"
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

func (r *SessionRepository) FindAll(ctx context.Context, limit int, cursor string, sortBy string) ([]*domain.Session, string, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()

	sessions := make([]*domain.Session, 0, len(r.sessions))
	for _, s := range r.sessions {
		sessions = append(sessions, s)
	}

	// Sort by specified field descending (newest first)
	getSortTime := func(s *domain.Session) time.Time {
		if sortBy == "created_at" {
			return s.CreatedAt
		}
		return s.UpdatedAt
	}

	sort.Slice(sessions, func(i, j int) bool {
		ti, tj := getSortTime(sessions[i]), getSortTime(sessions[j])
		if ti.Equal(tj) {
			return sessions[i].ID > sessions[j].ID
		}
		return ti.After(tj)
	})

	// Apply cursor filter
	if cursor != "" {
		cursorInfo := repository.DecodeCursor(cursor)
		if cursorInfo != nil {
			cursorTime, err := cursorInfo.ParseSortTime()
			if err == nil {
				startIdx := 0
				for i, s := range sessions {
					sortTime := getSortTime(s)
					if sortTime.Before(cursorTime) || (sortTime.Equal(cursorTime) && s.ID < cursorInfo.ID) {
						startIdx = i
						break
					}
				}
				sessions = sessions[startIdx:]
			}
		}
	}

	// Apply limit and generate next cursor
	var nextCursor string
	if limit > 0 && limit < len(sessions) {
		lastItem := sessions[limit-1]
		nextCursor = repository.EncodeCursor(getSortTime(lastItem), lastItem.ID)
		sessions = sessions[:limit]
	}

	return sessions, nextCursor, nil
}

func (r *SessionRepository) FindByProjectID(ctx context.Context, projectID string, limit int, cursor string, sortBy string) ([]*domain.Session, string, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()

	sessions := make([]*domain.Session, 0)
	for _, s := range r.sessions {
		if s.ProjectID == projectID {
			sessions = append(sessions, s)
		}
	}

	// Sort by specified field descending (newest first)
	getSortTime := func(s *domain.Session) time.Time {
		if sortBy == "created_at" {
			return s.CreatedAt
		}
		return s.UpdatedAt
	}

	sort.Slice(sessions, func(i, j int) bool {
		ti, tj := getSortTime(sessions[i]), getSortTime(sessions[j])
		if ti.Equal(tj) {
			return sessions[i].ID > sessions[j].ID
		}
		return ti.After(tj)
	})

	// Apply cursor filter
	if cursor != "" {
		cursorInfo := repository.DecodeCursor(cursor)
		if cursorInfo != nil {
			cursorTime, err := cursorInfo.ParseSortTime()
			if err == nil {
				startIdx := 0
				for i, s := range sessions {
					sortTime := getSortTime(s)
					if sortTime.Before(cursorTime) || (sortTime.Equal(cursorTime) && s.ID < cursorInfo.ID) {
						startIdx = i
						break
					}
				}
				sessions = sessions[startIdx:]
			}
		}
	}

	// Apply limit and generate next cursor
	var nextCursor string
	if limit > 0 && limit < len(sessions) {
		lastItem := sessions[limit-1]
		nextCursor = repository.EncodeCursor(getSortTime(lastItem), lastItem.ID)
		sessions = sessions[:limit]
	}

	return sessions, nextCursor, nil
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

func (r *SessionRepository) UpdateTitle(ctx context.Context, id string, title string) error {
	r.mu.Lock()
	defer r.mu.Unlock()

	session, ok := r.sessions[id]
	if !ok {
		return nil
	}
	session.Title = &title
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
