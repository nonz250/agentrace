package testsuite

import (
	"context"
	"time"

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

// createTestUser creates a user for FK constraint tests
func (s *APIKeyRepositorySuite) createTestUser(id string) {
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

func (s *APIKeyRepositorySuite) TearDownTest() {
	if s.Cleanup != nil {
		s.Cleanup()
	}
}

func (s *APIKeyRepositorySuite) TestCreate() {
	ctx := context.Background()

	s.createTestUser("user-1")

	key := &domain.APIKey{
		UserID:    "user-1",
		Name:      "My API Key",
		KeyHash:   "hash123",
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

	s.createTestUser("user-2")

	key := &domain.APIKey{
		UserID:    "user-2",
		Name:      "Find By Hash Key",
		KeyHash:   "unique-hash-123",
		KeyPrefix: "agtr_yyyy",
	}
	err := s.Repo.Create(ctx, key)
	s.Require().NoError(err)

	found, err := s.Repo.FindByKeyHash(ctx, "unique-hash-123")
	s.Require().NoError(err)
	s.Require().NotNil(found)
	s.Equal(key.ID, found.ID)
	s.Equal(key.UserID, found.UserID)
}

func (s *APIKeyRepositorySuite) TestFindByKeyHash_NotFound() {
	ctx := context.Background()

	found, err := s.Repo.FindByKeyHash(ctx, "non-existing-hash")
	s.NoError(err)
	s.Nil(found)
}

func (s *APIKeyRepositorySuite) TestFindByUserID() {
	ctx := context.Background()

	userID := "user-with-keys"
	s.createTestUser(userID)
	s.createTestUser("other-user")

	// Create multiple keys for user
	for i := 0; i < 3; i++ {
		key := &domain.APIKey{
			UserID:    userID,
			Name:      "Key " + string(rune('A'+i)),
			KeyHash:   "hash-" + string(rune('a'+i)),
			KeyPrefix: "agtr_" + string(rune('a'+i)),
		}
		err := s.Repo.Create(ctx, key)
		s.Require().NoError(err)
	}

	// Create key for different user
	otherKey := &domain.APIKey{
		UserID:    "other-user",
		Name:      "Other Key",
		KeyHash:   "other-hash",
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

	s.createTestUser("user-3")

	key := &domain.APIKey{
		UserID:    "user-3",
		Name:      "Find By ID Key",
		KeyHash:   "findbyid-hash",
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

	found, err := s.Repo.FindByID(ctx, "non-existing-id")
	s.NoError(err)
	s.Nil(found)
}

func (s *APIKeyRepositorySuite) TestDelete() {
	ctx := context.Background()

	s.createTestUser("user-4")

	key := &domain.APIKey{
		UserID:    "user-4",
		Name:      "Delete Me Key",
		KeyHash:   "delete-hash",
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

	s.createTestUser("user-5")

	key := &domain.APIKey{
		UserID:    "user-5",
		Name:      "Update LastUsed Key",
		KeyHash:   "lastused-hash",
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
