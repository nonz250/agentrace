package testsuite

import (
	"context"

	"github.com/satetsu888/agentrace/server/internal/domain"
	"github.com/satetsu888/agentrace/server/internal/repository"
	"github.com/stretchr/testify/suite"
)

// PasswordCredentialRepositorySuite tests PasswordCredentialRepository implementations
type PasswordCredentialRepositorySuite struct {
	suite.Suite
	Repo     repository.PasswordCredentialRepository
	UserRepo repository.UserRepository // Optional: for FK constraint support
	Cleanup  func()
}

// createTestUser creates a user for FK constraint tests
func (s *PasswordCredentialRepositorySuite) createTestUser(id string) {
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

func (s *PasswordCredentialRepositorySuite) TearDownTest() {
	if s.Cleanup != nil {
		s.Cleanup()
	}
}

func (s *PasswordCredentialRepositorySuite) TestCreate() {
	ctx := context.Background()

	s.createTestUser("user-1")

	cred := &domain.PasswordCredential{
		UserID:       "user-1",
		PasswordHash: "hashed-password-123",
	}

	err := s.Repo.Create(ctx, cred)
	s.Require().NoError(err)

	// ID should be auto-generated
	s.NotEmpty(cred.ID)

	// Timestamps should be set
	s.False(cred.CreatedAt.IsZero())
	s.False(cred.UpdatedAt.IsZero())
}

func (s *PasswordCredentialRepositorySuite) TestFindByUserID() {
	ctx := context.Background()

	s.createTestUser("user-with-cred")

	cred := &domain.PasswordCredential{
		UserID:       "user-with-cred",
		PasswordHash: "hashed-password-456",
	}
	err := s.Repo.Create(ctx, cred)
	s.Require().NoError(err)

	found, err := s.Repo.FindByUserID(ctx, "user-with-cred")
	s.Require().NoError(err)
	s.Require().NotNil(found)
	s.Equal(cred.ID, found.ID)
	s.Equal(cred.PasswordHash, found.PasswordHash)
}

func (s *PasswordCredentialRepositorySuite) TestFindByUserID_NotFound() {
	ctx := context.Background()

	found, err := s.Repo.FindByUserID(ctx, "non-existing-user")
	s.NoError(err)
	s.Nil(found)
}

func (s *PasswordCredentialRepositorySuite) TestUpdate() {
	ctx := context.Background()

	s.createTestUser("user-to-update")

	cred := &domain.PasswordCredential{
		UserID:       "user-to-update",
		PasswordHash: "original-hash",
	}
	err := s.Repo.Create(ctx, cred)
	s.Require().NoError(err)

	// Update password hash
	cred.PasswordHash = "updated-hash"
	err = s.Repo.Update(ctx, cred)
	s.Require().NoError(err)

	// Verify update
	found, err := s.Repo.FindByUserID(ctx, "user-to-update")
	s.Require().NoError(err)
	s.Equal("updated-hash", found.PasswordHash)
}

func (s *PasswordCredentialRepositorySuite) TestDelete() {
	ctx := context.Background()

	s.createTestUser("user-to-delete")

	cred := &domain.PasswordCredential{
		UserID:       "user-to-delete",
		PasswordHash: "delete-hash",
	}
	err := s.Repo.Create(ctx, cred)
	s.Require().NoError(err)

	err = s.Repo.Delete(ctx, cred.ID)
	s.Require().NoError(err)

	// Verify deleted
	found, err := s.Repo.FindByUserID(ctx, "user-to-delete")
	s.NoError(err)
	s.Nil(found)
}
