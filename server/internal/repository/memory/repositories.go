package memory

import "github.com/satetsu888/agentrace/server/internal/repository"

func NewRepositories() *repository.Repositories {
	projectRepo := NewProjectRepository()
	sessionRepo := NewSessionRepository()
	planDocRepo := NewPlanDocumentRepository()

	// Set related repos for HasRelatedData
	projectRepo.SetRelatedRepos(sessionRepo, planDocRepo)

	return &repository.Repositories{
		Project:            projectRepo,
		Session:            sessionRepo,
		Event:              NewEventRepository(),
		User:               NewUserRepository(),
		APIKey:             NewAPIKeyRepository(),
		WebSession:         NewWebSessionRepository(),
		PasswordCredential: NewPasswordCredentialRepository(),
		OAuthConnection:    NewOAuthConnectionRepository(),
		PlanDocument:       planDocRepo,
		PlanDocumentEvent:  NewPlanDocumentEventRepository(),
		UserFavorite:       NewUserFavoriteRepository(),
	}
}
