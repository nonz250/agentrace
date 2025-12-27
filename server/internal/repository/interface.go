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
	FindOrCreateByClaudeSessionID(ctx context.Context, claudeSessionID string) (*domain.Session, error)
}

// EventRepository はイベントの永続化を担当する
type EventRepository interface {
	Create(ctx context.Context, event *domain.Event) error
	FindBySessionID(ctx context.Context, sessionID string) ([]*domain.Event, error)
}

// Repositories は全リポジトリをまとめる
type Repositories struct {
	Session SessionRepository
	Event   EventRepository
}
