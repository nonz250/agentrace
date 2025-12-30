package memory

import "github.com/satetsu888/agentrace/server/internal/repository"

func NewRepositories() *repository.Repositories {
	return &repository.Repositories{
		Session:            NewSessionRepository(),
		Event:              NewEventRepository(),
		User:               NewUserRepository(),
		APIKey:             NewAPIKeyRepository(),
		WebSession:         NewWebSessionRepository(),
		PasswordCredential: NewPasswordCredentialRepository(),
		OAuthConnection:    NewOAuthConnectionRepository(),
	}
}
