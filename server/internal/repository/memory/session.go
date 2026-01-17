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

func (r *SessionRepository) Find(ctx context.Context, query domain.SessionQuery) ([]*domain.Session, string, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()

	// Build user ID set for filtering
	userIDSet := make(map[string]bool)
	for _, uid := range query.UserIDs {
		userIDSet[uid] = true
	}

	sessions := make([]*domain.Session, 0, len(r.sessions))
	for _, s := range r.sessions {
		// Exclude subagents from session list
		if s.IsSidechain {
			continue
		}
		// Filter by project ID
		if query.ProjectID != "" && s.ProjectID != query.ProjectID {
			continue
		}
		// Filter by user IDs
		if len(userIDSet) > 0 {
			if s.UserID == nil || !userIDSet[*s.UserID] {
				continue
			}
		}
		sessions = append(sessions, s)
	}

	// Sort by specified field descending (newest first)
	getSortTime := func(s *domain.Session) time.Time {
		if query.SortBy == "created_at" {
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
	if query.Cursor != "" {
		cursorInfo := repository.DecodeCursor(query.Cursor)
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
	limit := query.Limit
	if limit <= 0 {
		limit = 100
	}
	var nextCursor string
	if limit < len(sessions) {
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

func (r *SessionRepository) FindSubagentsByParentID(ctx context.Context, parentID string) ([]*domain.Session, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()

	sessions := make([]*domain.Session, 0)
	for _, s := range r.sessions {
		if s.ParentSessionID != nil && *s.ParentSessionID == parentID {
			sessions = append(sessions, s)
		}
	}

	// Sort by created_at ascending (oldest first)
	sort.Slice(sessions, func(i, j int) bool {
		return sessions[i].CreatedAt.Before(sessions[j].CreatedAt)
	})

	return sessions, nil
}

func (r *SessionRepository) Delete(ctx context.Context, id string) error {
	r.mu.Lock()
	defer r.mu.Unlock()

	delete(r.sessions, id)
	return nil
}
