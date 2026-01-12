package testsuite

import (
	"context"
	"time"

	"github.com/satetsu888/agentrace/server/internal/domain"
	"github.com/satetsu888/agentrace/server/internal/repository"
	"github.com/stretchr/testify/suite"
)

// PlanDocumentEventRepositorySuite tests PlanDocumentEventRepository implementations
type PlanDocumentEventRepositorySuite struct {
	suite.Suite
	Repo            repository.PlanDocumentEventRepository
	PlanDocRepo     repository.PlanDocumentRepository // Optional: for FK constraint support
	ProjectRepo     repository.ProjectRepository      // Optional: for FK constraint support
	UserRepo        repository.UserRepository         // Optional: for FK constraint support
	Cleanup         func()
	planDocCounter  int
	projectCreated  map[string]bool
	planDocCreated  map[string]bool
	userCreated     map[string]bool
}

// createTestProject creates a project for FK constraint tests
func (s *PlanDocumentEventRepositorySuite) createTestProject(id string) {
	if s.ProjectRepo == nil {
		return
	}
	if s.projectCreated == nil {
		s.projectCreated = make(map[string]bool)
	}
	if s.projectCreated[id] {
		return
	}
	ctx := context.Background()
	project := &domain.Project{
		ID:                     id,
		CanonicalGitRepository: "https://github.com/test/" + id,
	}
	_ = s.ProjectRepo.Create(ctx, project)
	s.projectCreated[id] = true
}

// createTestUser creates a user for FK constraint tests
func (s *PlanDocumentEventRepositorySuite) createTestUser(id string) {
	if s.UserRepo == nil {
		return
	}
	if s.userCreated == nil {
		s.userCreated = make(map[string]bool)
	}
	if s.userCreated[id] {
		return
	}
	ctx := context.Background()
	user := &domain.User{
		ID:    id,
		Email: id + "@example.com",
	}
	_ = s.UserRepo.Create(ctx, user)
	s.userCreated[id] = true
}

// createTestPlanDocument creates a plan document for FK constraint tests
func (s *PlanDocumentEventRepositorySuite) createTestPlanDocument(id string) {
	if s.PlanDocRepo == nil {
		return
	}
	if s.planDocCreated == nil {
		s.planDocCreated = make(map[string]bool)
	}
	if s.planDocCreated[id] {
		return
	}
	ctx := context.Background()
	s.planDocCounter++
	projectID := "test-project-" + id
	s.createTestProject(projectID)
	doc := &domain.PlanDocument{
		ID:          id,
		ProjectID:   projectID,
		Description: "Test Plan " + id,
		Body:        "Body",
		Status:      domain.PlanDocumentStatusPlanning,
	}
	_ = s.PlanDocRepo.Create(ctx, doc)
	s.planDocCreated[id] = true
}

func (s *PlanDocumentEventRepositorySuite) TearDownTest() {
	if s.Cleanup != nil {
		s.Cleanup()
	}
}

func (s *PlanDocumentEventRepositorySuite) TestCreate() {
	ctx := context.Background()

	s.createTestPlanDocument("plan-1")
	s.createTestUser("user-1")

	claudeSessionID := "claude-session-1"
	toolUseID := "toolu_01ABC"
	userID := "user-1"

	event := &domain.PlanDocumentEvent{
		PlanDocumentID:  "plan-1",
		ClaudeSessionID: &claudeSessionID,
		ToolUseID:       &toolUseID,
		UserID:          &userID,
		EventType:       domain.PlanDocumentEventTypeBodyChange,
		Patch:           "@@ -1,3 +1,4 @@\n+Added line",
	}

	err := s.Repo.Create(ctx, event)
	s.Require().NoError(err)

	// ID should be auto-generated
	s.NotEmpty(event.ID)

	// CreatedAt should be set
	s.False(event.CreatedAt.IsZero())
}

func (s *PlanDocumentEventRepositorySuite) TestCreate_StatusChange() {
	ctx := context.Background()

	s.createTestPlanDocument("plan-status")

	event := &domain.PlanDocumentEvent{
		PlanDocumentID: "plan-status",
		EventType:      domain.PlanDocumentEventTypeStatusChange,
		Patch:          "planning -> implementation",
	}

	err := s.Repo.Create(ctx, event)
	s.Require().NoError(err)
	s.NotEmpty(event.ID)
}

func (s *PlanDocumentEventRepositorySuite) TestCreate_WithMessage() {
	ctx := context.Background()

	s.createTestPlanDocument("plan-with-message")

	event := &domain.PlanDocumentEvent{
		PlanDocumentID: "plan-with-message",
		EventType:      domain.PlanDocumentEventTypeBodyChange,
		Patch:          "@@ -1,3 +1,4 @@\n+Added line",
		Message:        "Added background section",
	}

	err := s.Repo.Create(ctx, event)
	s.Require().NoError(err)
	s.NotEmpty(event.ID)

	// Retrieve and verify message is persisted
	events, err := s.Repo.FindByPlanDocumentID(ctx, "plan-with-message")
	s.Require().NoError(err)
	s.Require().Len(events, 1)
	s.Equal("Added background section", events[0].Message)
}

func (s *PlanDocumentEventRepositorySuite) TestFindByPlanDocumentID() {
	ctx := context.Background()

	planID := "plan-find"
	s.createTestPlanDocument(planID)
	s.createTestPlanDocument("other-plan")

	// Create multiple events
	for i := 0; i < 5; i++ {
		event := &domain.PlanDocumentEvent{
			PlanDocumentID: planID,
			EventType:      domain.PlanDocumentEventTypeBodyChange,
			Patch:          "Patch " + string(rune('a'+i)),
		}
		time.Sleep(1 * time.Millisecond)
		err := s.Repo.Create(ctx, event)
		s.Require().NoError(err)
	}

	// Create event for different plan
	otherEvent := &domain.PlanDocumentEvent{
		PlanDocumentID: "other-plan",
		EventType:      domain.PlanDocumentEventTypeBodyChange,
		Patch:          "Other patch",
	}
	err := s.Repo.Create(ctx, otherEvent)
	s.Require().NoError(err)

	events, err := s.Repo.FindByPlanDocumentID(ctx, planID)
	s.Require().NoError(err)
	s.Len(events, 5)

	for _, e := range events {
		s.Equal(planID, e.PlanDocumentID)
	}
}

func (s *PlanDocumentEventRepositorySuite) TestFindByClaudeSessionID() {
	ctx := context.Background()

	claudeSessionID := "claude-session-find"

	// Create events for claude session
	for i := 0; i < 3; i++ {
		planID := "plan-cs-" + string(rune('a'+i))
		s.createTestPlanDocument(planID)
		event := &domain.PlanDocumentEvent{
			PlanDocumentID:  planID,
			ClaudeSessionID: &claudeSessionID,
			EventType:       domain.PlanDocumentEventTypeBodyChange,
			Patch:           "Patch " + string(rune('a'+i)),
		}
		err := s.Repo.Create(ctx, event)
		s.Require().NoError(err)
	}

	// Create event for different claude session
	s.createTestPlanDocument("plan-other")
	otherSessionID := "other-claude-session"
	otherEvent := &domain.PlanDocumentEvent{
		PlanDocumentID:  "plan-other",
		ClaudeSessionID: &otherSessionID,
		EventType:       domain.PlanDocumentEventTypeBodyChange,
		Patch:           "Other patch",
	}
	err := s.Repo.Create(ctx, otherEvent)
	s.Require().NoError(err)

	events, err := s.Repo.FindByClaudeSessionID(ctx, claudeSessionID)
	s.Require().NoError(err)
	s.Len(events, 3)

	for _, e := range events {
		s.Require().NotNil(e.ClaudeSessionID)
		s.Equal(claudeSessionID, *e.ClaudeSessionID)
	}
}

func (s *PlanDocumentEventRepositorySuite) TestGetCollaboratorUserIDs() {
	ctx := context.Background()

	planID := "plan-collaborators"
	s.createTestPlanDocument(planID)
	s.createTestUser("user-a")
	s.createTestUser("user-b")
	s.createTestUser("user-c")

	// Create events from different users
	userIDs := []string{"user-a", "user-b", "user-a", "user-c"} // user-a appears twice
	for i, userID := range userIDs {
		uid := userID
		event := &domain.PlanDocumentEvent{
			PlanDocumentID: planID,
			UserID:         &uid,
			EventType:      domain.PlanDocumentEventTypeBodyChange,
			Patch:          "Patch " + string(rune('a'+i)),
		}
		err := s.Repo.Create(ctx, event)
		s.Require().NoError(err)
	}

	collaborators, err := s.Repo.GetCollaboratorUserIDs(ctx, planID)
	s.Require().NoError(err)

	// Should return unique user IDs (3 unique users)
	s.Len(collaborators, 3)

	// Should contain all unique users
	collaboratorSet := make(map[string]bool)
	for _, c := range collaborators {
		collaboratorSet[c] = true
	}
	s.True(collaboratorSet["user-a"])
	s.True(collaboratorSet["user-b"])
	s.True(collaboratorSet["user-c"])
}

func (s *PlanDocumentEventRepositorySuite) TestGetPlanDocumentIDsByUserIDs() {
	ctx := context.Background()

	s.createTestUser("user-x")
	s.createTestUser("user-y")
	s.createTestUser("user-z")

	// Create events for different users on different plans
	userData := []struct {
		planID string
		userID string
	}{
		{"plan-u1", "user-x"},
		{"plan-u2", "user-x"},
		{"plan-u3", "user-y"},
		{"plan-u4", "user-z"},
		{"plan-u1", "user-y"}, // user-y also on plan-u1
	}

	for _, data := range userData {
		s.createTestPlanDocument(data.planID)
		uid := data.userID
		event := &domain.PlanDocumentEvent{
			PlanDocumentID: data.planID,
			UserID:         &uid,
			EventType:      domain.PlanDocumentEventTypeBodyChange,
			Patch:          "Patch",
		}
		err := s.Repo.Create(ctx, event)
		s.Require().NoError(err)
	}

	// Get plans for user-x and user-y
	planIDs, err := s.Repo.GetPlanDocumentIDsByUserIDs(ctx, []string{"user-x", "user-y"})
	s.Require().NoError(err)

	// Should return plan-u1, plan-u2, plan-u3 (unique plans for user-x and user-y)
	s.Len(planIDs, 3)

	planIDSet := make(map[string]bool)
	for _, id := range planIDs {
		planIDSet[id] = true
	}
	s.True(planIDSet["plan-u1"])
	s.True(planIDSet["plan-u2"])
	s.True(planIDSet["plan-u3"])
	s.False(planIDSet["plan-u4"]) // user-z only
}

func (s *PlanDocumentEventRepositorySuite) TestGetPlanDocumentIDsByUserIDs_Empty() {
	ctx := context.Background()

	planIDs, err := s.Repo.GetPlanDocumentIDsByUserIDs(ctx, []string{})
	s.Require().NoError(err)
	s.Empty(planIDs)
}
