package memory

import (
	"testing"

	"github.com/satetsu888/agentrace/server/internal/repository/testsuite"
	"github.com/stretchr/testify/suite"
)

func TestProjectRepository(t *testing.T) {
	repos := NewRepositories()
	s := &testsuite.ProjectRepositorySuite{
		Repo:  repos.Project,
		Repos: repos,
	}
	suite.Run(t, s)
}

func TestSessionRepository(t *testing.T) {
	s := &testsuite.SessionRepositorySuite{
		Repo: NewSessionRepository(),
	}
	suite.Run(t, s)
}

func TestEventRepository(t *testing.T) {
	s := &testsuite.EventRepositorySuite{
		Repo:        NewEventRepository(),
		SessionRepo: NewSessionRepository(),
	}
	suite.Run(t, s)
}

func TestUserRepository(t *testing.T) {
	s := &testsuite.UserRepositorySuite{
		Repo: NewUserRepository(),
	}
	suite.Run(t, s)
}

func TestAPIKeyRepository(t *testing.T) {
	s := &testsuite.APIKeyRepositorySuite{
		Repo: NewAPIKeyRepository(),
	}
	suite.Run(t, s)
}

func TestWebSessionRepository(t *testing.T) {
	s := &testsuite.WebSessionRepositorySuite{
		Repo: NewWebSessionRepository(),
	}
	suite.Run(t, s)
}

func TestPasswordCredentialRepository(t *testing.T) {
	s := &testsuite.PasswordCredentialRepositorySuite{
		Repo: NewPasswordCredentialRepository(),
	}
	suite.Run(t, s)
}

func TestOAuthConnectionRepository(t *testing.T) {
	s := &testsuite.OAuthConnectionRepositorySuite{
		Repo: NewOAuthConnectionRepository(),
	}
	suite.Run(t, s)
}

func TestPlanDocumentRepository(t *testing.T) {
	s := &testsuite.PlanDocumentRepositorySuite{
		Repo: NewPlanDocumentRepository(),
	}
	suite.Run(t, s)
}

func TestPlanDocumentEventRepository(t *testing.T) {
	s := &testsuite.PlanDocumentEventRepositorySuite{
		Repo: NewPlanDocumentEventRepository(),
	}
	suite.Run(t, s)
}

func TestUserFavoriteRepository(t *testing.T) {
	s := &testsuite.UserFavoriteRepositorySuite{
		Repo: NewUserFavoriteRepository(),
	}
	suite.Run(t, s)
}

func TestPlanCommentThreadRepository(t *testing.T) {
	repos := NewRepositories()
	s := &testsuite.PlanCommentThreadRepositorySuite{
		Repo:        repos.PlanCommentThread,
		PlanDocRepo: repos.PlanDocument,
		ProjectRepo: repos.Project,
		UserRepo:    repos.User,
	}
	suite.Run(t, s)
}

func TestPlanCommentMessageRepository(t *testing.T) {
	repos := NewRepositories()
	s := &testsuite.PlanCommentMessageRepositorySuite{
		Repo:        repos.PlanCommentMessage,
		ThreadRepo:  repos.PlanCommentThread,
		PlanDocRepo: repos.PlanDocument,
		ProjectRepo: repos.Project,
		UserRepo:    repos.User,
	}
	suite.Run(t, s)
}
