package postgres

import (
	"context"
	"database/sql"
	"time"

	"github.com/google/uuid"
	"github.com/satetsu888/agentrace/server/internal/domain"
)

type APIKeyRepository struct {
	db *DB
}

func NewAPIKeyRepository(db *DB) *APIKeyRepository {
	return &APIKeyRepository{db: db}
}

func (r *APIKeyRepository) Create(ctx context.Context, key *domain.APIKey) error {
	if key.ID == "" {
		key.ID = uuid.New().String()
	}
	if key.CreatedAt.IsZero() {
		key.CreatedAt = time.Now()
	}

	_, err := r.db.ExecContext(ctx,
		`INSERT INTO api_keys (id, user_id, name, key_hash, key_prefix, last_used_at, created_at)
		 VALUES ($1, $2, $3, $4, $5, $6, $7)`,
		key.ID, key.UserID, key.Name, key.KeyHash, key.KeyPrefix, key.LastUsedAt, key.CreatedAt,
	)
	return err
}

func (r *APIKeyRepository) FindByKeyHash(ctx context.Context, keyHash string) (*domain.APIKey, error) {
	return r.scanKey(r.db.QueryRowContext(ctx,
		`SELECT id, user_id, name, key_hash, key_prefix, last_used_at, created_at
		 FROM api_keys WHERE key_hash = $1`,
		keyHash,
	))
}

func (r *APIKeyRepository) FindByUserID(ctx context.Context, userID string) ([]*domain.APIKey, error) {
	rows, err := r.db.QueryContext(ctx,
		`SELECT id, user_id, name, key_hash, key_prefix, last_used_at, created_at
		 FROM api_keys WHERE user_id = $1 ORDER BY created_at DESC`,
		userID,
	)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var keys []*domain.APIKey
	for rows.Next() {
		key, err := r.scanKeyFromRows(rows)
		if err != nil {
			return nil, err
		}
		keys = append(keys, key)
	}

	return keys, rows.Err()
}

func (r *APIKeyRepository) FindByID(ctx context.Context, id string) (*domain.APIKey, error) {
	return r.scanKey(r.db.QueryRowContext(ctx,
		`SELECT id, user_id, name, key_hash, key_prefix, last_used_at, created_at
		 FROM api_keys WHERE id = $1`,
		id,
	))
}

func (r *APIKeyRepository) Delete(ctx context.Context, id string) error {
	_, err := r.db.ExecContext(ctx, `DELETE FROM api_keys WHERE id = $1`, id)
	return err
}

func (r *APIKeyRepository) UpdateLastUsedAt(ctx context.Context, id string) error {
	_, err := r.db.ExecContext(ctx,
		`UPDATE api_keys SET last_used_at = $1 WHERE id = $2`,
		time.Now(), id,
	)
	return err
}

func (r *APIKeyRepository) scanKey(row *sql.Row) (*domain.APIKey, error) {
	var key domain.APIKey
	var lastUsedAt sql.NullTime

	err := row.Scan(&key.ID, &key.UserID, &key.Name, &key.KeyHash, &key.KeyPrefix, &lastUsedAt, &key.CreatedAt)
	if err == sql.ErrNoRows {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}

	if lastUsedAt.Valid {
		key.LastUsedAt = &lastUsedAt.Time
	}

	return &key, nil
}

func (r *APIKeyRepository) scanKeyFromRows(rows *sql.Rows) (*domain.APIKey, error) {
	var key domain.APIKey
	var lastUsedAt sql.NullTime

	err := rows.Scan(&key.ID, &key.UserID, &key.Name, &key.KeyHash, &key.KeyPrefix, &lastUsedAt, &key.CreatedAt)
	if err != nil {
		return nil, err
	}

	if lastUsedAt.Valid {
		key.LastUsedAt = &lastUsedAt.Time
	}

	return &key, nil
}
