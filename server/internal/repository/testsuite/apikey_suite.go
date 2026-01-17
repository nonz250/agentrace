package testsuite

import (
	"context"
	"time"

	"github.com/google/uuid"
	"github.com/satetsu888/agentrace/server/internal/domain"
	"github.com/satetsu888/agentrace/server/internal/repository"
	"github.com/stretchr/testify/suite"
)

// APIKeyRepositorySuite tests APIKeyRepository implementations
type APIKeyRepositorySuite struct {
	suite.Suite
	Repo     repository.APIKeyRepository
	UserRepo repository.UserRepository // Optional: for FK constraint support
	Cleanup  func()
}

// createTestUser creates a user for FK constraint tests and returns the auto-generated ID
func (s *APIKeyRepositorySuite) createTestUser(emailPrefix string) string {
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

func (s *APIKeyRepositorySuite) TearDownTest() {
	if s.Cleanup != nil {
		s.Cleanup()
	}
}

func (s *APIKeyRepositorySuite) TestCreate() {
	ctx := context.Background()

	userID := s.createTestUser("apikey-create")
	if userID == "" {
		s.T().Skip("UserRepo not available, skipping test")
	}

	key := &domain.APIKey{
		UserID:    userID,
		Name:      "My API Key",
		KeyHash:   uuid.New().String(),
		KeyPrefix: "agtr_xxxx",
	}

	err := s.Repo.Create(ctx, key)
	s.Require().NoError(err)

	// ID should be auto-generated
	s.NotEmpty(key.ID)

	// CreatedAt should be set
	s.False(key.CreatedAt.IsZero())
}

func (s *APIKeyRepositorySuite) TestFindByKeyHash() {
	ctx := context.Background()

	userID := s.createTestUser("apikey-hash")
	if userID == "" {
		s.T().Skip("UserRepo not available, skipping test")
	}

	uniqueHash := uuid.New().String()
	key := &domain.APIKey{
		UserID:    userID,
		Name:      "Find By Hash Key",
		KeyHash:   uniqueHash,
		KeyPrefix: "agtr_yyyy",
	}
	err := s.Repo.Create(ctx, key)
	s.Require().NoError(err)

	found, err := s.Repo.FindByKeyHash(ctx, uniqueHash)
	s.Require().NoError(err)
	s.Require().NotNil(found)
	s.Equal(key.ID, found.ID)
	s.Equal(key.UserID, found.UserID)
}

func (s *APIKeyRepositorySuite) TestFindByKeyHash_NotFound() {
	ctx := context.Background()

	found, err := s.Repo.FindByKeyHash(ctx, "non-existing-hash-"+uuid.New().String())
	s.NoError(err)
	s.Nil(found)
}

func (s *APIKeyRepositorySuite) TestFindByUserID() {
	ctx := context.Background()

	userID := s.createTestUser("apikey-finduser")
	otherUserID := s.createTestUser("apikey-other")
	if userID == "" || otherUserID == "" {
		s.T().Skip("UserRepo not available, skipping test")
	}

	// Create multiple keys for user
	for i := 0; i < 3; i++ {
		key := &domain.APIKey{
			UserID:    userID,
			Name:      "Key " + string(rune('A'+i)),
			KeyHash:   uuid.New().String(),
			KeyPrefix: "agtr_" + string(rune('a'+i)),
		}
		err := s.Repo.Create(ctx, key)
		s.Require().NoError(err)
	}

	// Create key for different user
	otherKey := &domain.APIKey{
		UserID:    otherUserID,
		Name:      "Other Key",
		KeyHash:   uuid.New().String(),
		KeyPrefix: "agtr_other",
	}
	err := s.Repo.Create(ctx, otherKey)
	s.Require().NoError(err)

	keys, err := s.Repo.FindByUserID(ctx, userID)
	s.Require().NoError(err)
	s.Len(keys, 3)

	for _, k := range keys {
		s.Equal(userID, k.UserID)
	}
}

func (s *APIKeyRepositorySuite) TestFindByID() {
	ctx := context.Background()

	userID := s.createTestUser("apikey-findid")
	if userID == "" {
		s.T().Skip("UserRepo not available, skipping test")
	}

	key := &domain.APIKey{
		UserID:    userID,
		Name:      "Find By ID Key",
		KeyHash:   uuid.New().String(),
		KeyPrefix: "agtr_zzzz",
	}
	err := s.Repo.Create(ctx, key)
	s.Require().NoError(err)

	found, err := s.Repo.FindByID(ctx, key.ID)
	s.Require().NoError(err)
	s.Require().NotNil(found)
	s.Equal(key.ID, found.ID)
}

func (s *APIKeyRepositorySuite) TestFindByID_NotFound() {
	ctx := context.Background()

	nonExistingID := uuid.New().String()
	found, err := s.Repo.FindByID(ctx, nonExistingID)
	s.NoError(err)
	s.Nil(found)
}

func (s *APIKeyRepositorySuite) TestDelete() {
	ctx := context.Background()

	userID := s.createTestUser("apikey-delete")
	if userID == "" {
		s.T().Skip("UserRepo not available, skipping test")
	}

	key := &domain.APIKey{
		UserID:    userID,
		Name:      "Delete Me Key",
		KeyHash:   uuid.New().String(),
		KeyPrefix: "agtr_del",
	}
	err := s.Repo.Create(ctx, key)
	s.Require().NoError(err)

	err = s.Repo.Delete(ctx, key.ID)
	s.Require().NoError(err)

	// Verify deleted
	found, err := s.Repo.FindByID(ctx, key.ID)
	s.NoError(err)
	s.Nil(found)
}

func (s *APIKeyRepositorySuite) TestUpdateLastUsedAt() {
	ctx := context.Background()

	userID := s.createTestUser("apikey-lastused")
	if userID == "" {
		s.T().Skip("UserRepo not available, skipping test")
	}

	key := &domain.APIKey{
		UserID:    userID,
		Name:      "Update LastUsed Key",
		KeyHash:   uuid.New().String(),
		KeyPrefix: "agtr_last",
	}
	err := s.Repo.Create(ctx, key)
	s.Require().NoError(err)

	// Initially LastUsedAt should be nil
	s.Nil(key.LastUsedAt)

	err = s.Repo.UpdateLastUsedAt(ctx, key.ID)
	s.Require().NoError(err)

	found, err := s.Repo.FindByID(ctx, key.ID)
	s.Require().NoError(err)
	s.Require().NotNil(found.LastUsedAt)
	s.WithinDuration(time.Now(), *found.LastUsedAt, 2*time.Second)
}
