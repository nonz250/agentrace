package testsuite

import (
	"context"

	"github.com/google/uuid"
	"github.com/satetsu888/agentrace/server/internal/domain"
	"github.com/satetsu888/agentrace/server/internal/repository"
	"github.com/stretchr/testify/suite"
)

// OAuthConnectionRepositorySuite tests OAuthConnectionRepository implementations
type OAuthConnectionRepositorySuite struct {
	suite.Suite
	Repo     repository.OAuthConnectionRepository
	UserRepo repository.UserRepository // Optional: for FK constraint support
	Cleanup  func()
}

// createTestUser creates a user for FK constraint tests and returns the auto-generated ID
func (s *OAuthConnectionRepositorySuite) createTestUser(emailPrefix string) string {
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

func (s *OAuthConnectionRepositorySuite) TearDownTest() {
	if s.Cleanup != nil {
		s.Cleanup()
	}
}

func (s *OAuthConnectionRepositorySuite) TestCreate() {
	ctx := context.Background()

	userID := s.createTestUser("oauth-create")
	if userID == "" {
		s.T().Skip("UserRepo not available, skipping test")
	}

	conn := &domain.OAuthConnection{
		UserID:     userID,
		Provider:   "github",
		ProviderID: "github-user-" + uuid.New().String(),
	}

	err := s.Repo.Create(ctx, conn)
	s.Require().NoError(err)

	// ID should be auto-generated
	s.NotEmpty(conn.ID)

	// CreatedAt should be set
	s.False(conn.CreatedAt.IsZero())
}

func (s *OAuthConnectionRepositorySuite) TestFindByProviderAndProviderID() {
	ctx := context.Background()

	userID := s.createTestUser("oauth-find")
	if userID == "" {
		s.T().Skip("UserRepo not available, skipping test")
	}

	providerID := "github-user-" + uuid.New().String()
	conn := &domain.OAuthConnection{
		UserID:     userID,
		Provider:   "github",
		ProviderID: providerID,
	}
	err := s.Repo.Create(ctx, conn)
	s.Require().NoError(err)

	found, err := s.Repo.FindByProviderAndProviderID(ctx, "github", providerID)
	s.Require().NoError(err)
	s.Require().NotNil(found)
	s.Equal(conn.ID, found.ID)
	s.Equal(conn.UserID, found.UserID)
}

func (s *OAuthConnectionRepositorySuite) TestFindByProviderAndProviderID_NotFound() {
	ctx := context.Background()

	found, err := s.Repo.FindByProviderAndProviderID(ctx, "github", "non-existing-"+uuid.New().String())
	s.NoError(err)
	s.Nil(found)
}

func (s *OAuthConnectionRepositorySuite) TestFindByUserID() {
	ctx := context.Background()

	userID := s.createTestUser("oauth-finduser")
	otherUserID := s.createTestUser("oauth-other")
	if userID == "" || otherUserID == "" {
		s.T().Skip("UserRepo not available, skipping test")
	}

	// Create multiple connections
	providers := []string{"github", "google"}
	for i, provider := range providers {
		conn := &domain.OAuthConnection{
			UserID:     userID,
			Provider:   provider,
			ProviderID: provider + "-id-" + string(rune('a'+i)) + "-" + uuid.New().String(),
		}
		err := s.Repo.Create(ctx, conn)
		s.Require().NoError(err)
	}

	// Create connection for different user
	otherConn := &domain.OAuthConnection{
		UserID:     otherUserID,
		Provider:   "github",
		ProviderID: "other-github-id-" + uuid.New().String(),
	}
	err := s.Repo.Create(ctx, otherConn)
	s.Require().NoError(err)

	conns, err := s.Repo.FindByUserID(ctx, userID)
	s.Require().NoError(err)
	s.Len(conns, 2)

	for _, c := range conns {
		s.Equal(userID, c.UserID)
	}
}

func (s *OAuthConnectionRepositorySuite) TestDelete() {
	ctx := context.Background()

	userID := s.createTestUser("oauth-delete")
	if userID == "" {
		s.T().Skip("UserRepo not available, skipping test")
	}

	providerID := "delete-github-id-" + uuid.New().String()
	conn := &domain.OAuthConnection{
		UserID:     userID,
		Provider:   "github",
		ProviderID: providerID,
	}
	err := s.Repo.Create(ctx, conn)
	s.Require().NoError(err)

	err = s.Repo.Delete(ctx, conn.ID)
	s.Require().NoError(err)

	// Verify deleted
	found, err := s.Repo.FindByProviderAndProviderID(ctx, "github", providerID)
	s.NoError(err)
	s.Nil(found)
}
