package sqlite

import (
	"context"
	"database/sql"
	"time"

	"github.com/google/uuid"
	"github.com/satetsu888/agentrace/server/internal/domain"
)

type ProjectRepository struct {
	db *DB
}

func NewProjectRepository(db *DB) *ProjectRepository {
	return &ProjectRepository{db: db}
}

func (r *ProjectRepository) Create(ctx context.Context, project *domain.Project) error {
	if project.ID == "" {
		project.ID = uuid.New().String()
	}
	if project.CreatedAt.IsZero() {
		project.CreatedAt = time.Now()
	}

	_, err := r.db.ExecContext(ctx,
		`INSERT INTO projects (id, canonical_git_repository, created_at)
		 VALUES (?, ?, ?)`,
		project.ID, project.CanonicalGitRepository, project.CreatedAt.Format(time.RFC3339),
	)
	return err
}

func (r *ProjectRepository) FindByID(ctx context.Context, id string) (*domain.Project, error) {
	return r.scanProject(r.db.QueryRowContext(ctx,
		`SELECT id, canonical_git_repository, created_at
		 FROM projects WHERE id = ?`,
		id,
	))
}

func (r *ProjectRepository) FindByCanonicalGitRepository(ctx context.Context, canonicalGitRepo string) (*domain.Project, error) {
	return r.scanProject(r.db.QueryRowContext(ctx,
		`SELECT id, canonical_git_repository, created_at
		 FROM projects WHERE canonical_git_repository = ?`,
		canonicalGitRepo,
	))
}

func (r *ProjectRepository) FindOrCreateByCanonicalGitRepository(ctx context.Context, canonicalGitRepo string) (*domain.Project, error) {
	// First try to find existing project
	project, err := r.FindByCanonicalGitRepository(ctx, canonicalGitRepo)
	if err != nil {
		return nil, err
	}

	if project != nil {
		return project, nil
	}

	// Create new project
	newProject := &domain.Project{
		ID:                     uuid.New().String(),
		CanonicalGitRepository: canonicalGitRepo,
		CreatedAt:              time.Now(),
	}

	if err := r.Create(ctx, newProject); err != nil {
		// Handle race condition - another process may have created it
		existingProject, findErr := r.FindByCanonicalGitRepository(ctx, canonicalGitRepo)
		if findErr != nil {
			return nil, err // Return original error
		}
		if existingProject != nil {
			return existingProject, nil
		}
		return nil, err
	}

	return newProject, nil
}

func (r *ProjectRepository) FindAll(ctx context.Context, limit int, offset int) ([]*domain.Project, error) {
	query := `SELECT id, canonical_git_repository, created_at
		 FROM projects ORDER BY created_at DESC`

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

	var projects []*domain.Project
	for rows.Next() {
		project, err := r.scanProjectFromRows(rows)
		if err != nil {
			return nil, err
		}
		projects = append(projects, project)
	}

	return projects, rows.Err()
}

func (r *ProjectRepository) GetDefaultProject(ctx context.Context) (*domain.Project, error) {
	return r.FindByID(ctx, domain.DefaultProjectID)
}

func (r *ProjectRepository) scanProject(row *sql.Row) (*domain.Project, error) {
	var project domain.Project
	var createdAt sql.NullString

	err := row.Scan(&project.ID, &project.CanonicalGitRepository, &createdAt)
	if err == sql.ErrNoRows {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}

	if createdAt.Valid {
		project.CreatedAt, _ = time.Parse(time.RFC3339, createdAt.String)
	}

	return &project, nil
}

func (r *ProjectRepository) scanProjectFromRows(rows *sql.Rows) (*domain.Project, error) {
	var project domain.Project
	var createdAt sql.NullString

	err := rows.Scan(&project.ID, &project.CanonicalGitRepository, &createdAt)
	if err != nil {
		return nil, err
	}

	if createdAt.Valid {
		project.CreatedAt, _ = time.Parse(time.RFC3339, createdAt.String)
	}

	return &project, nil
}
