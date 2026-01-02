package repository

import (
	"context"
	"time"

	"github.com/satetsu888/agentrace/server/internal/domain"
)

// ProjectRepository はプロジェクトの永続化を担当する
type ProjectRepository interface {
	Create(ctx context.Context, project *domain.Project) error
	FindByID(ctx context.Context, id string) (*domain.Project, error)
	FindByCanonicalGitRepository(ctx context.Context, canonicalGitRepo string) (*domain.Project, error)
	FindOrCreateByCanonicalGitRepository(ctx context.Context, canonicalGitRepo string) (*domain.Project, error)
	FindAll(ctx context.Context, limit int, offset int) ([]*domain.Project, error)
	GetDefaultProject(ctx context.Context) (*domain.Project, error) // CanonicalGitRepository が空のプロジェクト
}

// SessionRepository はセッションの永続化を担当する
type SessionRepository interface {
	Create(ctx context.Context, session *domain.Session) error
	FindByID(ctx context.Context, id string) (*domain.Session, error)
	FindByClaudeSessionID(ctx context.Context, claudeSessionID string) (*domain.Session, error)
	FindAll(ctx context.Context, limit int, offset int) ([]*domain.Session, error)
	FindByProjectID(ctx context.Context, projectID string, limit int, offset int) ([]*domain.Session, error)
	FindOrCreateByClaudeSessionID(ctx context.Context, claudeSessionID string, userID *string) (*domain.Session, error)
	UpdateUserID(ctx context.Context, id string, userID string) error
	UpdateProjectPath(ctx context.Context, id string, projectPath string) error
	UpdateProjectID(ctx context.Context, id string, projectID string) error
	UpdateGitBranch(ctx context.Context, id string, gitBranch string) error
	UpdateTitle(ctx context.Context, id string, title string) error
	UpdateUpdatedAt(ctx context.Context, id string, updatedAt time.Time) error
}

// EventRepository はイベントの永続化を担当する
type EventRepository interface {
	Create(ctx context.Context, event *domain.Event) error
	FindBySessionID(ctx context.Context, sessionID string) ([]*domain.Event, error)
	CountBySessionID(ctx context.Context, sessionID string) (int, error)
}

// UserRepository はユーザーの永続化を担当する
type UserRepository interface {
	Create(ctx context.Context, user *domain.User) error
	FindByID(ctx context.Context, id string) (*domain.User, error)
	FindByEmail(ctx context.Context, email string) (*domain.User, error)
	FindAll(ctx context.Context) ([]*domain.User, error)
	UpdateDisplayName(ctx context.Context, id string, displayName string) error
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

// PasswordCredentialRepository はパスワード認証情報の永続化を担当する
type PasswordCredentialRepository interface {
	Create(ctx context.Context, cred *domain.PasswordCredential) error
	FindByUserID(ctx context.Context, userID string) (*domain.PasswordCredential, error)
	Update(ctx context.Context, cred *domain.PasswordCredential) error
	Delete(ctx context.Context, id string) error
}

// OAuthConnectionRepository はOAuth連携の永続化を担当する
type OAuthConnectionRepository interface {
	Create(ctx context.Context, conn *domain.OAuthConnection) error
	FindByProviderAndProviderID(ctx context.Context, provider, providerID string) (*domain.OAuthConnection, error)
	FindByUserID(ctx context.Context, userID string) ([]*domain.OAuthConnection, error)
	Delete(ctx context.Context, id string) error
}

// PlanDocumentRepository はPlanDocumentの永続化を担当する
type PlanDocumentRepository interface {
	Create(ctx context.Context, doc *domain.PlanDocument) error
	FindByID(ctx context.Context, id string) (*domain.PlanDocument, error)
	FindAll(ctx context.Context, limit int, offset int) ([]*domain.PlanDocument, error)
	FindByProjectID(ctx context.Context, projectID string, limit int, offset int) ([]*domain.PlanDocument, error)
	FindByStatuses(ctx context.Context, statuses []domain.PlanDocumentStatus, projectID string, limit int, offset int) ([]*domain.PlanDocument, error)
	Update(ctx context.Context, doc *domain.PlanDocument) error
	Delete(ctx context.Context, id string) error
	SetStatus(ctx context.Context, id string, status domain.PlanDocumentStatus) error
}

// PlanDocumentEventRepository はPlanDocumentEventの永続化を担当する
type PlanDocumentEventRepository interface {
	Create(ctx context.Context, event *domain.PlanDocumentEvent) error
	FindByPlanDocumentID(ctx context.Context, planDocumentID string) ([]*domain.PlanDocumentEvent, error)
	FindByClaudeSessionID(ctx context.Context, claudeSessionID string) ([]*domain.PlanDocumentEvent, error)
	GetCollaboratorUserIDs(ctx context.Context, planDocumentID string) ([]string, error)
}

// Repositories は全リポジトリをまとめる
type Repositories struct {
	Project            ProjectRepository
	Session            SessionRepository
	Event              EventRepository
	User               UserRepository
	APIKey             APIKeyRepository
	WebSession         WebSessionRepository
	PasswordCredential PasswordCredentialRepository
	OAuthConnection    OAuthConnectionRepository
	PlanDocument       PlanDocumentRepository
	PlanDocumentEvent  PlanDocumentEventRepository
}
