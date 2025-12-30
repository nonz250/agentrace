package memory

import (
	"context"
	"sync"
	"time"

	"github.com/google/uuid"
	"github.com/satetsu888/agentrace/server/internal/domain"
)

type OAuthConnectionRepository struct {
	mu           sync.RWMutex
	connections  map[string]*domain.OAuthConnection // key: ID
	byProviderID map[string]string                  // "provider:providerID" -> connection ID
}

func NewOAuthConnectionRepository() *OAuthConnectionRepository {
	return &OAuthConnectionRepository{
		connections:  make(map[string]*domain.OAuthConnection),
		byProviderID: make(map[string]string),
	}
}

func (r *OAuthConnectionRepository) Create(ctx context.Context, conn *domain.OAuthConnection) error {
	r.mu.Lock()
	defer r.mu.Unlock()

	if conn.ID == "" {
		conn.ID = uuid.New().String()
	}
	if conn.CreatedAt.IsZero() {
		conn.CreatedAt = time.Now()
	}

	r.connections[conn.ID] = conn
	r.byProviderID[conn.Provider+":"+conn.ProviderID] = conn.ID
	return nil
}

func (r *OAuthConnectionRepository) FindByProviderAndProviderID(ctx context.Context, provider, providerID string) (*domain.OAuthConnection, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()

	connID, ok := r.byProviderID[provider+":"+providerID]
	if !ok {
		return nil, nil
	}
	return r.connections[connID], nil
}

func (r *OAuthConnectionRepository) FindByUserID(ctx context.Context, userID string) ([]*domain.OAuthConnection, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()

	var result []*domain.OAuthConnection
	for _, conn := range r.connections {
		if conn.UserID == userID {
			result = append(result, conn)
		}
	}
	return result, nil
}

func (r *OAuthConnectionRepository) Delete(ctx context.Context, id string) error {
	r.mu.Lock()
	defer r.mu.Unlock()

	conn, ok := r.connections[id]
	if ok {
		delete(r.byProviderID, conn.Provider+":"+conn.ProviderID)
		delete(r.connections, id)
	}
	return nil
}
