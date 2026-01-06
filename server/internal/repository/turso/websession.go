package turso

import (
	"context"
	"database/sql"
	"time"

	"github.com/google/uuid"
	"github.com/satetsu888/agentrace/server/internal/domain"
)

type WebSessionRepository struct {
	db *DB
}

func NewWebSessionRepository(db *DB) *WebSessionRepository {
	return &WebSessionRepository{db: db}
}

func (r *WebSessionRepository) Create(ctx context.Context, session *domain.WebSession) error {
	if session.ID == "" {
		session.ID = uuid.New().String()
	}
	if session.CreatedAt.IsZero() {
		session.CreatedAt = time.Now()
	}

	_, err := r.db.ExecContext(ctx,
		`INSERT INTO web_sessions (id, user_id, token, expires_at, created_at)
		 VALUES (?, ?, ?, ?, ?)`,
		session.ID, session.UserID, session.Token,
		session.ExpiresAt.Format(time.RFC3339), session.CreatedAt.Format(time.RFC3339),
	)
	return err
}

func (r *WebSessionRepository) FindByToken(ctx context.Context, token string) (*domain.WebSession, error) {
	var session domain.WebSession
	var expiresAt, createdAt string

	err := r.db.QueryRowContext(ctx,
		`SELECT id, user_id, token, expires_at, created_at
		 FROM web_sessions WHERE token = ?`,
		token,
	).Scan(&session.ID, &session.UserID, &session.Token, &expiresAt, &createdAt)

	if err == sql.ErrNoRows {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}

	session.ExpiresAt, _ = time.Parse(time.RFC3339, expiresAt)
	session.CreatedAt, _ = time.Parse(time.RFC3339, createdAt)

	return &session, nil
}

func (r *WebSessionRepository) Delete(ctx context.Context, id string) error {
	_, err := r.db.ExecContext(ctx, `DELETE FROM web_sessions WHERE id = ?`, id)
	return err
}

func (r *WebSessionRepository) DeleteExpired(ctx context.Context) error {
	_, err := r.db.ExecContext(ctx,
		`DELETE FROM web_sessions WHERE expires_at < ?`,
		time.Now().Format(time.RFC3339),
	)
	return err
}
