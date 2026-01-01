package memory

import (
	"context"
	"sort"
	"sync"
	"time"

	"github.com/google/uuid"
	"github.com/satetsu888/agentrace/server/internal/domain"
)

type ProjectRepository struct {
	mu       sync.RWMutex
	projects map[string]*domain.Project
}

func NewProjectRepository() *ProjectRepository {
	repo := &ProjectRepository{
		projects: make(map[string]*domain.Project),
	}
	// Create default project
	repo.projects[domain.DefaultProjectID] = &domain.Project{
		ID:                     domain.DefaultProjectID,
		CanonicalGitRepository: "",
		CreatedAt:              time.Now(),
	}
	return repo
}

func (r *ProjectRepository) Create(ctx context.Context, project *domain.Project) error {
	r.mu.Lock()
	defer r.mu.Unlock()

	if project.ID == "" {
		project.ID = uuid.New().String()
	}
	if project.CreatedAt.IsZero() {
		project.CreatedAt = time.Now()
	}

	r.projects[project.ID] = project
	return nil
}

func (r *ProjectRepository) FindByID(ctx context.Context, id string) (*domain.Project, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()

	project, ok := r.projects[id]
	if !ok {
		return nil, nil
	}
	return project, nil
}

func (r *ProjectRepository) FindByCanonicalGitRepository(ctx context.Context, canonicalGitRepo string) (*domain.Project, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()

	for _, p := range r.projects {
		if p.CanonicalGitRepository == canonicalGitRepo {
			return p, nil
		}
	}
	return nil, nil
}

func (r *ProjectRepository) FindOrCreateByCanonicalGitRepository(ctx context.Context, canonicalGitRepo string) (*domain.Project, error) {
	// First try to find existing project (with read lock)
	r.mu.RLock()
	for _, p := range r.projects {
		if p.CanonicalGitRepository == canonicalGitRepo {
			r.mu.RUnlock()
			return p, nil
		}
	}
	r.mu.RUnlock()

	// Create new project (with write lock)
	r.mu.Lock()
	defer r.mu.Unlock()

	// Double-check after acquiring write lock
	for _, p := range r.projects {
		if p.CanonicalGitRepository == canonicalGitRepo {
			return p, nil
		}
	}

	project := &domain.Project{
		ID:                     uuid.New().String(),
		CanonicalGitRepository: canonicalGitRepo,
		CreatedAt:              time.Now(),
	}
	r.projects[project.ID] = project
	return project, nil
}

func (r *ProjectRepository) FindAll(ctx context.Context, limit int, offset int) ([]*domain.Project, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()

	projects := make([]*domain.Project, 0, len(r.projects))
	for _, p := range r.projects {
		projects = append(projects, p)
	}

	// Sort by CreatedAt descending (newest first)
	sort.Slice(projects, func(i, j int) bool {
		return projects[i].CreatedAt.After(projects[j].CreatedAt)
	})

	// Apply offset and limit
	if offset >= len(projects) {
		return []*domain.Project{}, nil
	}
	projects = projects[offset:]
	if limit > 0 && limit < len(projects) {
		projects = projects[:limit]
	}

	return projects, nil
}

func (r *ProjectRepository) GetDefaultProject(ctx context.Context) (*domain.Project, error) {
	return r.FindByID(ctx, domain.DefaultProjectID)
}
