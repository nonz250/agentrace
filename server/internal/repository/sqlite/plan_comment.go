package sqlite

import (
	"context"
	"database/sql"
	"time"

	"github.com/google/uuid"
	"github.com/satetsu888/agentrace/server/internal/domain"
)

type PlanCommentRepository struct {
	db *DB
}

func NewPlanCommentRepository(db *DB) *PlanCommentRepository {
	return &PlanCommentRepository{db: db}
}

func (r *PlanCommentRepository) Create(ctx context.Context, comment *domain.PlanComment) error {
	if comment.ID == "" {
		comment.ID = uuid.New().String()
	}
	now := time.Now()
	if comment.CreatedAt.IsZero() {
		comment.CreatedAt = now
	}
	if comment.UpdatedAt.IsZero() {
		comment.UpdatedAt = now
	}
	if comment.Status == "" {
		comment.Status = domain.PlanCommentStatusActive
	}

	_, err := r.db.ExecContext(ctx,
		`INSERT INTO plan_comments (id, plan_document_id, user_id, target_text, context_before, context_after, original_body_hash, content, status, created_at, updated_at)
		 VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?)`,
		comment.ID, comment.PlanDocumentID, comment.UserID,
		comment.TargetText, comment.ContextBefore, comment.ContextAfter, comment.OriginalBodyHash,
		comment.Content, string(comment.Status),
		comment.CreatedAt.Format(time.RFC3339Nano), comment.UpdatedAt.Format(time.RFC3339Nano),
	)
	return err
}

func (r *PlanCommentRepository) FindByID(ctx context.Context, id string) (*domain.PlanComment, error) {
	return r.scanComment(r.db.QueryRowContext(ctx,
		`SELECT id, plan_document_id, user_id, target_text, context_before, context_after, original_body_hash, content, status, created_at, updated_at
		 FROM plan_comments WHERE id = ?`,
		id,
	))
}

func (r *PlanCommentRepository) FindByPlanDocumentID(ctx context.Context, planDocumentID string, status *domain.PlanCommentStatus) ([]*domain.PlanComment, error) {
	query := `SELECT id, plan_document_id, user_id, target_text, context_before, context_after, original_body_hash, content, status, created_at, updated_at
		 FROM plan_comments WHERE plan_document_id = ?`
	args := []any{planDocumentID}

	if status != nil {
		query += " AND status = ?"
		args = append(args, string(*status))
	}

	query += " ORDER BY created_at ASC"

	rows, err := r.db.QueryContext(ctx, query, args...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var comments []*domain.PlanComment
	for rows.Next() {
		comment, err := r.scanCommentFromRows(rows)
		if err != nil {
			return nil, err
		}
		comments = append(comments, comment)
	}

	if err := rows.Err(); err != nil {
		return nil, err
	}

	return comments, nil
}

func (r *PlanCommentRepository) Update(ctx context.Context, comment *domain.PlanComment) error {
	comment.UpdatedAt = time.Now()

	_, err := r.db.ExecContext(ctx,
		`UPDATE plan_comments SET content = ?, status = ?, updated_at = ?
		 WHERE id = ?`,
		comment.Content, string(comment.Status), comment.UpdatedAt.Format(time.RFC3339Nano), comment.ID,
	)
	return err
}

func (r *PlanCommentRepository) Delete(ctx context.Context, id string) error {
	_, err := r.db.ExecContext(ctx,
		`DELETE FROM plan_comments WHERE id = ?`,
		id,
	)
	return err
}

func (r *PlanCommentRepository) MarkOutdatedByPlanDocumentID(ctx context.Context, planDocumentID string) error {
	_, err := r.db.ExecContext(ctx,
		`UPDATE plan_comments SET status = ?, updated_at = ?
		 WHERE plan_document_id = ? AND status = ?`,
		string(domain.PlanCommentStatusOutdated), time.Now().Format(time.RFC3339Nano),
		planDocumentID, string(domain.PlanCommentStatusActive),
	)
	return err
}

func (r *PlanCommentRepository) scanComment(row *sql.Row) (*domain.PlanComment, error) {
	var comment domain.PlanComment
	var status string
	var createdAt, updatedAt string

	err := row.Scan(
		&comment.ID, &comment.PlanDocumentID, &comment.UserID,
		&comment.TargetText, &comment.ContextBefore, &comment.ContextAfter, &comment.OriginalBodyHash,
		&comment.Content, &status, &createdAt, &updatedAt,
	)
	if err == sql.ErrNoRows {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}

	comment.Status = domain.PlanCommentStatus(status)
	comment.CreatedAt, _ = time.Parse(time.RFC3339Nano, createdAt)
	comment.UpdatedAt, _ = time.Parse(time.RFC3339Nano, updatedAt)

	return &comment, nil
}

func (r *PlanCommentRepository) scanCommentFromRows(rows *sql.Rows) (*domain.PlanComment, error) {
	var comment domain.PlanComment
	var status string
	var createdAt, updatedAt string

	err := rows.Scan(
		&comment.ID, &comment.PlanDocumentID, &comment.UserID,
		&comment.TargetText, &comment.ContextBefore, &comment.ContextAfter, &comment.OriginalBodyHash,
		&comment.Content, &status, &createdAt, &updatedAt,
	)
	if err != nil {
		return nil, err
	}

	comment.Status = domain.PlanCommentStatus(status)
	comment.CreatedAt, _ = time.Parse(time.RFC3339Nano, createdAt)
	comment.UpdatedAt, _ = time.Parse(time.RFC3339Nano, updatedAt)

	return &comment, nil
}
