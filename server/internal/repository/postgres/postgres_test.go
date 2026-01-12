//go:build integration

package postgres

import (
	"os"
	"testing"

	"github.com/satetsu888/agentrace/server/internal/repository/testsuite"
	"github.com/stretchr/testify/suite"
)

// testDB creates a PostgreSQL database connection for testing.
// Requires DATABASE_URL environment variable to be set.
// Example: DATABASE_URL=postgres://user:pass@localhost:5432/agentrace_test?sslmode=disable
func testDB(t *testing.T) (*DB, func()) {
	t.Helper()

	databaseURL := os.Getenv("DATABASE_URL")
	if databaseURL == "" {
		t.Skip("DATABASE_URL not set, skipping PostgreSQL integration tests")
	}

	db, err := Open(databaseURL)
	if err != nil {
		t.Fatalf("failed to open database: %v", err)
	}

	// Clean up tables before tests
	cleanup := func() {
		tables := []string{
			"plan_document_events",
			"plan_documents",
			"user_favorites",
			"events",
			"sessions",
			"web_sessions",
			"api_keys",
			"oauth_connections",
			"password_credentials",
			"users",
			"projects",
		}
		for _, table := range tables {
			db.Exec("DELETE FROM " + table + " WHERE id != '00000000-0000-0000-0000-000000000000'")
		}
		db.Close()
	}

	// Clean before test as well
	tables := []string{
		"plan_document_events",
		"plan_documents",
		"user_favorites",
		"events",
		"sessions",
		"web_sessions",
		"api_keys",
		"oauth_connections",
		"password_credentials",
		"users",
		"projects",
	}
	for _, table := range tables {
		db.Exec("DELETE FROM " + table + " WHERE id != '00000000-0000-0000-0000-000000000000'")
	}

	return db, cleanup
}

func TestProjectRepository(t *testing.T) {
	db, cleanup := testDB(t)
	defer cleanup()

	s := &testsuite.ProjectRepositorySuite{
		Repo: NewProjectRepository(db),
	}
	suite.Run(t, s)
}

func TestSessionRepository(t *testing.T) {
	db, cleanup := testDB(t)
	defer cleanup()

	s := &testsuite.SessionRepositorySuite{
		Repo:        NewSessionRepository(db),
		UserRepo:    NewUserRepository(db),
		ProjectRepo: NewProjectRepository(db),
	}
	suite.Run(t, s)
}

func TestEventRepository(t *testing.T) {
	db, cleanup := testDB(t)
	defer cleanup()

	s := &testsuite.EventRepositorySuite{
		Repo:        NewEventRepository(db),
		SessionRepo: NewSessionRepository(db),
	}
	suite.Run(t, s)
}

func TestUserRepository(t *testing.T) {
	db, cleanup := testDB(t)
	defer cleanup()

	s := &testsuite.UserRepositorySuite{
		Repo: NewUserRepository(db),
	}
	suite.Run(t, s)
}

func TestAPIKeyRepository(t *testing.T) {
	db, cleanup := testDB(t)
	defer cleanup()

	s := &testsuite.APIKeyRepositorySuite{
		Repo:     NewAPIKeyRepository(db),
		UserRepo: NewUserRepository(db),
	}
	suite.Run(t, s)
}

func TestWebSessionRepository(t *testing.T) {
	db, cleanup := testDB(t)
	defer cleanup()

	s := &testsuite.WebSessionRepositorySuite{
		Repo:     NewWebSessionRepository(db),
		UserRepo: NewUserRepository(db),
	}
	suite.Run(t, s)
}

func TestPasswordCredentialRepository(t *testing.T) {
	db, cleanup := testDB(t)
	defer cleanup()

	s := &testsuite.PasswordCredentialRepositorySuite{
		Repo:     NewPasswordCredentialRepository(db),
		UserRepo: NewUserRepository(db),
	}
	suite.Run(t, s)
}

func TestOAuthConnectionRepository(t *testing.T) {
	db, cleanup := testDB(t)
	defer cleanup()

	s := &testsuite.OAuthConnectionRepositorySuite{
		Repo:     NewOAuthConnectionRepository(db),
		UserRepo: NewUserRepository(db),
	}
	suite.Run(t, s)
}

func TestPlanDocumentRepository(t *testing.T) {
	db, cleanup := testDB(t)
	defer cleanup()

	s := &testsuite.PlanDocumentRepositorySuite{
		Repo:        NewPlanDocumentRepository(db),
		ProjectRepo: NewProjectRepository(db),
	}
	suite.Run(t, s)
}

func TestPlanDocumentEventRepository(t *testing.T) {
	db, cleanup := testDB(t)
	defer cleanup()

	s := &testsuite.PlanDocumentEventRepositorySuite{
		Repo:        NewPlanDocumentEventRepository(db),
		PlanDocRepo: NewPlanDocumentRepository(db),
		ProjectRepo: NewProjectRepository(db),
		UserRepo:    NewUserRepository(db),
	}
	suite.Run(t, s)
}

func TestUserFavoriteRepository(t *testing.T) {
	db, cleanup := testDB(t)
	defer cleanup()

	s := &testsuite.UserFavoriteRepositorySuite{
		Repo:     NewUserFavoriteRepository(db),
		UserRepo: NewUserRepository(db),
	}
	suite.Run(t, s)
}
