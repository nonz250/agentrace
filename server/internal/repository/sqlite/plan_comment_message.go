package sqlite

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
		 VALUES (?, ?, ?, ?, ?, ?)`,
		message.ID, message.ThreadID, message.UserID,
		message.Content,
		message.CreatedAt.Format(time.RFC3339Nano), message.UpdatedAt.Format(time.RFC3339Nano),
	)
	return err
}

func (r *PlanCommentMessageRepository) FindByID(ctx context.Context, id string) (*domain.PlanCommentMessage, error) {
	return r.scanMessage(r.db.QueryRowContext(ctx,
		`SELECT id, thread_id, user_id, content, created_at, updated_at
		 FROM plan_comment_messages WHERE id = ?`,
		id,
	))
}

func (r *PlanCommentMessageRepository) FindByThreadID(ctx context.Context, threadID string) ([]*domain.PlanCommentMessage, error) {
	rows, err := r.db.QueryContext(ctx,
		`SELECT id, thread_id, user_id, content, created_at, updated_at
		 FROM plan_comment_messages WHERE thread_id = ?
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
		`UPDATE plan_comment_messages SET content = ?, updated_at = ?
		 WHERE id = ?`,
		message.Content, message.UpdatedAt.Format(time.RFC3339Nano), message.ID,
	)
	return err
}

func (r *PlanCommentMessageRepository) Delete(ctx context.Context, id string) error {
	_, err := r.db.ExecContext(ctx,
		`DELETE FROM plan_comment_messages WHERE id = ?`,
		id,
	)
	return err
}

func (r *PlanCommentMessageRepository) DeleteByThreadID(ctx context.Context, threadID string) error {
	_, err := r.db.ExecContext(ctx,
		`DELETE FROM plan_comment_messages WHERE thread_id = ?`,
		threadID,
	)
	return err
}

func (r *PlanCommentMessageRepository) scanMessage(row *sql.Row) (*domain.PlanCommentMessage, error) {
	var message domain.PlanCommentMessage
	var createdAt, updatedAt string

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

	message.CreatedAt, _ = time.Parse(time.RFC3339Nano, createdAt)
	message.UpdatedAt, _ = time.Parse(time.RFC3339Nano, updatedAt)

	return &message, nil
}

func (r *PlanCommentMessageRepository) scanMessageFromRows(rows *sql.Rows) (*domain.PlanCommentMessage, error) {
	var message domain.PlanCommentMessage
	var createdAt, updatedAt string

	err := rows.Scan(
		&message.ID, &message.ThreadID, &message.UserID,
		&message.Content, &createdAt, &updatedAt,
	)
	if err != nil {
		return nil, err
	}

	message.CreatedAt, _ = time.Parse(time.RFC3339Nano, createdAt)
	message.UpdatedAt, _ = time.Parse(time.RFC3339Nano, updatedAt)

	return &message, nil
}
