package sqlite

import (
	"github.com/satetsu888/agentrace/server/internal/repository"
)

// NewRepositories creates all SQLite repositories
func NewRepositories(db *DB) *repository.Repositories {
	return &repository.Repositories{
		Session:            NewSessionRepository(db),
		Event:              NewEventRepository(db),
		User:               NewUserRepository(db),
		APIKey:             NewAPIKeyRepository(db),
		WebSession:         NewWebSessionRepository(db),
		PasswordCredential: NewPasswordCredentialRepository(db),
		OAuthConnection:    NewOAuthConnectionRepository(db),
	}
}
