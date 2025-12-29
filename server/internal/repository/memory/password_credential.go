package memory

import (
	"context"
	"sync"
	"time"

	"github.com/google/uuid"
	"github.com/satetsu888/agentrace/server/internal/domain"
)

type PasswordCredentialRepository struct {
	mu          sync.RWMutex
	credentials map[string]*domain.PasswordCredential // key: ID
	byUserID    map[string]string                     // userID -> credential ID
}

func NewPasswordCredentialRepository() *PasswordCredentialRepository {
	return &PasswordCredentialRepository{
		credentials: make(map[string]*domain.PasswordCredential),
		byUserID:    make(map[string]string),
	}
}

func (r *PasswordCredentialRepository) Create(ctx context.Context, cred *domain.PasswordCredential) error {
	r.mu.Lock()
	defer r.mu.Unlock()

	if cred.ID == "" {
		cred.ID = uuid.New().String()
	}
	now := time.Now()
	if cred.CreatedAt.IsZero() {
		cred.CreatedAt = now
	}
	if cred.UpdatedAt.IsZero() {
		cred.UpdatedAt = now
	}

	r.credentials[cred.ID] = cred
	r.byUserID[cred.UserID] = cred.ID
	return nil
}

func (r *PasswordCredentialRepository) FindByUserID(ctx context.Context, userID string) (*domain.PasswordCredential, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()

	credID, ok := r.byUserID[userID]
	if !ok {
		return nil, nil
	}
	return r.credentials[credID], nil
}

func (r *PasswordCredentialRepository) Update(ctx context.Context, cred *domain.PasswordCredential) error {
	r.mu.Lock()
	defer r.mu.Unlock()

	if _, ok := r.credentials[cred.ID]; !ok {
		return nil
	}

	cred.UpdatedAt = time.Now()
	r.credentials[cred.ID] = cred
	return nil
}

func (r *PasswordCredentialRepository) Delete(ctx context.Context, id string) error {
	r.mu.Lock()
	defer r.mu.Unlock()

	cred, ok := r.credentials[id]
	if ok {
		delete(r.byUserID, cred.UserID)
		delete(r.credentials, id)
	}
	return nil
}
