package testsuite

import (
	"context"

	"github.com/satetsu888/agentrace/server/internal/domain"
	"github.com/satetsu888/agentrace/server/internal/repository"
	"github.com/stretchr/testify/suite"
)

// UserRepositorySuite tests UserRepository implementations
type UserRepositorySuite struct {
	suite.Suite
	Repo    repository.UserRepository
	Cleanup func()
}

func (s *UserRepositorySuite) TearDownTest() {
	if s.Cleanup != nil {
		s.Cleanup()
	}
}

func (s *UserRepositorySuite) TestCreate() {
	ctx := context.Background()

	user := &domain.User{
		Email:       "test@example.com",
		DisplayName: "Test User",
	}

	err := s.Repo.Create(ctx, user)
	s.Require().NoError(err)

	// ID should be auto-generated
	s.NotEmpty(user.ID)

	// CreatedAt should be set
	s.False(user.CreatedAt.IsZero())
}

func (s *UserRepositorySuite) TestCreate_WithID() {
	ctx := context.Background()

	user := &domain.User{
		ID:          "custom-user-id",
		Email:       "custom@example.com",
		DisplayName: "Custom User",
	}

	err := s.Repo.Create(ctx, user)
	s.Require().NoError(err)

	// ID should remain as specified
	s.Equal("custom-user-id", user.ID)
}

func (s *UserRepositorySuite) TestFindByID() {
	ctx := context.Background()

	user := &domain.User{
		Email:       "findbyid@example.com",
		DisplayName: "FindByID User",
	}
	err := s.Repo.Create(ctx, user)
	s.Require().NoError(err)

	found, err := s.Repo.FindByID(ctx, user.ID)
	s.Require().NoError(err)
	s.Require().NotNil(found)
	s.Equal(user.ID, found.ID)
	s.Equal(user.Email, found.Email)
	s.Equal(user.DisplayName, found.DisplayName)
}

func (s *UserRepositorySuite) TestFindByID_NotFound() {
	ctx := context.Background()

	found, err := s.Repo.FindByID(ctx, "non-existing-id")
	s.NoError(err)
	s.Nil(found)
}

func (s *UserRepositorySuite) TestFindByEmail() {
	ctx := context.Background()

	user := &domain.User{
		Email:       "findbyemail@example.com",
		DisplayName: "FindByEmail User",
	}
	err := s.Repo.Create(ctx, user)
	s.Require().NoError(err)

	found, err := s.Repo.FindByEmail(ctx, "findbyemail@example.com")
	s.Require().NoError(err)
	s.Require().NotNil(found)
	s.Equal(user.ID, found.ID)
}

func (s *UserRepositorySuite) TestFindByEmail_NotFound() {
	ctx := context.Background()

	found, err := s.Repo.FindByEmail(ctx, "nonexisting@example.com")
	s.NoError(err)
	s.Nil(found)
}

func (s *UserRepositorySuite) TestFindAll() {
	ctx := context.Background()

	// Create multiple users
	for i := 0; i < 3; i++ {
		user := &domain.User{
			Email:       "user" + string(rune('a'+i)) + "@example.com",
			DisplayName: "User " + string(rune('A'+i)),
		}
		err := s.Repo.Create(ctx, user)
		s.Require().NoError(err)
	}

	users, err := s.Repo.FindAll(ctx)
	s.Require().NoError(err)
	s.GreaterOrEqual(len(users), 3)
}

func (s *UserRepositorySuite) TestUpdateDisplayName() {
	ctx := context.Background()

	user := &domain.User{
		Email:       "update@example.com",
		DisplayName: "Original Name",
	}
	err := s.Repo.Create(ctx, user)
	s.Require().NoError(err)

	err = s.Repo.UpdateDisplayName(ctx, user.ID, "Updated Name")
	s.Require().NoError(err)

	found, err := s.Repo.FindByID(ctx, user.ID)
	s.Require().NoError(err)
	s.Equal("Updated Name", found.DisplayName)
}
