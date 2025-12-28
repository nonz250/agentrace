package sqlite

import (
	"context"
	"database/sql"
	"time"

	"github.com/google/uuid"
	"github.com/satetsu888/agentrace/server/internal/domain"
)

type UserRepository struct {
	db *DB
}

func NewUserRepository(db *DB) *UserRepository {
	return &UserRepository{db: db}
}

func (r *UserRepository) Create(ctx context.Context, user *domain.User) error {
	if user.ID == "" {
		user.ID = uuid.New().String()
	}
	if user.CreatedAt.IsZero() {
		user.CreatedAt = time.Now()
	}

	_, err := r.db.ExecContext(ctx,
		`INSERT INTO users (id, name, created_at) VALUES (?, ?, ?)`,
		user.ID, user.Name, user.CreatedAt.Format(time.RFC3339),
	)
	return err
}

func (r *UserRepository) FindByID(ctx context.Context, id string) (*domain.User, error) {
	var user domain.User
	var createdAt string

	err := r.db.QueryRowContext(ctx,
		`SELECT id, name, created_at FROM users WHERE id = ?`,
		id,
	).Scan(&user.ID, &user.Name, &createdAt)

	if err == sql.ErrNoRows {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}

	user.CreatedAt, _ = time.Parse(time.RFC3339, createdAt)
	return &user, nil
}

func (r *UserRepository) FindAll(ctx context.Context) ([]*domain.User, error) {
	rows, err := r.db.QueryContext(ctx, `SELECT id, name, created_at FROM users ORDER BY created_at DESC`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var users []*domain.User
	for rows.Next() {
		var user domain.User
		var createdAt string

		if err := rows.Scan(&user.ID, &user.Name, &createdAt); err != nil {
			return nil, err
		}

		user.CreatedAt, _ = time.Parse(time.RFC3339, createdAt)
		users = append(users, &user)
	}

	return users, rows.Err()
}
