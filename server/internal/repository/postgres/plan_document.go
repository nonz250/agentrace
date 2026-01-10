package postgres

import (
	"context"
	"database/sql"
	"fmt"
	"strings"
	"time"

	"github.com/google/uuid"
	"github.com/satetsu888/agentrace/server/internal/domain"
)

type PlanDocumentRepository struct {
	db *DB
}

func NewPlanDocumentRepository(db *DB) *PlanDocumentRepository {
	return &PlanDocumentRepository{db: db}
}

func (r *PlanDocumentRepository) Create(ctx context.Context, doc *domain.PlanDocument) error {
	if doc.ID == "" {
		doc.ID = uuid.New().String()
	}
	now := time.Now()
	if doc.CreatedAt.IsZero() {
		doc.CreatedAt = now
	}
	if doc.UpdatedAt.IsZero() {
		doc.UpdatedAt = now
	}
	if doc.Status == "" {
		doc.Status = domain.PlanDocumentStatusPlanning
	}
	if doc.ProjectID == "" {
		doc.ProjectID = domain.DefaultProjectID
	}

	_, err := r.db.ExecContext(ctx,
		`INSERT INTO plan_documents (id, project_id, description, body, status, created_at, updated_at)
		 VALUES ($1, $2, $3, $4, $5, $6, $7)`,
		doc.ID, doc.ProjectID, doc.Description, doc.Body, string(doc.Status), doc.CreatedAt, doc.UpdatedAt,
	)
	return err
}

func (r *PlanDocumentRepository) FindByID(ctx context.Context, id string) (*domain.PlanDocument, error) {
	return r.scanDocument(r.db.QueryRowContext(ctx,
		`SELECT id, project_id, description, body, status, created_at, updated_at
		 FROM plan_documents WHERE id = $1`,
		id,
	))
}

func (r *PlanDocumentRepository) Find(ctx context.Context, query domain.PlanDocumentQuery) ([]*domain.PlanDocument, error) {
	baseQuery := `SELECT id, project_id, description, body, status, created_at, updated_at FROM plan_documents`
	var conditions []string
	var args []any
	paramIdx := 1

	// Build WHERE conditions
	if len(query.PlanDocumentIDs) > 0 {
		placeholders := make([]string, len(query.PlanDocumentIDs))
		for i, id := range query.PlanDocumentIDs {
			placeholders[i] = fmt.Sprintf("$%d", paramIdx)
			args = append(args, id)
			paramIdx++
		}
		conditions = append(conditions, "id IN ("+strings.Join(placeholders, ", ")+")")
	}

	if len(query.Statuses) > 0 {
		placeholders := make([]string, len(query.Statuses))
		for i, s := range query.Statuses {
			placeholders[i] = fmt.Sprintf("$%d", paramIdx)
			args = append(args, string(s))
			paramIdx++
		}
		conditions = append(conditions, "status IN ("+strings.Join(placeholders, ", ")+")")
	}

	if query.ProjectID != "" {
		conditions = append(conditions, fmt.Sprintf("project_id = $%d", paramIdx))
		args = append(args, query.ProjectID)
		paramIdx++
	}

	if query.DescriptionContains != "" {
		conditions = append(conditions, fmt.Sprintf("description ILIKE $%d", paramIdx))
		args = append(args, "%"+query.DescriptionContains+"%")
		paramIdx++
	}

	// Combine query parts
	if len(conditions) > 0 {
		baseQuery += " WHERE " + strings.Join(conditions, " AND ")
	}

	// Validate sortBy to prevent SQL injection
	orderColumn := "updated_at"
	if query.SortBy == "created_at" {
		orderColumn = "created_at"
	}
	baseQuery += " ORDER BY " + orderColumn + " DESC"

	if query.Limit > 0 {
		baseQuery += fmt.Sprintf(" LIMIT $%d OFFSET $%d", paramIdx, paramIdx+1)
		args = append(args, query.Limit, query.Offset)
	}

	rows, err := r.db.QueryContext(ctx, baseQuery, args...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var docs []*domain.PlanDocument
	for rows.Next() {
		doc, err := r.scanDocumentFromRows(rows)
		if err != nil {
			return nil, err
		}
		docs = append(docs, doc)
	}

	return docs, rows.Err()
}

func (r *PlanDocumentRepository) Update(ctx context.Context, doc *domain.PlanDocument) error {
	doc.UpdatedAt = time.Now()

	_, err := r.db.ExecContext(ctx,
		`UPDATE plan_documents SET project_id = $1, description = $2, body = $3, status = $4, updated_at = $5
		 WHERE id = $6`,
		doc.ProjectID, doc.Description, doc.Body, string(doc.Status), doc.UpdatedAt, doc.ID,
	)
	return err
}

func (r *PlanDocumentRepository) Delete(ctx context.Context, id string) error {
	_, err := r.db.ExecContext(ctx,
		`DELETE FROM plan_documents WHERE id = $1`,
		id,
	)
	return err
}

func (r *PlanDocumentRepository) SetStatus(ctx context.Context, id string, status domain.PlanDocumentStatus) error {
	_, err := r.db.ExecContext(ctx,
		`UPDATE plan_documents SET status = $1, updated_at = $2 WHERE id = $3`,
		string(status), time.Now(), id,
	)
	return err
}

func (r *PlanDocumentRepository) scanDocument(row *sql.Row) (*domain.PlanDocument, error) {
	var doc domain.PlanDocument
	var projectID sql.NullString
	var status string
	var createdAt, updatedAt sql.NullTime

	err := row.Scan(&doc.ID, &projectID, &doc.Description, &doc.Body, &status, &createdAt, &updatedAt)
	if err == sql.ErrNoRows {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}

	if projectID.Valid {
		doc.ProjectID = projectID.String
	} else {
		doc.ProjectID = domain.DefaultProjectID
	}
	doc.Status = domain.PlanDocumentStatus(status)
	if createdAt.Valid {
		doc.CreatedAt = createdAt.Time
	}
	if updatedAt.Valid {
		doc.UpdatedAt = updatedAt.Time
	}

	return &doc, nil
}

func (r *PlanDocumentRepository) scanDocumentFromRows(rows *sql.Rows) (*domain.PlanDocument, error) {
	var doc domain.PlanDocument
	var projectID sql.NullString
	var status string
	var createdAt, updatedAt sql.NullTime

	err := rows.Scan(&doc.ID, &projectID, &doc.Description, &doc.Body, &status, &createdAt, &updatedAt)
	if err != nil {
		return nil, err
	}

	if projectID.Valid {
		doc.ProjectID = projectID.String
	} else {
		doc.ProjectID = domain.DefaultProjectID
	}
	doc.Status = domain.PlanDocumentStatus(status)
	if createdAt.Valid {
		doc.CreatedAt = createdAt.Time
	}
	if updatedAt.Valid {
		doc.UpdatedAt = updatedAt.Time
	}

	return &doc, nil
}
