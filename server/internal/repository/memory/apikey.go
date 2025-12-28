package memory

import (
	"context"
	"sync"
	"time"

	"github.com/google/uuid"
	"github.com/satetsu888/agentrace/server/internal/domain"
)

type APIKeyRepository struct {
	mu   sync.RWMutex
	keys map[string]*domain.APIKey
}

func NewAPIKeyRepository() *APIKeyRepository {
	return &APIKeyRepository{
		keys: make(map[string]*domain.APIKey),
	}
}

func (r *APIKeyRepository) Create(ctx context.Context, key *domain.APIKey) error {
	r.mu.Lock()
	defer r.mu.Unlock()

	if key.ID == "" {
		key.ID = uuid.New().String()
	}
	if key.CreatedAt.IsZero() {
		key.CreatedAt = time.Now()
	}

	r.keys[key.ID] = key
	return nil
}

func (r *APIKeyRepository) FindByKeyHash(ctx context.Context, keyHash string) (*domain.APIKey, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()

	for _, k := range r.keys {
		if k.KeyHash == keyHash {
			return k, nil
		}
	}
	return nil, nil
}

func (r *APIKeyRepository) FindByUserID(ctx context.Context, userID string) ([]*domain.APIKey, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()

	keys := make([]*domain.APIKey, 0)
	for _, k := range r.keys {
		if k.UserID == userID {
			keys = append(keys, k)
		}
	}
	return keys, nil
}

func (r *APIKeyRepository) FindByID(ctx context.Context, id string) (*domain.APIKey, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()

	key, ok := r.keys[id]
	if !ok {
		return nil, nil
	}
	return key, nil
}

func (r *APIKeyRepository) Delete(ctx context.Context, id string) error {
	r.mu.Lock()
	defer r.mu.Unlock()

	delete(r.keys, id)
	return nil
}

func (r *APIKeyRepository) UpdateLastUsedAt(ctx context.Context, id string) error {
	r.mu.Lock()
	defer r.mu.Unlock()

	key, ok := r.keys[id]
	if !ok {
		return nil
	}
	now := time.Now()
	key.LastUsedAt = &now
	return nil
}
