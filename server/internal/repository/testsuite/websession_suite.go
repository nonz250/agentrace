package testsuite

import (
	"context"
	"time"

	"github.com/google/uuid"
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

// createTestUser creates a user for FK constraint tests and returns the auto-generated ID
func (s *WebSessionRepositorySuite) createTestUser(emailPrefix string) string {
	if s.UserRepo == nil {
		return ""
	}
	ctx := context.Background()
	user := &domain.User{
		Email: emailPrefix + "@example.com",
	}
	err := s.UserRepo.Create(ctx, user)
	if err != nil {
		return ""
	}
	return user.ID
}

func (s *WebSessionRepositorySuite) TearDownTest() {
	if s.Cleanup != nil {
		s.Cleanup()
	}
}

func (s *WebSessionRepositorySuite) TestCreate() {
	ctx := context.Background()

	userID := s.createTestUser("websess-create")
	if userID == "" {
		s.T().Skip("UserRepo not available, skipping test")
	}

	session := &domain.WebSession{
		UserID:    userID,
		Token:     "token-" + uuid.New().String(),
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

	userID := s.createTestUser("websess-find")
	if userID == "" {
		s.T().Skip("UserRepo not available, skipping test")
	}

	token := "unique-token-" + uuid.New().String()
	session := &domain.WebSession{
		UserID:    userID,
		Token:     token,
		ExpiresAt: time.Now().Add(7 * 24 * time.Hour),
	}
	err := s.Repo.Create(ctx, session)
	s.Require().NoError(err)

	found, err := s.Repo.FindByToken(ctx, token)
	s.Require().NoError(err)
	s.Require().NotNil(found)
	s.Equal(session.ID, found.ID)
	s.Equal(session.UserID, found.UserID)
}

func (s *WebSessionRepositorySuite) TestFindByToken_NotFound() {
	ctx := context.Background()

	found, err := s.Repo.FindByToken(ctx, "non-existing-token-"+uuid.New().String())
	s.NoError(err)
	s.Nil(found)
}

func (s *WebSessionRepositorySuite) TestDelete() {
	ctx := context.Background()

	userID := s.createTestUser("websess-delete")
	if userID == "" {
		s.T().Skip("UserRepo not available, skipping test")
	}

	token := "delete-token-" + uuid.New().String()
	session := &domain.WebSession{
		UserID:    userID,
		Token:     token,
		ExpiresAt: time.Now().Add(7 * 24 * time.Hour),
	}
	err := s.Repo.Create(ctx, session)
	s.Require().NoError(err)

	err = s.Repo.Delete(ctx, session.ID)
	s.Require().NoError(err)

	// Verify deleted
	found, err := s.Repo.FindByToken(ctx, token)
	s.NoError(err)
	s.Nil(found)
}

func (s *WebSessionRepositorySuite) TestDeleteExpired() {
	ctx := context.Background()

	userID1 := s.createTestUser("websess-expired")
	userID2 := s.createTestUser("websess-valid")
	if userID1 == "" || userID2 == "" {
		s.T().Skip("UserRepo not available, skipping test")
	}

	expiredToken := "expired-token-" + uuid.New().String()
	validToken := "valid-token-" + uuid.New().String()

	// Create expired session
	expiredSession := &domain.WebSession{
		UserID:    userID1,
		Token:     expiredToken,
		ExpiresAt: time.Now().Add(-1 * time.Hour), // Already expired
	}
	err := s.Repo.Create(ctx, expiredSession)
	s.Require().NoError(err)

	// Create valid session
	validSession := &domain.WebSession{
		UserID:    userID2,
		Token:     validToken,
		ExpiresAt: time.Now().Add(7 * 24 * time.Hour),
	}
	err = s.Repo.Create(ctx, validSession)
	s.Require().NoError(err)

	// Delete expired
	err = s.Repo.DeleteExpired(ctx)
	s.Require().NoError(err)

	// Expired should be gone
	found, err := s.Repo.FindByToken(ctx, expiredToken)
	s.NoError(err)
	s.Nil(found)

	// Valid should still exist
	found, err = s.Repo.FindByToken(ctx, validToken)
	s.Require().NoError(err)
	s.NotNil(found)
}
