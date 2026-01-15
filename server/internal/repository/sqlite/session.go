package sqlite

import (
	"context"
	"database/sql"
	"time"

	"github.com/google/uuid"
	"github.com/satetsu888/agentrace/server/internal/domain"
	"github.com/satetsu888/agentrace/server/internal/repository"
)

type SessionRepository struct {
	db *DB
}

func NewSessionRepository(db *DB) *SessionRepository {
	return &SessionRepository{db: db}
}

func (r *SessionRepository) Create(ctx context.Context, session *domain.Session) error {
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

	var endedAt *string
	if session.EndedAt != nil {
		s := session.EndedAt.Format(time.RFC3339)
		endedAt = &s
	}

	_, err := r.db.ExecContext(ctx,
		`INSERT INTO sessions (id, user_id, project_id, claude_session_id, project_path, git_branch, title, started_at, ended_at, updated_at, created_at)
		 VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?)`,
		session.ID, session.UserID, session.ProjectID, session.ClaudeSessionID, session.ProjectPath,
		session.GitBranch, session.Title,
		session.StartedAt.Format(time.RFC3339), endedAt, session.UpdatedAt.Format(time.RFC3339), session.CreatedAt.Format(time.RFC3339),
	)
	return err
}

func (r *SessionRepository) FindByID(ctx context.Context, id string) (*domain.Session, error) {
	return r.scanSession(r.db.QueryRowContext(ctx,
		`SELECT id, user_id, project_id, claude_session_id, project_path, git_branch, title, started_at, ended_at, updated_at, created_at
		 FROM sessions WHERE id = ?`,
		id,
	))
}

func (r *SessionRepository) FindByClaudeSessionID(ctx context.Context, claudeSessionID string) (*domain.Session, error) {
	return r.scanSession(r.db.QueryRowContext(ctx,
		`SELECT id, user_id, project_id, claude_session_id, project_path, git_branch, title, started_at, ended_at, updated_at, created_at
		 FROM sessions WHERE claude_session_id = ?`,
		claudeSessionID,
	))
}

func (r *SessionRepository) FindAll(ctx context.Context, limit int, cursor string, sortBy string) ([]*domain.Session, string, error) {
	// Validate sortBy to prevent SQL injection
	orderColumn := "updated_at"
	if sortBy == "created_at" {
		orderColumn = "created_at"
	}

	query := `SELECT id, user_id, project_id, claude_session_id, project_path, git_branch, title, started_at, ended_at, updated_at, created_at
		 FROM sessions`

	var args []any

	// Apply cursor filter
	if cursor != "" {
		cursorInfo := repository.DecodeCursor(cursor)
		if cursorInfo != nil {
			cursorTime, err := cursorInfo.ParseSortTime()
			if err == nil {
				query += ` WHERE (` + orderColumn + ` < ? OR (` + orderColumn + ` = ? AND id < ?))`
				cursorTimeStr := cursorTime.Format(time.RFC3339Nano)
				args = append(args, cursorTimeStr, cursorTimeStr, cursorInfo.ID)
			}
		}
	}

	query += ` ORDER BY ` + orderColumn + ` DESC, id DESC`

	if limit > 0 {
		query += ` LIMIT ?`
		args = append(args, limit+1)
	}

	rows, err := r.db.QueryContext(ctx, query, args...)
	if err != nil {
		return nil, "", err
	}
	defer rows.Close()

	var sessions []*domain.Session
	for rows.Next() {
		session, err := r.scanSessionFromRows(rows)
		if err != nil {
			return nil, "", err
		}
		sessions = append(sessions, session)
	}

	if err := rows.Err(); err != nil {
		return nil, "", err
	}

	// Generate next cursor if there are more results
	var nextCursor string
	if limit > 0 && len(sessions) > limit {
		sessions = sessions[:limit]
		lastItem := sessions[limit-1]
		var sortTime time.Time
		if sortBy == "created_at" {
			sortTime = lastItem.CreatedAt
		} else {
			sortTime = lastItem.UpdatedAt
		}
		nextCursor = repository.EncodeCursor(sortTime, lastItem.ID)
	}

	return sessions, nextCursor, nil
}

func (r *SessionRepository) FindByProjectID(ctx context.Context, projectID string, limit int, cursor string, sortBy string) ([]*domain.Session, string, error) {
	// Validate sortBy to prevent SQL injection
	orderColumn := "updated_at"
	if sortBy == "created_at" {
		orderColumn = "created_at"
	}

	query := `SELECT id, user_id, project_id, claude_session_id, project_path, git_branch, title, started_at, ended_at, updated_at, created_at
		 FROM sessions WHERE project_id = ?`

	args := []any{projectID}

	// Apply cursor filter
	if cursor != "" {
		cursorInfo := repository.DecodeCursor(cursor)
		if cursorInfo != nil {
			cursorTime, err := cursorInfo.ParseSortTime()
			if err == nil {
				query += ` AND (` + orderColumn + ` < ? OR (` + orderColumn + ` = ? AND id < ?))`
				cursorTimeStr := cursorTime.Format(time.RFC3339Nano)
				args = append(args, cursorTimeStr, cursorTimeStr, cursorInfo.ID)
			}
		}
	}

	query += ` ORDER BY ` + orderColumn + ` DESC, id DESC`

	if limit > 0 {
		query += ` LIMIT ?`
		args = append(args, limit+1)
	}

	rows, err := r.db.QueryContext(ctx, query, args...)
	if err != nil {
		return nil, "", err
	}
	defer rows.Close()

	var sessions []*domain.Session
	for rows.Next() {
		session, err := r.scanSessionFromRows(rows)
		if err != nil {
			return nil, "", err
		}
		sessions = append(sessions, session)
	}

	if err := rows.Err(); err != nil {
		return nil, "", err
	}

	// Generate next cursor if there are more results
	var nextCursor string
	if limit > 0 && len(sessions) > limit {
		sessions = sessions[:limit]
		lastItem := sessions[limit-1]
		var sortTime time.Time
		if sortBy == "created_at" {
			sortTime = lastItem.CreatedAt
		} else {
			sortTime = lastItem.UpdatedAt
		}
		nextCursor = repository.EncodeCursor(sortTime, lastItem.ID)
	}

	return sessions, nextCursor, nil
}

func (r *SessionRepository) FindOrCreateByClaudeSessionID(ctx context.Context, claudeSessionID string, userID *string) (*domain.Session, error) {
	// First try to find existing session
	session, err := r.scanSession(r.db.QueryRowContext(ctx,
		`SELECT id, user_id, project_id, claude_session_id, project_path, git_branch, title, started_at, ended_at, updated_at, created_at
		 FROM sessions WHERE claude_session_id = ?`,
		claudeSessionID,
	))

	if err != nil {
		return nil, err
	}

	if session != nil {
		// Update UserID if provided and not already set
		if userID != nil && session.UserID == nil {
			_, err := r.db.ExecContext(ctx,
				`UPDATE sessions SET user_id = ? WHERE id = ?`,
				*userID, session.ID,
			)
			if err != nil {
				return nil, err
			}
			session.UserID = userID
		}
		return session, nil
	}

	// Create new session
	newSession := &domain.Session{
		ID:              uuid.New().String(),
		UserID:          userID,
		ProjectID:       domain.DefaultProjectID,
		ClaudeSessionID: claudeSessionID,
		StartedAt:       time.Now(),
		CreatedAt:       time.Now(),
	}

	if err := r.Create(ctx, newSession); err != nil {
		return nil, err
	}

	return newSession, nil
}

func (r *SessionRepository) UpdateUserID(ctx context.Context, id string, userID string) error {
	_, err := r.db.ExecContext(ctx,
		`UPDATE sessions SET user_id = ? WHERE id = ?`,
		userID, id,
	)
	return err
}

func (r *SessionRepository) UpdateProjectPath(ctx context.Context, id string, projectPath string) error {
	_, err := r.db.ExecContext(ctx,
		`UPDATE sessions SET project_path = ? WHERE id = ?`,
		projectPath, id,
	)
	return err
}

func (r *SessionRepository) UpdateProjectID(ctx context.Context, id string, projectID string) error {
	_, err := r.db.ExecContext(ctx,
		`UPDATE sessions SET project_id = ? WHERE id = ?`,
		projectID, id,
	)
	return err
}

func (r *SessionRepository) UpdateGitBranch(ctx context.Context, id string, gitBranch string) error {
	_, err := r.db.ExecContext(ctx,
		`UPDATE sessions SET git_branch = ? WHERE id = ?`,
		gitBranch, id,
	)
	return err
}

func (r *SessionRepository) UpdateTitle(ctx context.Context, id string, title string) error {
	_, err := r.db.ExecContext(ctx,
		`UPDATE sessions SET title = ? WHERE id = ?`,
		title, id,
	)
	return err
}

func (r *SessionRepository) UpdateUpdatedAt(ctx context.Context, id string, updatedAt time.Time) error {
	_, err := r.db.ExecContext(ctx,
		`UPDATE sessions SET updated_at = ? WHERE id = ?`,
		updatedAt.Format(time.RFC3339), id,
	)
	return err
}

func (r *SessionRepository) scanSession(row *sql.Row) (*domain.Session, error) {
	var session domain.Session
	var userID, projectID, projectPath, gitBranch, title, startedAt, endedAt, updatedAt, createdAt sql.NullString

	err := row.Scan(&session.ID, &userID, &projectID, &session.ClaudeSessionID, &projectPath, &gitBranch, &title, &startedAt, &endedAt, &updatedAt, &createdAt)
	if err == sql.ErrNoRows {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}

	if userID.Valid {
		session.UserID = &userID.String
	}
	if projectID.Valid {
		session.ProjectID = projectID.String
	} else {
		session.ProjectID = domain.DefaultProjectID
	}
	if projectPath.Valid {
		session.ProjectPath = projectPath.String
	}
	if gitBranch.Valid {
		session.GitBranch = gitBranch.String
	}
	if title.Valid {
		session.Title = &title.String
	}
	if startedAt.Valid {
		session.StartedAt, _ = time.Parse(time.RFC3339, startedAt.String)
	}
	if endedAt.Valid {
		t, _ := time.Parse(time.RFC3339, endedAt.String)
		session.EndedAt = &t
	}
	if updatedAt.Valid {
		session.UpdatedAt, _ = time.Parse(time.RFC3339, updatedAt.String)
	}
	if createdAt.Valid {
		session.CreatedAt, _ = time.Parse(time.RFC3339, createdAt.String)
	}

	return &session, nil
}

func (r *SessionRepository) scanSessionFromRows(rows *sql.Rows) (*domain.Session, error) {
	var session domain.Session
	var userID, projectID, projectPath, gitBranch, title, startedAt, endedAt, updatedAt, createdAt sql.NullString

	err := rows.Scan(&session.ID, &userID, &projectID, &session.ClaudeSessionID, &projectPath, &gitBranch, &title, &startedAt, &endedAt, &updatedAt, &createdAt)
	if err != nil {
		return nil, err
	}

	if userID.Valid {
		session.UserID = &userID.String
	}
	if projectID.Valid {
		session.ProjectID = projectID.String
	} else {
		session.ProjectID = domain.DefaultProjectID
	}
	if projectPath.Valid {
		session.ProjectPath = projectPath.String
	}
	if gitBranch.Valid {
		session.GitBranch = gitBranch.String
	}
	if title.Valid {
		session.Title = &title.String
	}
	if startedAt.Valid {
		session.StartedAt, _ = time.Parse(time.RFC3339, startedAt.String)
	}
	if endedAt.Valid {
		t, _ := time.Parse(time.RFC3339, endedAt.String)
		session.EndedAt = &t
	}
	if updatedAt.Valid {
		session.UpdatedAt, _ = time.Parse(time.RFC3339, updatedAt.String)
	}
	if createdAt.Valid {
		session.CreatedAt, _ = time.Parse(time.RFC3339, createdAt.String)
	}

	return &session, nil
}
