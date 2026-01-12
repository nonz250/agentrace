package testsuite

import (
	"context"

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

// createTestUser creates a user for FK constraint tests
func (s *OAuthConnectionRepositorySuite) createTestUser(id string) {
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

func (s *OAuthConnectionRepositorySuite) TearDownTest() {
	if s.Cleanup != nil {
		s.Cleanup()
	}
}

func (s *OAuthConnectionRepositorySuite) TestCreate() {
	ctx := context.Background()

	s.createTestUser("user-1")

	conn := &domain.OAuthConnection{
		UserID:     "user-1",
		Provider:   "github",
		ProviderID: "github-user-123",
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

	s.createTestUser("user-2")

	conn := &domain.OAuthConnection{
		UserID:     "user-2",
		Provider:   "github",
		ProviderID: "github-user-456",
	}
	err := s.Repo.Create(ctx, conn)
	s.Require().NoError(err)

	found, err := s.Repo.FindByProviderAndProviderID(ctx, "github", "github-user-456")
	s.Require().NoError(err)
	s.Require().NotNil(found)
	s.Equal(conn.ID, found.ID)
	s.Equal(conn.UserID, found.UserID)
}

func (s *OAuthConnectionRepositorySuite) TestFindByProviderAndProviderID_NotFound() {
	ctx := context.Background()

	found, err := s.Repo.FindByProviderAndProviderID(ctx, "github", "non-existing")
	s.NoError(err)
	s.Nil(found)
}

func (s *OAuthConnectionRepositorySuite) TestFindByUserID() {
	ctx := context.Background()

	userID := "user-with-connections"
	s.createTestUser(userID)
	s.createTestUser("other-user")

	// Create multiple connections
	providers := []string{"github", "google"}
	for i, provider := range providers {
		conn := &domain.OAuthConnection{
			UserID:     userID,
			Provider:   provider,
			ProviderID: provider + "-id-" + string(rune('a'+i)),
		}
		err := s.Repo.Create(ctx, conn)
		s.Require().NoError(err)
	}

	// Create connection for different user
	otherConn := &domain.OAuthConnection{
		UserID:     "other-user",
		Provider:   "github",
		ProviderID: "other-github-id",
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

	s.createTestUser("user-3")

	conn := &domain.OAuthConnection{
		UserID:     "user-3",
		Provider:   "github",
		ProviderID: "delete-github-id",
	}
	err := s.Repo.Create(ctx, conn)
	s.Require().NoError(err)

	err = s.Repo.Delete(ctx, conn.ID)
	s.Require().NoError(err)

	// Verify deleted
	found, err := s.Repo.FindByProviderAndProviderID(ctx, "github", "delete-github-id")
	s.NoError(err)
	s.Nil(found)
}
