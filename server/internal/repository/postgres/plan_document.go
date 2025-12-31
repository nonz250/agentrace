package postgres

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

	_, err := r.db.ExecContext(ctx,
		`INSERT INTO plan_documents (id, description, body, git_remote_url, created_at, updated_at)
		 VALUES ($1, $2, $3, $4, $5, $6)`,
		doc.ID, doc.Description, doc.Body, doc.GitRemoteURL, doc.CreatedAt, doc.UpdatedAt,
	)
	return err
}

func (r *PlanDocumentRepository) FindByID(ctx context.Context, id string) (*domain.PlanDocument, error) {
	return r.scanDocument(r.db.QueryRowContext(ctx,
		`SELECT id, description, body, git_remote_url, created_at, updated_at
		 FROM plan_documents WHERE id = $1`,
		id,
	))
}

func (r *PlanDocumentRepository) FindAll(ctx context.Context, limit int, offset int) ([]*domain.PlanDocument, error) {
	var rows *sql.Rows
	var err error

	if limit > 0 {
		rows, err = r.db.QueryContext(ctx,
			`SELECT id, description, body, git_remote_url, created_at, updated_at
			 FROM plan_documents ORDER BY updated_at DESC LIMIT $1 OFFSET $2`,
			limit, offset,
		)
	} else {
		rows, err = r.db.QueryContext(ctx,
			`SELECT id, description, body, git_remote_url, created_at, updated_at
			 FROM plan_documents ORDER BY updated_at DESC`,
		)
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

func (r *PlanDocumentRepository) FindByGitRemoteURL(ctx context.Context, gitRemoteURL string, limit int, offset int) ([]*domain.PlanDocument, error) {
	var rows *sql.Rows
	var err error

	if limit > 0 {
		rows, err = r.db.QueryContext(ctx,
			`SELECT id, description, body, git_remote_url, created_at, updated_at
			 FROM plan_documents WHERE git_remote_url = $1 ORDER BY updated_at DESC LIMIT $2 OFFSET $3`,
			gitRemoteURL, limit, offset,
		)
	} else {
		rows, err = r.db.QueryContext(ctx,
			`SELECT id, description, body, git_remote_url, created_at, updated_at
			 FROM plan_documents WHERE git_remote_url = $1 ORDER BY updated_at DESC`,
			gitRemoteURL,
		)
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
		`UPDATE plan_documents SET description = $1, body = $2, git_remote_url = $3, updated_at = $4
		 WHERE id = $5`,
		doc.Description, doc.Body, doc.GitRemoteURL, doc.UpdatedAt, doc.ID,
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

func (r *PlanDocumentRepository) scanDocument(row *sql.Row) (*domain.PlanDocument, error) {
	var doc domain.PlanDocument
	var createdAt, updatedAt sql.NullTime

	err := row.Scan(&doc.ID, &doc.Description, &doc.Body, &doc.GitRemoteURL, &createdAt, &updatedAt)
	if err == sql.ErrNoRows {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}

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
	var createdAt, updatedAt sql.NullTime

	err := rows.Scan(&doc.ID, &doc.Description, &doc.Body, &doc.GitRemoteURL, &createdAt, &updatedAt)
	if err != nil {
		return nil, err
	}

	if createdAt.Valid {
		doc.CreatedAt = createdAt.Time
	}
	if updatedAt.Valid {
		doc.UpdatedAt = updatedAt.Time
	}

	return &doc, nil
}
