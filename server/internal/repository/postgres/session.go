package postgres

import (
	"context"
	"database/sql"
	"fmt"
	"strings"
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

	_, err := r.db.ExecContext(ctx,
		`INSERT INTO sessions (id, user_id, project_id, claude_session_id, project_path, git_branch, title, started_at, ended_at, updated_at, created_at, parent_session_id, agent_id, is_sidechain)
		 VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12, $13, $14)`,
		session.ID, session.UserID, session.ProjectID, session.ClaudeSessionID, session.ProjectPath,
		session.GitBranch, session.Title,
		session.StartedAt, session.EndedAt, session.UpdatedAt, session.CreatedAt,
		session.ParentSessionID, session.AgentID, session.IsSidechain,
	)
	return err
}

func (r *SessionRepository) FindByID(ctx context.Context, id string) (*domain.Session, error) {
	return r.scanSession(r.db.QueryRowContext(ctx,
		`SELECT id, user_id, project_id, claude_session_id, project_path, git_branch, title, started_at, ended_at, updated_at, created_at, parent_session_id, agent_id, is_sidechain
		 FROM sessions WHERE id = $1`,
		id,
	))
}

func (r *SessionRepository) FindByClaudeSessionID(ctx context.Context, claudeSessionID string) (*domain.Session, error) {
	return r.scanSession(r.db.QueryRowContext(ctx,
		`SELECT id, user_id, project_id, claude_session_id, project_path, git_branch, title, started_at, ended_at, updated_at, created_at, parent_session_id, agent_id, is_sidechain
		 FROM sessions WHERE claude_session_id = $1`,
		claudeSessionID,
	))
}

func (r *SessionRepository) Find(ctx context.Context, query domain.SessionQuery) ([]*domain.Session, string, error) {
	// Validate sortBy to prevent SQL injection
	orderColumn := "updated_at"
	if query.SortBy == "created_at" {
		orderColumn = "created_at"
	}

	baseQuery := `SELECT id, user_id, project_id, claude_session_id, project_path, git_branch, title, started_at, ended_at, updated_at, created_at, parent_session_id, agent_id, is_sidechain
		 FROM sessions WHERE is_sidechain = FALSE`

	var args []any
	paramIdx := 1

	// Filter by project ID
	if query.ProjectID != "" {
		baseQuery += fmt.Sprintf(` AND project_id = $%d`, paramIdx)
		args = append(args, query.ProjectID)
		paramIdx++
	}

	// Filter by user IDs
	if len(query.UserIDs) > 0 {
		placeholders := make([]string, len(query.UserIDs))
		for i, uid := range query.UserIDs {
			placeholders[i] = fmt.Sprintf("$%d", paramIdx)
			args = append(args, uid)
			paramIdx++
		}
		baseQuery += ` AND user_id IN (` + strings.Join(placeholders, ",") + `)`
	}

	// Apply cursor filter
	if query.Cursor != "" {
		cursorInfo := repository.DecodeCursor(query.Cursor)
		if cursorInfo != nil {
			cursorTime, err := cursorInfo.ParseSortTime()
			if err == nil {
				baseQuery += fmt.Sprintf(` AND (%s < $%d OR (%s = $%d AND id < $%d))`, orderColumn, paramIdx, orderColumn, paramIdx+1, paramIdx+2)
				args = append(args, cursorTime, cursorTime, cursorInfo.ID)
				paramIdx += 3
			}
		}
	}

	baseQuery += ` ORDER BY ` + orderColumn + ` DESC, id DESC`

	limit := query.Limit
	if limit <= 0 {
		limit = 100
	}
	baseQuery += fmt.Sprintf(` LIMIT $%d`, paramIdx)
	args = append(args, limit+1)

	rows, err := r.db.QueryContext(ctx, baseQuery, args...)
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
	if len(sessions) > limit {
		sessions = sessions[:limit]
		lastItem := sessions[limit-1]
		var sortTime time.Time
		if query.SortBy == "created_at" {
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
		`SELECT id, user_id, project_id, claude_session_id, project_path, git_branch, title, started_at, ended_at, updated_at, created_at, parent_session_id, agent_id, is_sidechain
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
	var parentSessionID, agentID sql.NullString
	var isSidechain sql.NullBool

	err := row.Scan(&session.ID, &userID, &projectID, &session.ClaudeSessionID, &projectPath, &gitBranch, &title, &startedAt, &endedAt, &updatedAt, &createdAt, &parentSessionID, &agentID, &isSidechain)
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
	if parentSessionID.Valid {
		session.ParentSessionID = &parentSessionID.String
	}
	if agentID.Valid {
		session.AgentID = &agentID.String
	}
	if isSidechain.Valid {
		session.IsSidechain = isSidechain.Bool
	}

	return &session, nil
}

func (r *SessionRepository) scanSessionFromRows(rows *sql.Rows) (*domain.Session, error) {
	var session domain.Session
	var userID, projectID, projectPath, gitBranch, title sql.NullString
	var startedAt, endedAt, updatedAt, createdAt sql.NullTime
	var parentSessionID, agentID sql.NullString
	var isSidechain sql.NullBool

	err := rows.Scan(&session.ID, &userID, &projectID, &session.ClaudeSessionID, &projectPath, &gitBranch, &title, &startedAt, &endedAt, &updatedAt, &createdAt, &parentSessionID, &agentID, &isSidechain)
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
	if parentSessionID.Valid {
		session.ParentSessionID = &parentSessionID.String
	}
	if agentID.Valid {
		session.AgentID = &agentID.String
	}
	if isSidechain.Valid {
		session.IsSidechain = isSidechain.Bool
	}

	return &session, nil
}

func (r *SessionRepository) FindSubagentsByParentID(ctx context.Context, parentID string) ([]*domain.Session, error) {
	query := `SELECT id, user_id, project_id, claude_session_id, project_path, git_branch, title, started_at, ended_at, updated_at, created_at, parent_session_id, agent_id, is_sidechain
		 FROM sessions WHERE parent_session_id = $1
		 ORDER BY created_at ASC`

	rows, err := r.db.QueryContext(ctx, query, parentID)
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

	if err := rows.Err(); err != nil {
		return nil, err
	}

	return sessions, nil
}

func (r *SessionRepository) Delete(ctx context.Context, id string) error {
	_, err := r.db.ExecContext(ctx, `DELETE FROM sessions WHERE id = $1`, id)
	return err
}
