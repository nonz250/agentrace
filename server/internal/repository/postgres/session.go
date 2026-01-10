package postgres

import (
	"context"
	"database/sql"
	"time"

	"github.com/google/uuid"
	"github.com/satetsu888/agentrace/server/internal/domain"
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

	_, err := r.db.ExecContext(ctx,
		`INSERT INTO sessions (id, user_id, project_id, claude_session_id, project_path, git_branch, title, started_at, ended_at, updated_at, created_at)
		 VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11)`,
		session.ID, session.UserID, session.ProjectID, session.ClaudeSessionID, session.ProjectPath,
		session.GitBranch, session.Title,
		session.StartedAt, session.EndedAt, session.UpdatedAt, session.CreatedAt,
	)
	return err
}

func (r *SessionRepository) FindByID(ctx context.Context, id string) (*domain.Session, error) {
	return r.scanSession(r.db.QueryRowContext(ctx,
		`SELECT id, user_id, project_id, claude_session_id, project_path, git_branch, title, started_at, ended_at, updated_at, created_at
		 FROM sessions WHERE id = $1`,
		id,
	))
}

func (r *SessionRepository) FindByClaudeSessionID(ctx context.Context, claudeSessionID string) (*domain.Session, error) {
	return r.scanSession(r.db.QueryRowContext(ctx,
		`SELECT id, user_id, project_id, claude_session_id, project_path, git_branch, title, started_at, ended_at, updated_at, created_at
		 FROM sessions WHERE claude_session_id = $1`,
		claudeSessionID,
	))
}

func (r *SessionRepository) FindAll(ctx context.Context, limit int, offset int, sortBy string) ([]*domain.Session, error) {
	// Validate sortBy to prevent SQL injection
	orderColumn := "updated_at"
	if sortBy == "created_at" {
		orderColumn = "created_at"
	}

	var rows *sql.Rows
	var err error

	if limit > 0 {
		rows, err = r.db.QueryContext(ctx,
			`SELECT id, user_id, project_id, claude_session_id, project_path, git_branch, title, started_at, ended_at, updated_at, created_at
			 FROM sessions ORDER BY `+orderColumn+` DESC LIMIT $1 OFFSET $2`,
			limit, offset,
		)
	} else {
		rows, err = r.db.QueryContext(ctx,
			`SELECT id, user_id, project_id, claude_session_id, project_path, git_branch, title, started_at, ended_at, updated_at, created_at
			 FROM sessions ORDER BY `+orderColumn+` DESC`,
		)
	}

	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var sessions []*domain.Session
	for rows.Next() {
		session, err := r.scanSessionFromRows(rows)
		if err != nil {
			return nil, err
		}
		sessions = append(sessions, session)
	}

	return sessions, rows.Err()
}

func (r *SessionRepository) FindByProjectID(ctx context.Context, projectID string, limit int, offset int, sortBy string) ([]*domain.Session, error) {
	// Validate sortBy to prevent SQL injection
	orderColumn := "updated_at"
	if sortBy == "created_at" {
		orderColumn = "created_at"
	}

	var rows *sql.Rows
	var err error

	if limit > 0 {
		rows, err = r.db.QueryContext(ctx,
			`SELECT id, user_id, project_id, claude_session_id, project_path, git_branch, title, started_at, ended_at, updated_at, created_at
			 FROM sessions WHERE project_id = $1 ORDER BY `+orderColumn+` DESC LIMIT $2 OFFSET $3`,
			projectID, limit, offset,
		)
	} else {
		rows, err = r.db.QueryContext(ctx,
			`SELECT id, user_id, project_id, claude_session_id, project_path, git_branch, title, started_at, ended_at, updated_at, created_at
			 FROM sessions WHERE project_id = $1 ORDER BY `+orderColumn+` DESC`,
			projectID,
		)
	}

	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var sessions []*domain.Session
	for rows.Next() {
		session, err := r.scanSessionFromRows(rows)
		if err != nil {
			return nil, err
		}
		sessions = append(sessions, session)
	}

	return sessions, rows.Err()
}

func (r *SessionRepository) FindOrCreateByClaudeSessionID(ctx context.Context, claudeSessionID string, userID *string) (*domain.Session, error) {
	// First try to find existing session
	session, err := r.scanSession(r.db.QueryRowContext(ctx,
		`SELECT id, user_id, project_id, claude_session_id, project_path, git_branch, title, started_at, ended_at, updated_at, created_at
		 FROM sessions WHERE claude_session_id = $1`,
		claudeSessionID,
	))

	if err != nil {
		return nil, err
	}

	if session != nil {
		// Update UserID if provided and not already set
		if userID != nil && session.UserID == nil {
			_, err := r.db.ExecContext(ctx,
				`UPDATE sessions SET user_id = $1 WHERE id = $2`,
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
		`UPDATE sessions SET user_id = $1 WHERE id = $2`,
		userID, id,
	)
	return err
}

func (r *SessionRepository) UpdateProjectPath(ctx context.Context, id string, projectPath string) error {
	_, err := r.db.ExecContext(ctx,
		`UPDATE sessions SET project_path = $1 WHERE id = $2`,
		projectPath, id,
	)
	return err
}

func (r *SessionRepository) UpdateProjectID(ctx context.Context, id string, projectID string) error {
	_, err := r.db.ExecContext(ctx,
		`UPDATE sessions SET project_id = $1 WHERE id = $2`,
		projectID, id,
	)
	return err
}

func (r *SessionRepository) UpdateGitBranch(ctx context.Context, id string, gitBranch string) error {
	_, err := r.db.ExecContext(ctx,
		`UPDATE sessions SET git_branch = $1 WHERE id = $2`,
		gitBranch, id,
	)
	return err
}

func (r *SessionRepository) UpdateTitle(ctx context.Context, id string, title string) error {
	_, err := r.db.ExecContext(ctx,
		`UPDATE sessions SET title = $1 WHERE id = $2`,
		title, id,
	)
	return err
}

func (r *SessionRepository) UpdateUpdatedAt(ctx context.Context, id string, updatedAt time.Time) error {
	_, err := r.db.ExecContext(ctx,
		`UPDATE sessions SET updated_at = $1 WHERE id = $2`,
		updatedAt, id,
	)
	return err
}

func (r *SessionRepository) scanSession(row *sql.Row) (*domain.Session, error) {
	var session domain.Session
	var userID, projectID, projectPath, gitBranch, title sql.NullString
	var startedAt, endedAt, updatedAt, createdAt sql.NullTime

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
		session.StartedAt = startedAt.Time
	}
	if endedAt.Valid {
		session.EndedAt = &endedAt.Time
	}
	if updatedAt.Valid {
		session.UpdatedAt = updatedAt.Time
	}
	if createdAt.Valid {
		session.CreatedAt = createdAt.Time
	}

	return &session, nil
}

func (r *SessionRepository) scanSessionFromRows(rows *sql.Rows) (*domain.Session, error) {
	var session domain.Session
	var userID, projectID, projectPath, gitBranch, title sql.NullString
	var startedAt, endedAt, updatedAt, createdAt sql.NullTime

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
		session.StartedAt = startedAt.Time
	}
	if endedAt.Valid {
		session.EndedAt = &endedAt.Time
	}
	if updatedAt.Valid {
		session.UpdatedAt = updatedAt.Time
	}
	if createdAt.Valid {
		session.CreatedAt = createdAt.Time
	}

	return &session, nil
}
