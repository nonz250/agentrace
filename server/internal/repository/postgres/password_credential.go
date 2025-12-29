package postgres

import (
	"context"
	"database/sql"
	"time"

	"github.com/google/uuid"
	"github.com/satetsu888/agentrace/server/internal/domain"
)

type PasswordCredentialRepository struct {
	db *DB
}

func NewPasswordCredentialRepository(db *DB) *PasswordCredentialRepository {
	return &PasswordCredentialRepository{db: db}
}

func (r *PasswordCredentialRepository) Create(ctx context.Context, cred *domain.PasswordCredential) error {
	if cred.ID == "" {
		cred.ID = uuid.New().String()
	}
	now := time.Now()
	if cred.CreatedAt.IsZero() {
		cred.CreatedAt = now
	}
	if cred.UpdatedAt.IsZero() {
		cred.UpdatedAt = now
	}

	_, err := r.db.ExecContext(ctx,
		`INSERT INTO password_credentials (id, user_id, password_hash, created_at, updated_at) VALUES ($1, $2, $3, $4, $5)`,
		cred.ID, cred.UserID, cred.PasswordHash, cred.CreatedAt, cred.UpdatedAt,
	)
	return err
}

func (r *PasswordCredentialRepository) FindByUserID(ctx context.Context, userID string) (*domain.PasswordCredential, error) {
	var cred domain.PasswordCredential

	err := r.db.QueryRowContext(ctx,
		`SELECT id, user_id, password_hash, created_at, updated_at FROM password_credentials WHERE user_id = $1`,
		userID,
	).Scan(&cred.ID, &cred.UserID, &cred.PasswordHash, &cred.CreatedAt, &cred.UpdatedAt)

	if err == sql.ErrNoRows {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}

	return &cred, nil
}

func (r *PasswordCredentialRepository) Update(ctx context.Context, cred *domain.PasswordCredential) error {
	cred.UpdatedAt = time.Now()
	_, err := r.db.ExecContext(ctx,
		`UPDATE password_credentials SET password_hash = $1, updated_at = $2 WHERE id = $3`,
		cred.PasswordHash, cred.UpdatedAt, cred.ID,
	)
	return err
}

func (r *PasswordCredentialRepository) Delete(ctx context.Context, id string) error {
	_, err := r.db.ExecContext(ctx, `DELETE FROM password_credentials WHERE id = $1`, id)
	return err
}
