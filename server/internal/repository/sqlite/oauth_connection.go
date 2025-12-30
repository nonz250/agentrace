package sqlite

import (
	"context"
	"database/sql"
	"time"

	"github.com/google/uuid"
	"github.com/satetsu888/agentrace/server/internal/domain"
)

type OAuthConnectionRepository struct {
	db *DB
}

func NewOAuthConnectionRepository(db *DB) *OAuthConnectionRepository {
	return &OAuthConnectionRepository{db: db}
}

func (r *OAuthConnectionRepository) Create(ctx context.Context, conn *domain.OAuthConnection) error {
	if conn.ID == "" {
		conn.ID = uuid.New().String()
	}
	if conn.CreatedAt.IsZero() {
		conn.CreatedAt = time.Now()
	}

	_, err := r.db.ExecContext(ctx,
		`INSERT INTO oauth_connections (id, user_id, provider, provider_id, created_at) VALUES (?, ?, ?, ?, ?)`,
		conn.ID, conn.UserID, conn.Provider, conn.ProviderID, conn.CreatedAt.Format(time.RFC3339),
	)
	return err
}

func (r *OAuthConnectionRepository) FindByProviderAndProviderID(ctx context.Context, provider, providerID string) (*domain.OAuthConnection, error) {
	var conn domain.OAuthConnection
	var createdAt string

	err := r.db.QueryRowContext(ctx,
		`SELECT id, user_id, provider, provider_id, created_at FROM oauth_connections WHERE provider = ? AND provider_id = ?`,
		provider, providerID,
	).Scan(&conn.ID, &conn.UserID, &conn.Provider, &conn.ProviderID, &createdAt)

	if err == sql.ErrNoRows {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}

	conn.CreatedAt, _ = time.Parse(time.RFC3339, createdAt)
	return &conn, nil
}

func (r *OAuthConnectionRepository) FindByUserID(ctx context.Context, userID string) ([]*domain.OAuthConnection, error) {
	rows, err := r.db.QueryContext(ctx,
		`SELECT id, user_id, provider, provider_id, created_at FROM oauth_connections WHERE user_id = ?`,
		userID,
	)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var connections []*domain.OAuthConnection
	for rows.Next() {
		var conn domain.OAuthConnection
		var createdAt string

		if err := rows.Scan(&conn.ID, &conn.UserID, &conn.Provider, &conn.ProviderID, &createdAt); err != nil {
			return nil, err
		}

		conn.CreatedAt, _ = time.Parse(time.RFC3339, createdAt)
		connections = append(connections, &conn)
	}

	return connections, rows.Err()
}

func (r *OAuthConnectionRepository) Delete(ctx context.Context, id string) error {
	_, err := r.db.ExecContext(ctx, `DELETE FROM oauth_connections WHERE id = ?`, id)
	return err
}
