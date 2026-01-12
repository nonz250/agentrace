package testsuite

import (
	"context"

	"github.com/satetsu888/agentrace/server/internal/domain"
	"github.com/satetsu888/agentrace/server/internal/repository"
	"github.com/stretchr/testify/suite"
)

// UserFavoriteRepositorySuite tests UserFavoriteRepository implementations
type UserFavoriteRepositorySuite struct {
	suite.Suite
	Repo     repository.UserFavoriteRepository
	UserRepo repository.UserRepository // Optional: for FK constraint support
	Cleanup  func()
}

// createTestUser creates a user for FK constraint tests
func (s *UserFavoriteRepositorySuite) createTestUser(id string) {
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

func (s *UserFavoriteRepositorySuite) TearDownTest() {
	if s.Cleanup != nil {
		s.Cleanup()
	}
}

func (s *UserFavoriteRepositorySuite) TestCreate() {
	ctx := context.Background()

	s.createTestUser("user-1")

	fav := &domain.UserFavorite{
		UserID:     "user-1",
		TargetType: domain.UserFavoriteTargetTypeSession,
		TargetID:   "session-1",
	}

	err := s.Repo.Create(ctx, fav)
	s.Require().NoError(err)

	// ID should be auto-generated
	s.NotEmpty(fav.ID)

	// CreatedAt should be set
	s.False(fav.CreatedAt.IsZero())
}

func (s *UserFavoriteRepositorySuite) TestDelete() {
	ctx := context.Background()

	s.createTestUser("user-del")

	fav := &domain.UserFavorite{
		UserID:     "user-del",
		TargetType: domain.UserFavoriteTargetTypeSession,
		TargetID:   "session-del",
	}
	err := s.Repo.Create(ctx, fav)
	s.Require().NoError(err)

	err = s.Repo.Delete(ctx, fav.ID)
	s.Require().NoError(err)

	// Verify deleted
	found, err := s.Repo.FindByUserAndTarget(ctx, "user-del", domain.UserFavoriteTargetTypeSession, "session-del")
	s.NoError(err)
	s.Nil(found)
}

func (s *UserFavoriteRepositorySuite) TestDeleteByUserAndTarget() {
	ctx := context.Background()

	s.createTestUser("user-del-target")

	fav := &domain.UserFavorite{
		UserID:     "user-del-target",
		TargetType: domain.UserFavoriteTargetTypePlan,
		TargetID:   "plan-del",
	}
	err := s.Repo.Create(ctx, fav)
	s.Require().NoError(err)

	err = s.Repo.DeleteByUserAndTarget(ctx, "user-del-target", domain.UserFavoriteTargetTypePlan, "plan-del")
	s.Require().NoError(err)

	// Verify deleted
	found, err := s.Repo.FindByUserAndTarget(ctx, "user-del-target", domain.UserFavoriteTargetTypePlan, "plan-del")
	s.NoError(err)
	s.Nil(found)
}

func (s *UserFavoriteRepositorySuite) TestFindByUserID() {
	ctx := context.Background()

	userID := "user-find-all"
	s.createTestUser(userID)
	s.createTestUser("other-user")

	// Create favorites for user
	targets := []struct {
		targetType domain.UserFavoriteTargetType
		targetID   string
	}{
		{domain.UserFavoriteTargetTypeSession, "session-1"},
		{domain.UserFavoriteTargetTypeSession, "session-2"},
		{domain.UserFavoriteTargetTypePlan, "plan-1"},
	}

	for _, t := range targets {
		fav := &domain.UserFavorite{
			UserID:     userID,
			TargetType: t.targetType,
			TargetID:   t.targetID,
		}
		err := s.Repo.Create(ctx, fav)
		s.Require().NoError(err)
	}

	// Create favorite for different user
	otherFav := &domain.UserFavorite{
		UserID:     "other-user",
		TargetType: domain.UserFavoriteTargetTypeSession,
		TargetID:   "session-other",
	}
	err := s.Repo.Create(ctx, otherFav)
	s.Require().NoError(err)

	favs, err := s.Repo.FindByUserID(ctx, userID)
	s.Require().NoError(err)
	s.Len(favs, 3)

	for _, f := range favs {
		s.Equal(userID, f.UserID)
	}
}

func (s *UserFavoriteRepositorySuite) TestFindByUserAndTargetType() {
	ctx := context.Background()

	userID := "user-find-type"
	s.createTestUser(userID)

	// Create favorites with different types
	targets := []struct {
		targetType domain.UserFavoriteTargetType
		targetID   string
	}{
		{domain.UserFavoriteTargetTypeSession, "session-1"},
		{domain.UserFavoriteTargetTypeSession, "session-2"},
		{domain.UserFavoriteTargetTypePlan, "plan-1"},
	}

	for _, t := range targets {
		fav := &domain.UserFavorite{
			UserID:     userID,
			TargetType: t.targetType,
			TargetID:   t.targetID,
		}
		err := s.Repo.Create(ctx, fav)
		s.Require().NoError(err)
	}

	// Find only sessions
	favs, err := s.Repo.FindByUserAndTargetType(ctx, userID, domain.UserFavoriteTargetTypeSession)
	s.Require().NoError(err)
	s.Len(favs, 2)

	for _, f := range favs {
		s.Equal(domain.UserFavoriteTargetTypeSession, f.TargetType)
	}

	// Find only plans
	favs, err = s.Repo.FindByUserAndTargetType(ctx, userID, domain.UserFavoriteTargetTypePlan)
	s.Require().NoError(err)
	s.Len(favs, 1)
	s.Equal(domain.UserFavoriteTargetTypePlan, favs[0].TargetType)
}

func (s *UserFavoriteRepositorySuite) TestFindByUserAndTarget() {
	ctx := context.Background()

	s.createTestUser("user-find-target")

	fav := &domain.UserFavorite{
		UserID:     "user-find-target",
		TargetType: domain.UserFavoriteTargetTypeSession,
		TargetID:   "specific-session",
	}
	err := s.Repo.Create(ctx, fav)
	s.Require().NoError(err)

	found, err := s.Repo.FindByUserAndTarget(ctx, "user-find-target", domain.UserFavoriteTargetTypeSession, "specific-session")
	s.Require().NoError(err)
	s.Require().NotNil(found)
	s.Equal(fav.ID, found.ID)
}

func (s *UserFavoriteRepositorySuite) TestFindByUserAndTarget_NotFound() {
	ctx := context.Background()

	found, err := s.Repo.FindByUserAndTarget(ctx, "non-existing-user", domain.UserFavoriteTargetTypeSession, "session")
	s.NoError(err)
	s.Nil(found)
}

func (s *UserFavoriteRepositorySuite) TestGetTargetIDs() {
	ctx := context.Background()

	userID := "user-get-ids"
	s.createTestUser(userID)

	// Create favorites
	sessionIDs := []string{"session-a", "session-b", "session-c"}
	for _, id := range sessionIDs {
		fav := &domain.UserFavorite{
			UserID:     userID,
			TargetType: domain.UserFavoriteTargetTypeSession,
			TargetID:   id,
		}
		err := s.Repo.Create(ctx, fav)
		s.Require().NoError(err)
	}

	// Add a plan too
	planFav := &domain.UserFavorite{
		UserID:     userID,
		TargetType: domain.UserFavoriteTargetTypePlan,
		TargetID:   "plan-x",
	}
	err := s.Repo.Create(ctx, planFav)
	s.Require().NoError(err)

	// Get only session IDs
	targetIDs, err := s.Repo.GetTargetIDs(ctx, userID, domain.UserFavoriteTargetTypeSession)
	s.Require().NoError(err)
	s.Len(targetIDs, 3)

	idSet := make(map[string]bool)
	for _, id := range targetIDs {
		idSet[id] = true
	}
	for _, expected := range sessionIDs {
		s.True(idSet[expected])
	}
}

func (s *UserFavoriteRepositorySuite) TestGetTargetIDs_Empty() {
	ctx := context.Background()

	targetIDs, err := s.Repo.GetTargetIDs(ctx, "non-existing-user", domain.UserFavoriteTargetTypeSession)
	s.Require().NoError(err)
	s.Empty(targetIDs)
}
