package testsuite

import (
	"context"

	"github.com/google/uuid"
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

// createTestUser creates a user for FK constraint tests and returns the auto-generated ID
func (s *UserFavoriteRepositorySuite) createTestUser(emailPrefix string) string {
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

func (s *UserFavoriteRepositorySuite) TearDownTest() {
	if s.Cleanup != nil {
		s.Cleanup()
	}
}

func (s *UserFavoriteRepositorySuite) TestCreate() {
	ctx := context.Background()

	userID := s.createTestUser("fav-create")
	if userID == "" {
		s.T().Skip("UserRepo not available, skipping test")
	}

	fav := &domain.UserFavorite{
		UserID:     userID,
		TargetType: domain.UserFavoriteTargetTypeSession,
		TargetID:   uuid.New().String(),
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

	userID := s.createTestUser("fav-delete")
	if userID == "" {
		s.T().Skip("UserRepo not available, skipping test")
	}

	targetID := uuid.New().String()
	fav := &domain.UserFavorite{
		UserID:     userID,
		TargetType: domain.UserFavoriteTargetTypeSession,
		TargetID:   targetID,
	}
	err := s.Repo.Create(ctx, fav)
	s.Require().NoError(err)

	err = s.Repo.Delete(ctx, fav.ID)
	s.Require().NoError(err)

	// Verify deleted
	found, err := s.Repo.FindByUserAndTarget(ctx, userID, domain.UserFavoriteTargetTypeSession, targetID)
	s.NoError(err)
	s.Nil(found)
}

func (s *UserFavoriteRepositorySuite) TestDeleteByUserAndTarget() {
	ctx := context.Background()

	userID := s.createTestUser("fav-del-target")
	if userID == "" {
		s.T().Skip("UserRepo not available, skipping test")
	}

	targetID := uuid.New().String()
	fav := &domain.UserFavorite{
		UserID:     userID,
		TargetType: domain.UserFavoriteTargetTypePlan,
		TargetID:   targetID,
	}
	err := s.Repo.Create(ctx, fav)
	s.Require().NoError(err)

	err = s.Repo.DeleteByUserAndTarget(ctx, userID, domain.UserFavoriteTargetTypePlan, targetID)
	s.Require().NoError(err)

	// Verify deleted
	found, err := s.Repo.FindByUserAndTarget(ctx, userID, domain.UserFavoriteTargetTypePlan, targetID)
	s.NoError(err)
	s.Nil(found)
}

func (s *UserFavoriteRepositorySuite) TestFindByUserID() {
	ctx := context.Background()

	userID := s.createTestUser("fav-findall")
	otherUserID := s.createTestUser("fav-other")
	if userID == "" || otherUserID == "" {
		s.T().Skip("UserRepo not available, skipping test")
	}

	// Create favorites for user
	targets := []struct {
		targetType domain.UserFavoriteTargetType
		targetID   string
	}{
		{domain.UserFavoriteTargetTypeSession, uuid.New().String()},
		{domain.UserFavoriteTargetTypeSession, uuid.New().String()},
		{domain.UserFavoriteTargetTypePlan, uuid.New().String()},
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
		UserID:     otherUserID,
		TargetType: domain.UserFavoriteTargetTypeSession,
		TargetID:   uuid.New().String(),
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

	userID := s.createTestUser("fav-findtype")
	if userID == "" {
		s.T().Skip("UserRepo not available, skipping test")
	}

	// Create favorites with different types
	targets := []struct {
		targetType domain.UserFavoriteTargetType
		targetID   string
	}{
		{domain.UserFavoriteTargetTypeSession, uuid.New().String()},
		{domain.UserFavoriteTargetTypeSession, uuid.New().String()},
		{domain.UserFavoriteTargetTypePlan, uuid.New().String()},
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

	userID := s.createTestUser("fav-findtarget")
	if userID == "" {
		s.T().Skip("UserRepo not available, skipping test")
	}

	targetID := uuid.New().String()
	fav := &domain.UserFavorite{
		UserID:     userID,
		TargetType: domain.UserFavoriteTargetTypeSession,
		TargetID:   targetID,
	}
	err := s.Repo.Create(ctx, fav)
	s.Require().NoError(err)

	found, err := s.Repo.FindByUserAndTarget(ctx, userID, domain.UserFavoriteTargetTypeSession, targetID)
	s.Require().NoError(err)
	s.Require().NotNil(found)
	s.Equal(fav.ID, found.ID)
}

func (s *UserFavoriteRepositorySuite) TestFindByUserAndTarget_NotFound() {
	ctx := context.Background()

	nonExistingUserID := uuid.New().String()
	found, err := s.Repo.FindByUserAndTarget(ctx, nonExistingUserID, domain.UserFavoriteTargetTypeSession, uuid.New().String())
	s.NoError(err)
	s.Nil(found)
}

func (s *UserFavoriteRepositorySuite) TestGetTargetIDs() {
	ctx := context.Background()

	userID := s.createTestUser("fav-getids")
	if userID == "" {
		s.T().Skip("UserRepo not available, skipping test")
	}

	// Create favorites
	sessionIDs := []string{uuid.New().String(), uuid.New().String(), uuid.New().String()}
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
		TargetID:   uuid.New().String(),
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

	nonExistingUserID := uuid.New().String()
	targetIDs, err := s.Repo.GetTargetIDs(ctx, nonExistingUserID, domain.UserFavoriteTargetTypeSession)
	s.Require().NoError(err)
	s.Empty(targetIDs)
}
