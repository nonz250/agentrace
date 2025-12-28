package repository

import (
	"context"

	"github.com/satetsu888/agentrace/server/internal/domain"
)

// SessionRepository はセッションの永続化を担当する
type SessionRepository interface {
	Create(ctx context.Context, session *domain.Session) error
	FindByID(ctx context.Context, id string) (*domain.Session, error)
	FindAll(ctx context.Context, limit int, offset int) ([]*domain.Session, error)
	FindOrCreateByClaudeSessionID(ctx context.Context, claudeSessionID string, userID *string) (*domain.Session, error)
	UpdateUserID(ctx context.Context, id string, userID string) error
}

// EventRepository はイベントの永続化を担当する
type EventRepository interface {
	Create(ctx context.Context, event *domain.Event) error
	FindBySessionID(ctx context.Context, sessionID string) ([]*domain.Event, error)
}

// UserRepository はユーザーの永続化を担当する
type UserRepository interface {
	Create(ctx context.Context, user *domain.User) error
	FindByID(ctx context.Context, id string) (*domain.User, error)
	FindAll(ctx context.Context) ([]*domain.User, error)
}

// APIKeyRepository はAPIキーの永続化を担当する
type APIKeyRepository interface {
	Create(ctx context.Context, key *domain.APIKey) error
	FindByKeyHash(ctx context.Context, keyHash string) (*domain.APIKey, error)
	FindByUserID(ctx context.Context, userID string) ([]*domain.APIKey, error)
	FindByID(ctx context.Context, id string) (*domain.APIKey, error)
	Delete(ctx context.Context, id string) error
	UpdateLastUsedAt(ctx context.Context, id string) error
}

// WebSessionRepository はWebセッションの永続化を担当する
type WebSessionRepository interface {
	Create(ctx context.Context, session *domain.WebSession) error
	FindByToken(ctx context.Context, token string) (*domain.WebSession, error)
	Delete(ctx context.Context, id string) error
	DeleteExpired(ctx context.Context) error
}

// Repositories は全リポジトリをまとめる
type Repositories struct {
	Session    SessionRepository
	Event      EventRepository
	User       UserRepository
	APIKey     APIKeyRepository
	WebSession WebSessionRepository
}
