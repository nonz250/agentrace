package postgres

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
		 VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11)`,
		comment.ID, comment.PlanDocumentID, comment.UserID,
		comment.TargetText, comment.ContextBefore, comment.ContextAfter, comment.OriginalBodyHash,
		comment.Content, string(comment.Status),
		comment.CreatedAt, comment.UpdatedAt,
	)
	return err
}

func (r *PlanCommentRepository) FindByID(ctx context.Context, id string) (*domain.PlanComment, error) {
	return r.scanComment(r.db.QueryRowContext(ctx,
		`SELECT id, plan_document_id, user_id, target_text, context_before, context_after, original_body_hash, content, status, created_at, updated_at
		 FROM plan_comments WHERE id = $1`,
		id,
	))
}

func (r *PlanCommentRepository) FindByPlanDocumentID(ctx context.Context, planDocumentID string, status *domain.PlanCommentStatus) ([]*domain.PlanComment, error) {
	query := `SELECT id, plan_document_id, user_id, target_text, context_before, context_after, original_body_hash, content, status, created_at, updated_at
		 FROM plan_comments WHERE plan_document_id = $1`
	args := []any{planDocumentID}

	if status != nil {
		query += " AND status = $2"
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
		`UPDATE plan_comments SET content = $1, status = $2, updated_at = $3
		 WHERE id = $4`,
		comment.Content, string(comment.Status), comment.UpdatedAt, comment.ID,
	)
	return err
}

func (r *PlanCommentRepository) Delete(ctx context.Context, id string) error {
	_, err := r.db.ExecContext(ctx,
		`DELETE FROM plan_comments WHERE id = $1`,
		id,
	)
	return err
}

func (r *PlanCommentRepository) MarkOutdatedByPlanDocumentID(ctx context.Context, planDocumentID string) error {
	_, err := r.db.ExecContext(ctx,
		`UPDATE plan_comments SET status = $1, updated_at = $2
		 WHERE plan_document_id = $3 AND status = $4`,
		string(domain.PlanCommentStatusOutdated), time.Now(),
		planDocumentID, string(domain.PlanCommentStatusActive),
	)
	return err
}

func (r *PlanCommentRepository) scanComment(row *sql.Row) (*domain.PlanComment, error) {
	var comment domain.PlanComment
	var status string
	var createdAt, updatedAt sql.NullTime

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
	if createdAt.Valid {
		comment.CreatedAt = createdAt.Time
	}
	if updatedAt.Valid {
		comment.UpdatedAt = updatedAt.Time
	}

	return &comment, nil
}

func (r *PlanCommentRepository) scanCommentFromRows(rows *sql.Rows) (*domain.PlanComment, error) {
	var comment domain.PlanComment
	var status string
	var createdAt, updatedAt sql.NullTime

	err := rows.Scan(
		&comment.ID, &comment.PlanDocumentID, &comment.UserID,
		&comment.TargetText, &comment.ContextBefore, &comment.ContextAfter, &comment.OriginalBodyHash,
		&comment.Content, &status, &createdAt, &updatedAt,
	)
	if err != nil {
		return nil, err
	}

	comment.Status = domain.PlanCommentStatus(status)
	if createdAt.Valid {
		comment.CreatedAt = createdAt.Time
	}
	if updatedAt.Valid {
		comment.UpdatedAt = updatedAt.Time
	}

	return &comment, nil
}
