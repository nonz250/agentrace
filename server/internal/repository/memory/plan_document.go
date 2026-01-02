package memory

import (
	"context"
	"sort"
	"sync"
	"time"

	"github.com/google/uuid"
	"github.com/satetsu888/agentrace/server/internal/domain"
)

type PlanDocumentRepository struct {
	mu        sync.RWMutex
	documents map[string]*domain.PlanDocument
}

func NewPlanDocumentRepository() *PlanDocumentRepository {
	return &PlanDocumentRepository{
		documents: make(map[string]*domain.PlanDocument),
	}
}

func (r *PlanDocumentRepository) Create(ctx context.Context, doc *domain.PlanDocument) error {
	r.mu.Lock()
	defer r.mu.Unlock()

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

	r.documents[doc.ID] = doc
	return nil
}

func (r *PlanDocumentRepository) FindByID(ctx context.Context, id string) (*domain.PlanDocument, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()

	doc, ok := r.documents[id]
	if !ok {
		return nil, nil
	}
	return doc, nil
}

func (r *PlanDocumentRepository) FindAll(ctx context.Context, limit int, offset int) ([]*domain.PlanDocument, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()

	docs := make([]*domain.PlanDocument, 0, len(r.documents))
	for _, d := range r.documents {
		docs = append(docs, d)
	}

	// Sort by UpdatedAt descending (newest first)
	sort.Slice(docs, func(i, j int) bool {
		return docs[i].UpdatedAt.After(docs[j].UpdatedAt)
	})

	// Apply offset and limit
	if offset >= len(docs) {
		return []*domain.PlanDocument{}, nil
	}
	docs = docs[offset:]
	if limit > 0 && limit < len(docs) {
		docs = docs[:limit]
	}

	return docs, nil
}

func (r *PlanDocumentRepository) FindByProjectID(ctx context.Context, projectID string, limit int, offset int) ([]*domain.PlanDocument, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()

	docs := make([]*domain.PlanDocument, 0)
	for _, d := range r.documents {
		if d.ProjectID == projectID {
			docs = append(docs, d)
		}
	}

	// Sort by UpdatedAt descending (newest first)
	sort.Slice(docs, func(i, j int) bool {
		return docs[i].UpdatedAt.After(docs[j].UpdatedAt)
	})

	// Apply offset and limit
	if offset >= len(docs) {
		return []*domain.PlanDocument{}, nil
	}
	docs = docs[offset:]
	if limit > 0 && limit < len(docs) {
		docs = docs[:limit]
	}

	return docs, nil
}

func (r *PlanDocumentRepository) FindByStatuses(ctx context.Context, statuses []domain.PlanDocumentStatus, projectID string, limit int, offset int) ([]*domain.PlanDocument, error) {
	if len(statuses) == 0 {
		if projectID != "" {
			return r.FindByProjectID(ctx, projectID, limit, offset)
		}
		return r.FindAll(ctx, limit, offset)
	}

	r.mu.RLock()
	defer r.mu.RUnlock()

	// Build status set for fast lookup
	statusSet := make(map[domain.PlanDocumentStatus]bool)
	for _, s := range statuses {
		statusSet[s] = true
	}

	docs := make([]*domain.PlanDocument, 0)
	for _, d := range r.documents {
		if !statusSet[d.Status] {
			continue
		}
		if projectID != "" && d.ProjectID != projectID {
			continue
		}
		docs = append(docs, d)
	}

	// Sort by UpdatedAt descending (newest first)
	sort.Slice(docs, func(i, j int) bool {
		return docs[i].UpdatedAt.After(docs[j].UpdatedAt)
	})

	// Apply offset and limit
	if offset >= len(docs) {
		return []*domain.PlanDocument{}, nil
	}
	docs = docs[offset:]
	if limit > 0 && limit < len(docs) {
		docs = docs[:limit]
	}

	return docs, nil
}

func (r *PlanDocumentRepository) Update(ctx context.Context, doc *domain.PlanDocument) error {
	r.mu.Lock()
	defer r.mu.Unlock()

	if _, ok := r.documents[doc.ID]; !ok {
		return nil
	}

	doc.UpdatedAt = time.Now()
	r.documents[doc.ID] = doc
	return nil
}

func (r *PlanDocumentRepository) Delete(ctx context.Context, id string) error {
	r.mu.Lock()
	defer r.mu.Unlock()

	delete(r.documents, id)
	return nil
}

func (r *PlanDocumentRepository) SetStatus(ctx context.Context, id string, status domain.PlanDocumentStatus) error {
	r.mu.Lock()
	defer r.mu.Unlock()

	doc, ok := r.documents[id]
	if !ok {
		return nil
	}

	doc.Status = status
	doc.UpdatedAt = time.Now()
	return nil
}
