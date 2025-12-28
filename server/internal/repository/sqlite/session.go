package sqlite

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

	var endedAt *string
	if session.EndedAt != nil {
		s := session.EndedAt.Format(time.RFC3339)
		endedAt = &s
	}

	_, err := r.db.ExecContext(ctx,
		`INSERT INTO sessions (id, user_id, claude_session_id, project_path, started_at, ended_at, created_at)
		 VALUES (?, ?, ?, ?, ?, ?, ?)`,
		session.ID, session.UserID, session.ClaudeSessionID, session.ProjectPath,
		session.StartedAt.Format(time.RFC3339), endedAt, session.CreatedAt.Format(time.RFC3339),
	)
	return err
}

func (r *SessionRepository) FindByID(ctx context.Context, id string) (*domain.Session, error) {
	return r.scanSession(r.db.QueryRowContext(ctx,
		`SELECT id, user_id, claude_session_id, project_path, started_at, ended_at, created_at
		 FROM sessions WHERE id = ?`,
		id,
	))
}

func (r *SessionRepository) FindAll(ctx context.Context, limit int, offset int) ([]*domain.Session, error) {
	query := `SELECT id, user_id, claude_session_id, project_path, started_at, ended_at, created_at
		 FROM sessions ORDER BY started_at DESC`

	if limit > 0 {
		query += ` LIMIT ? OFFSET ?`
	}

	var rows *sql.Rows
	var err error

	if limit > 0 {
		rows, err = r.db.QueryContext(ctx, query, limit, offset)
	} else {
		rows, err = r.db.QueryContext(ctx, query)
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
		`SELECT id, user_id, claude_session_id, project_path, started_at, ended_at, created_at
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

func (r *SessionRepository) scanSession(row *sql.Row) (*domain.Session, error) {
	var session domain.Session
	var userID, projectPath, startedAt, endedAt, createdAt sql.NullString

	err := row.Scan(&session.ID, &userID, &session.ClaudeSessionID, &projectPath, &startedAt, &endedAt, &createdAt)
	if err == sql.ErrNoRows {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}

	if userID.Valid {
		session.UserID = &userID.String
	}
	if projectPath.Valid {
		session.ProjectPath = projectPath.String
	}
	if startedAt.Valid {
		session.StartedAt, _ = time.Parse(time.RFC3339, startedAt.String)
	}
	if endedAt.Valid {
		t, _ := time.Parse(time.RFC3339, endedAt.String)
		session.EndedAt = &t
	}
	if createdAt.Valid {
		session.CreatedAt, _ = time.Parse(time.RFC3339, createdAt.String)
	}

	return &session, nil
}

func (r *SessionRepository) scanSessionFromRows(rows *sql.Rows) (*domain.Session, error) {
	var session domain.Session
	var userID, projectPath, startedAt, endedAt, createdAt sql.NullString

	err := rows.Scan(&session.ID, &userID, &session.ClaudeSessionID, &projectPath, &startedAt, &endedAt, &createdAt)
	if err != nil {
		return nil, err
	}

	if userID.Valid {
		session.UserID = &userID.String
	}
	if projectPath.Valid {
		session.ProjectPath = projectPath.String
	}
	if startedAt.Valid {
		session.StartedAt, _ = time.Parse(time.RFC3339, startedAt.String)
	}
	if endedAt.Valid {
		t, _ := time.Parse(time.RFC3339, endedAt.String)
		session.EndedAt = &t
	}
	if createdAt.Valid {
		session.CreatedAt, _ = time.Parse(time.RFC3339, createdAt.String)
	}

	return &session, nil
}
