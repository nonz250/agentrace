package sqlite

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
		`INSERT INTO password_credentials (id, user_id, password_hash, created_at, updated_at) VALUES (?, ?, ?, ?, ?)`,
		cred.ID, cred.UserID, cred.PasswordHash, cred.CreatedAt.Format(time.RFC3339), cred.UpdatedAt.Format(time.RFC3339),
	)
	return err
}

func (r *PasswordCredentialRepository) FindByUserID(ctx context.Context, userID string) (*domain.PasswordCredential, error) {
	var cred domain.PasswordCredential
	var createdAt, updatedAt string

	err := r.db.QueryRowContext(ctx,
		`SELECT id, user_id, password_hash, created_at, updated_at FROM password_credentials WHERE user_id = ?`,
		userID,
	).Scan(&cred.ID, &cred.UserID, &cred.PasswordHash, &createdAt, &updatedAt)

	if err == sql.ErrNoRows {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}

	cred.CreatedAt, _ = time.Parse(time.RFC3339, createdAt)
	cred.UpdatedAt, _ = time.Parse(time.RFC3339, updatedAt)
	return &cred, nil
}

func (r *PasswordCredentialRepository) Update(ctx context.Context, cred *domain.PasswordCredential) error {
	cred.UpdatedAt = time.Now()
	_, err := r.db.ExecContext(ctx,
		`UPDATE password_credentials SET password_hash = ?, updated_at = ? WHERE id = ?`,
		cred.PasswordHash, cred.UpdatedAt.Format(time.RFC3339), cred.ID,
	)
	return err
}

func (r *PasswordCredentialRepository) Delete(ctx context.Context, id string) error {
	_, err := r.db.ExecContext(ctx, `DELETE FROM password_credentials WHERE id = ?`, id)
	return err
}
