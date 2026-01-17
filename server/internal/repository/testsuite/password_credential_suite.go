package testsuite

import (
	"context"

	"github.com/google/uuid"
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

// createTestUser creates a user for FK constraint tests and returns the auto-generated ID
func (s *PasswordCredentialRepositorySuite) createTestUser(emailPrefix string) string {
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

func (s *PasswordCredentialRepositorySuite) TearDownTest() {
	if s.Cleanup != nil {
		s.Cleanup()
	}
}

func (s *PasswordCredentialRepositorySuite) TestCreate() {
	ctx := context.Background()

	userID := s.createTestUser("pwcred-create")
	if userID == "" {
		s.T().Skip("UserRepo not available, skipping test")
	}

	cred := &domain.PasswordCredential{
		UserID:       userID,
		PasswordHash: "hashed-password-" + uuid.New().String(),
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

	userID := s.createTestUser("pwcred-find")
	if userID == "" {
		s.T().Skip("UserRepo not available, skipping test")
	}

	passwordHash := "hashed-password-" + uuid.New().String()
	cred := &domain.PasswordCredential{
		UserID:       userID,
		PasswordHash: passwordHash,
	}
	err := s.Repo.Create(ctx, cred)
	s.Require().NoError(err)

	found, err := s.Repo.FindByUserID(ctx, userID)
	s.Require().NoError(err)
	s.Require().NotNil(found)
	s.Equal(cred.ID, found.ID)
	s.Equal(cred.PasswordHash, found.PasswordHash)
}

func (s *PasswordCredentialRepositorySuite) TestFindByUserID_NotFound() {
	ctx := context.Background()

	nonExistingID := uuid.New().String()
	found, err := s.Repo.FindByUserID(ctx, nonExistingID)
	s.NoError(err)
	s.Nil(found)
}

func (s *PasswordCredentialRepositorySuite) TestUpdate() {
	ctx := context.Background()

	userID := s.createTestUser("pwcred-update")
	if userID == "" {
		s.T().Skip("UserRepo not available, skipping test")
	}

	cred := &domain.PasswordCredential{
		UserID:       userID,
		PasswordHash: "original-hash-" + uuid.New().String(),
	}
	err := s.Repo.Create(ctx, cred)
	s.Require().NoError(err)

	// Update password hash
	updatedHash := "updated-hash-" + uuid.New().String()
	cred.PasswordHash = updatedHash
	err = s.Repo.Update(ctx, cred)
	s.Require().NoError(err)

	// Verify update
	found, err := s.Repo.FindByUserID(ctx, userID)
	s.Require().NoError(err)
	s.Equal(updatedHash, found.PasswordHash)
}

func (s *PasswordCredentialRepositorySuite) TestDelete() {
	ctx := context.Background()

	userID := s.createTestUser("pwcred-delete")
	if userID == "" {
		s.T().Skip("UserRepo not available, skipping test")
	}

	cred := &domain.PasswordCredential{
		UserID:       userID,
		PasswordHash: "delete-hash-" + uuid.New().String(),
	}
	err := s.Repo.Create(ctx, cred)
	s.Require().NoError(err)

	err = s.Repo.Delete(ctx, cred.ID)
	s.Require().NoError(err)

	// Verify deleted
	found, err := s.Repo.FindByUserID(ctx, userID)
	s.NoError(err)
	s.Nil(found)
}
