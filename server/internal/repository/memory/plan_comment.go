package memory

import (
	"context"
	"sort"
	"sync"
	"time"

	"github.com/google/uuid"
	"github.com/satetsu888/agentrace/server/internal/domain"
)

type PlanCommentRepository struct {
	mu       sync.RWMutex
	comments map[string]*domain.PlanComment
}

func NewPlanCommentRepository() *PlanCommentRepository {
	return &PlanCommentRepository{
		comments: make(map[string]*domain.PlanComment),
	}
}

func (r *PlanCommentRepository) Create(ctx context.Context, comment *domain.PlanComment) error {
	r.mu.Lock()
	defer r.mu.Unlock()

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

	r.comments[comment.ID] = comment
	return nil
}

func (r *PlanCommentRepository) FindByID(ctx context.Context, id string) (*domain.PlanComment, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()

	comment, ok := r.comments[id]
	if !ok {
		return nil, nil
	}
	return comment, nil
}

func (r *PlanCommentRepository) FindByPlanDocumentID(ctx context.Context, planDocumentID string, status *domain.PlanCommentStatus) ([]*domain.PlanComment, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()

	var comments []*domain.PlanComment
	for _, c := range r.comments {
		if c.PlanDocumentID != planDocumentID {
			continue
		}
		if status != nil && c.Status != *status {
			continue
		}
		comments = append(comments, c)
	}

	// Sort by created_at ascending
	sort.Slice(comments, func(i, j int) bool {
		return comments[i].CreatedAt.Before(comments[j].CreatedAt)
	})

	return comments, nil
}

func (r *PlanCommentRepository) Update(ctx context.Context, comment *domain.PlanComment) error {
	r.mu.Lock()
	defer r.mu.Unlock()

	if _, ok := r.comments[comment.ID]; !ok {
		return nil
	}

	comment.UpdatedAt = time.Now()
	r.comments[comment.ID] = comment
	return nil
}

func (r *PlanCommentRepository) Delete(ctx context.Context, id string) error {
	r.mu.Lock()
	defer r.mu.Unlock()

	delete(r.comments, id)
	return nil
}

func (r *PlanCommentRepository) MarkOutdatedByPlanDocumentID(ctx context.Context, planDocumentID string) error {
	r.mu.Lock()
	defer r.mu.Unlock()

	now := time.Now()
	for _, c := range r.comments {
		if c.PlanDocumentID == planDocumentID && c.Status == domain.PlanCommentStatusActive {
			c.Status = domain.PlanCommentStatusOutdated
			c.UpdatedAt = now
		}
	}
	return nil
}
