package sqlite

import (
	"context"
	"database/sql"
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
		 VALUES (?, ?, ?, ?, ?, ?, ?)`,
		doc.ID, doc.ProjectID, doc.Description, doc.Body, string(doc.Status),
		doc.CreatedAt.Format(time.RFC3339), doc.UpdatedAt.Format(time.RFC3339),
	)
	return err
}

func (r *PlanDocumentRepository) FindByID(ctx context.Context, id string) (*domain.PlanDocument, error) {
	return r.scanDocument(r.db.QueryRowContext(ctx,
		`SELECT id, project_id, description, body, status, created_at, updated_at
		 FROM plan_documents WHERE id = ?`,
		id,
	))
}

func (r *PlanDocumentRepository) FindAll(ctx context.Context, limit int, offset int) ([]*domain.PlanDocument, error) {
	query := `SELECT id, project_id, description, body, status, created_at, updated_at
		 FROM plan_documents ORDER BY updated_at DESC`

	if limit > 0 {
		query += ` LIMIT ? OFFSET ?`
	}

	var rows *sql.Rows
	var err error

	if limit > 0 {
		rows, err = r.db.QueryContext(ctx, query, limit, offset)
	} else {
		rows, err = r.db.QueryContext(ctx, query)
	}

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

func (r *PlanDocumentRepository) FindByProjectID(ctx context.Context, projectID string, limit int, offset int) ([]*domain.PlanDocument, error) {
	query := `SELECT id, project_id, description, body, status, created_at, updated_at
		 FROM plan_documents WHERE project_id = ? ORDER BY updated_at DESC`

	if limit > 0 {
		query += ` LIMIT ? OFFSET ?`
	}

	var rows *sql.Rows
	var err error

	if limit > 0 {
		rows, err = r.db.QueryContext(ctx, query, projectID, limit, offset)
	} else {
		rows, err = r.db.QueryContext(ctx, query, projectID)
	}

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
		`UPDATE plan_documents SET project_id = ?, description = ?, body = ?, status = ?, updated_at = ?
		 WHERE id = ?`,
		doc.ProjectID, doc.Description, doc.Body, string(doc.Status), doc.UpdatedAt.Format(time.RFC3339), doc.ID,
	)
	return err
}

func (r *PlanDocumentRepository) Delete(ctx context.Context, id string) error {
	_, err := r.db.ExecContext(ctx,
		`DELETE FROM plan_documents WHERE id = ?`,
		id,
	)
	return err
}

func (r *PlanDocumentRepository) SetStatus(ctx context.Context, id string, status domain.PlanDocumentStatus) error {
	_, err := r.db.ExecContext(ctx,
		`UPDATE plan_documents SET status = ?, updated_at = ? WHERE id = ?`,
		string(status), time.Now().Format(time.RFC3339), id,
	)
	return err
}

func (r *PlanDocumentRepository) scanDocument(row *sql.Row) (*domain.PlanDocument, error) {
	var doc domain.PlanDocument
	var projectID sql.NullString
	var status string
	var createdAt, updatedAt string

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
	doc.CreatedAt, _ = time.Parse(time.RFC3339, createdAt)
	doc.UpdatedAt, _ = time.Parse(time.RFC3339, updatedAt)

	return &doc, nil
}

func (r *PlanDocumentRepository) scanDocumentFromRows(rows *sql.Rows) (*domain.PlanDocument, error) {
	var doc domain.PlanDocument
	var projectID sql.NullString
	var status string
	var createdAt, updatedAt string

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
	doc.CreatedAt, _ = time.Parse(time.RFC3339, createdAt)
	doc.UpdatedAt, _ = time.Parse(time.RFC3339, updatedAt)

	return &doc, nil
}
