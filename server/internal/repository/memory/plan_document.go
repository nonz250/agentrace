package memory

import (
	"context"
	"sort"
	"strings"
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

func (r *PlanDocumentRepository) Find(ctx context.Context, query domain.PlanDocumentQuery) ([]*domain.PlanDocument, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()

	// Build status set for fast lookup
	var statusSet map[domain.PlanDocumentStatus]bool
	if len(query.Statuses) > 0 {
		statusSet = make(map[domain.PlanDocumentStatus]bool)
		for _, s := range query.Statuses {
			statusSet[s] = true
		}
	}

	// Prepare description filter
	lowerDescFilter := strings.ToLower(query.DescriptionContains)

	// Filter documents
	docs := make([]*domain.PlanDocument, 0)
	for _, d := range r.documents {
		// Check status filter
		if statusSet != nil && !statusSet[d.Status] {
			continue
		}
		// Check project filter
		if query.ProjectID != "" && d.ProjectID != query.ProjectID {
			continue
		}
		// Check description filter (case-insensitive)
		if lowerDescFilter != "" && !strings.Contains(strings.ToLower(d.Description), lowerDescFilter) {
			continue
		}
		docs = append(docs, d)
	}

	// Sort by UpdatedAt descending (newest first)
	sort.Slice(docs, func(i, j int) bool {
		return docs[i].UpdatedAt.After(docs[j].UpdatedAt)
	})

	// Apply offset and limit
	if query.Offset >= len(docs) {
		return []*domain.PlanDocument{}, nil
	}
	docs = docs[query.Offset:]
	if query.Limit > 0 && query.Limit < len(docs) {
		docs = docs[:query.Limit]
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
