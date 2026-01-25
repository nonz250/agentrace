package postgres

import (
	"context"
	"database/sql"
	"time"

	"github.com/google/uuid"
	"github.com/satetsu888/agentrace/server/internal/domain"
)

type PlanCommentMessageRepository struct {
	db *DB
}

func NewPlanCommentMessageRepository(db *DB) *PlanCommentMessageRepository {
	return &PlanCommentMessageRepository{db: db}
}

func (r *PlanCommentMessageRepository) Create(ctx context.Context, message *domain.PlanCommentMessage) error {
	if message.ID == "" {
		message.ID = uuid.New().String()
	}
	now := time.Now()
	if message.CreatedAt.IsZero() {
		message.CreatedAt = now
	}
	if message.UpdatedAt.IsZero() {
		message.UpdatedAt = now
	}

	_, err := r.db.ExecContext(ctx,
		`INSERT INTO plan_comment_messages (id, thread_id, user_id, content, created_at, updated_at)
		 VALUES ($1, $2, $3, $4, $5, $6)`,
		message.ID, message.ThreadID, message.UserID,
		message.Content,
		message.CreatedAt, message.UpdatedAt,
	)
	return err
}

func (r *PlanCommentMessageRepository) FindByID(ctx context.Context, id string) (*domain.PlanCommentMessage, error) {
	return r.scanMessage(r.db.QueryRowContext(ctx,
		`SELECT id, thread_id, user_id, content, created_at, updated_at
		 FROM plan_comment_messages WHERE id = $1`,
		id,
	))
}

func (r *PlanCommentMessageRepository) FindByThreadID(ctx context.Context, threadID string) ([]*domain.PlanCommentMessage, error) {
	rows, err := r.db.QueryContext(ctx,
		`SELECT id, thread_id, user_id, content, created_at, updated_at
		 FROM plan_comment_messages WHERE thread_id = $1
		 ORDER BY created_at ASC`,
		threadID,
	)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var messages []*domain.PlanCommentMessage
	for rows.Next() {
		message, err := r.scanMessageFromRows(rows)
		if err != nil {
			return nil, err
		}
		messages = append(messages, message)
	}

	if err := rows.Err(); err != nil {
		return nil, err
	}

	return messages, nil
}

func (r *PlanCommentMessageRepository) Update(ctx context.Context, message *domain.PlanCommentMessage) error {
	message.UpdatedAt = time.Now()

	_, err := r.db.ExecContext(ctx,
		`UPDATE plan_comment_messages SET content = $1, updated_at = $2
		 WHERE id = $3`,
		message.Content, message.UpdatedAt, message.ID,
	)
	return err
}

func (r *PlanCommentMessageRepository) Delete(ctx context.Context, id string) error {
	_, err := r.db.ExecContext(ctx,
		`DELETE FROM plan_comment_messages WHERE id = $1`,
		id,
	)
	return err
}

func (r *PlanCommentMessageRepository) DeleteByThreadID(ctx context.Context, threadID string) error {
	_, err := r.db.ExecContext(ctx,
		`DELETE FROM plan_comment_messages WHERE thread_id = $1`,
		threadID,
	)
	return err
}

func (r *PlanCommentMessageRepository) scanMessage(row *sql.Row) (*domain.PlanCommentMessage, error) {
	var message domain.PlanCommentMessage
	var createdAt, updatedAt sql.NullTime

	err := row.Scan(
		&message.ID, &message.ThreadID, &message.UserID,
		&message.Content, &createdAt, &updatedAt,
	)
	if err == sql.ErrNoRows {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}

	if createdAt.Valid {
		message.CreatedAt = createdAt.Time
	}
	if updatedAt.Valid {
		message.UpdatedAt = updatedAt.Time
	}

	return &message, nil
}

func (r *PlanCommentMessageRepository) scanMessageFromRows(rows *sql.Rows) (*domain.PlanCommentMessage, error) {
	var message domain.PlanCommentMessage
	var createdAt, updatedAt sql.NullTime

	err := rows.Scan(
		&message.ID, &message.ThreadID, &message.UserID,
		&message.Content, &createdAt, &updatedAt,
	)
	if err != nil {
		return nil, err
	}

	if createdAt.Valid {
		message.CreatedAt = createdAt.Time
	}
	if updatedAt.Valid {
		message.UpdatedAt = updatedAt.Time
	}

	return &message, nil
}
