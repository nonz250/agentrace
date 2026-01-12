package testsuite

import (
	"context"
	"time"

	"github.com/satetsu888/agentrace/server/internal/domain"
	"github.com/satetsu888/agentrace/server/internal/repository"
	"github.com/stretchr/testify/suite"
)

// WebSessionRepositorySuite tests WebSessionRepository implementations
type WebSessionRepositorySuite struct {
	suite.Suite
	Repo     repository.WebSessionRepository
	UserRepo repository.UserRepository // Optional: for FK constraint support
	Cleanup  func()
}

// createTestUser creates a user for FK constraint tests
func (s *WebSessionRepositorySuite) createTestUser(id string) {
	if s.UserRepo == nil {
		return
	}
	ctx := context.Background()
	user := &domain.User{
		ID:    id,
		Email: id + "@example.com",
	}
	_ = s.UserRepo.Create(ctx, user)
}

func (s *WebSessionRepositorySuite) TearDownTest() {
	if s.Cleanup != nil {
		s.Cleanup()
	}
}

func (s *WebSessionRepositorySuite) TestCreate() {
	ctx := context.Background()

	s.createTestUser("user-1")

	session := &domain.WebSession{
		UserID:    "user-1",
		Token:     "token-123",
		ExpiresAt: time.Now().Add(7 * 24 * time.Hour),
	}

	err := s.Repo.Create(ctx, session)
	s.Require().NoError(err)

	// ID should be auto-generated
	s.NotEmpty(session.ID)

	// CreatedAt should be set
	s.False(session.CreatedAt.IsZero())
}

func (s *WebSessionRepositorySuite) TestFindByToken() {
	ctx := context.Background()

	s.createTestUser("user-2")

	session := &domain.WebSession{
		UserID:    "user-2",
		Token:     "unique-token-456",
		ExpiresAt: time.Now().Add(7 * 24 * time.Hour),
	}
	err := s.Repo.Create(ctx, session)
	s.Require().NoError(err)

	found, err := s.Repo.FindByToken(ctx, "unique-token-456")
	s.Require().NoError(err)
	s.Require().NotNil(found)
	s.Equal(session.ID, found.ID)
	s.Equal(session.UserID, found.UserID)
}

func (s *WebSessionRepositorySuite) TestFindByToken_NotFound() {
	ctx := context.Background()

	found, err := s.Repo.FindByToken(ctx, "non-existing-token")
	s.NoError(err)
	s.Nil(found)
}

func (s *WebSessionRepositorySuite) TestDelete() {
	ctx := context.Background()

	s.createTestUser("user-3")

	session := &domain.WebSession{
		UserID:    "user-3",
		Token:     "delete-token",
		ExpiresAt: time.Now().Add(7 * 24 * time.Hour),
	}
	err := s.Repo.Create(ctx, session)
	s.Require().NoError(err)

	err = s.Repo.Delete(ctx, session.ID)
	s.Require().NoError(err)

	// Verify deleted
	found, err := s.Repo.FindByToken(ctx, "delete-token")
	s.NoError(err)
	s.Nil(found)
}

func (s *WebSessionRepositorySuite) TestDeleteExpired() {
	ctx := context.Background()

	s.createTestUser("user-4")
	s.createTestUser("user-5")

	// Create expired session
	expiredSession := &domain.WebSession{
		UserID:    "user-4",
		Token:     "expired-token",
		ExpiresAt: time.Now().Add(-1 * time.Hour), // Already expired
	}
	err := s.Repo.Create(ctx, expiredSession)
	s.Require().NoError(err)

	// Create valid session
	validSession := &domain.WebSession{
		UserID:    "user-5",
		Token:     "valid-token",
		ExpiresAt: time.Now().Add(7 * 24 * time.Hour),
	}
	err = s.Repo.Create(ctx, validSession)
	s.Require().NoError(err)

	// Delete expired
	err = s.Repo.DeleteExpired(ctx)
	s.Require().NoError(err)

	// Expired should be gone
	found, err := s.Repo.FindByToken(ctx, "expired-token")
	s.NoError(err)
	s.Nil(found)

	// Valid should still exist
	found, err = s.Repo.FindByToken(ctx, "valid-token")
	s.Require().NoError(err)
	s.NotNil(found)
}
