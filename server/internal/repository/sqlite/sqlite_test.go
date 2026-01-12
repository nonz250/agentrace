package sqlite

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/satetsu888/agentrace/server/internal/repository/testsuite"
	"github.com/stretchr/testify/suite"
)

// testDB creates a temporary SQLite database for testing
func testDB(t *testing.T) (*DB, func()) {
	t.Helper()

	// Create temp directory
	tmpDir, err := os.MkdirTemp("", "sqlite_test_*")
	if err != nil {
		t.Fatalf("failed to create temp dir: %v", err)
	}

	dbPath := filepath.Join(tmpDir, "test.db")
	db, err := Open(dbPath)
	if err != nil {
		os.RemoveAll(tmpDir)
		t.Fatalf("failed to open database: %v", err)
	}

	cleanup := func() {
		db.Close()
		os.RemoveAll(tmpDir)
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
