package testsuite

import (
	"context"
	"time"

	"github.com/google/uuid"
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
	projectCreated  map[string]string // gitRepoSuffix -> projectID
	planDocCreated  map[string]string // planDocSuffix -> planDocID
	userCreated     map[string]string // emailPrefix -> userID
}

// createTestProject creates a project for FK constraint tests and returns the auto-generated ID
func (s *PlanDocumentEventRepositorySuite) createTestProject(gitRepoSuffix string) string {
	if s.ProjectRepo == nil {
		return ""
	}
	if s.projectCreated == nil {
		s.projectCreated = make(map[string]string)
	}
	if id, ok := s.projectCreated[gitRepoSuffix]; ok {
		return id
	}
	ctx := context.Background()
	project := &domain.Project{
		CanonicalGitRepository: "https://github.com/test/" + gitRepoSuffix,
	}
	err := s.ProjectRepo.Create(ctx, project)
	if err != nil {
		return ""
	}
	s.projectCreated[gitRepoSuffix] = project.ID
	return project.ID
}

// createTestUser creates a user for FK constraint tests and returns the auto-generated ID
func (s *PlanDocumentEventRepositorySuite) createTestUser(emailPrefix string) string {
	if s.UserRepo == nil {
		return ""
	}
	if s.userCreated == nil {
		s.userCreated = make(map[string]string)
	}
	if id, ok := s.userCreated[emailPrefix]; ok {
		return id
	}
	ctx := context.Background()
	user := &domain.User{
		Email: emailPrefix + "@example.com",
	}
	err := s.UserRepo.Create(ctx, user)
	if err != nil {
		return ""
	}
	s.userCreated[emailPrefix] = user.ID
	return user.ID
}

// createTestPlanDocument creates a plan document for FK constraint tests and returns the auto-generated ID
func (s *PlanDocumentEventRepositorySuite) createTestPlanDocument(planDocSuffix string) string {
	if s.PlanDocRepo == nil {
		return ""
	}
	if s.planDocCreated == nil {
		s.planDocCreated = make(map[string]string)
	}
	if id, ok := s.planDocCreated[planDocSuffix]; ok {
		return id
	}
	ctx := context.Background()
	projectID := s.createTestProject("project-" + planDocSuffix)
	if projectID == "" {
		return ""
	}
	doc := &domain.PlanDocument{
		ProjectID:   projectID,
		Description: "Test Plan " + planDocSuffix,
		Body:        "Body",
		Status:      domain.PlanDocumentStatusPlanning,
	}
	err := s.PlanDocRepo.Create(ctx, doc)
	if err != nil {
		return ""
	}
	s.planDocCreated[planDocSuffix] = doc.ID
	return doc.ID
}

func (s *PlanDocumentEventRepositorySuite) TearDownTest() {
	if s.Cleanup != nil {
		s.Cleanup()
	}
}

func (s *PlanDocumentEventRepositorySuite) TestCreate() {
	ctx := context.Background()

	planDocID := s.createTestPlanDocument("event-create")
	userID := s.createTestUser("event-create")
	if planDocID == "" {
		s.T().Skip("PlanDocRepo not available, skipping test")
	}

	claudeSessionID := uuid.New().String()
	toolUseID := "toolu_" + uuid.New().String()

	event := &domain.PlanDocumentEvent{
		PlanDocumentID:  planDocID,
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

	planDocID := s.createTestPlanDocument("event-status")
	if planDocID == "" {
		s.T().Skip("PlanDocRepo not available, skipping test")
	}

	event := &domain.PlanDocumentEvent{
		PlanDocumentID: planDocID,
		EventType:      domain.PlanDocumentEventTypeStatusChange,
		Patch:          "planning -> implementation",
	}

	err := s.Repo.Create(ctx, event)
	s.Require().NoError(err)
	s.NotEmpty(event.ID)
}

func (s *PlanDocumentEventRepositorySuite) TestCreate_WithMessage() {
	ctx := context.Background()

	planDocID := s.createTestPlanDocument("event-message")
	if planDocID == "" {
		s.T().Skip("PlanDocRepo not available, skipping test")
	}

	event := &domain.PlanDocumentEvent{
		PlanDocumentID: planDocID,
		EventType:      domain.PlanDocumentEventTypeBodyChange,
		Patch:          "@@ -1,3 +1,4 @@\n+Added line",
		Message:        "Added background section",
	}

	err := s.Repo.Create(ctx, event)
	s.Require().NoError(err)
	s.NotEmpty(event.ID)

	// Retrieve and verify message is persisted
	events, err := s.Repo.FindByPlanDocumentID(ctx, planDocID)
	s.Require().NoError(err)
	s.Require().Len(events, 1)
	s.Equal("Added background section", events[0].Message)
}

func (s *PlanDocumentEventRepositorySuite) TestFindByPlanDocumentID() {
	ctx := context.Background()

	planDocID := s.createTestPlanDocument("event-find")
	otherPlanDocID := s.createTestPlanDocument("event-other")
	if planDocID == "" || otherPlanDocID == "" {
		s.T().Skip("PlanDocRepo not available, skipping test")
	}

	// Create multiple events
	for i := 0; i < 5; i++ {
		event := &domain.PlanDocumentEvent{
			PlanDocumentID: planDocID,
			EventType:      domain.PlanDocumentEventTypeBodyChange,
			Patch:          "Patch " + string(rune('a'+i)),
		}
		time.Sleep(1 * time.Millisecond)
		err := s.Repo.Create(ctx, event)
		s.Require().NoError(err)
	}

	// Create event for different plan
	otherEvent := &domain.PlanDocumentEvent{
		PlanDocumentID: otherPlanDocID,
		EventType:      domain.PlanDocumentEventTypeBodyChange,
		Patch:          "Other patch",
	}
	err := s.Repo.Create(ctx, otherEvent)
	s.Require().NoError(err)

	events, err := s.Repo.FindByPlanDocumentID(ctx, planDocID)
	s.Require().NoError(err)
	s.Len(events, 5)

	for _, e := range events {
		s.Equal(planDocID, e.PlanDocumentID)
	}
}

func (s *PlanDocumentEventRepositorySuite) TestFindByPlanDocumentID_ChronologicalOrder() {
	ctx := context.Background()

	planDocID := s.createTestPlanDocument("event-chrono")
	if planDocID == "" {
		s.T().Skip("PlanDocRepo not available, skipping test")
	}

	baseTime := time.Now()

	// Create events in non-sequential order but with specific timestamps
	// We'll create them out of order to ensure sorting is by CreatedAt, not insertion order
	timestamps := []time.Duration{
		300 * time.Millisecond, // third
		100 * time.Millisecond, // first
		500 * time.Millisecond, // fifth
		200 * time.Millisecond, // second
		400 * time.Millisecond, // fourth
	}

	for i, offset := range timestamps {
		event := &domain.PlanDocumentEvent{
			PlanDocumentID: planDocID,
			EventType:      domain.PlanDocumentEventTypeBodyChange,
			Patch:          "Patch " + string(rune('a'+i)),
			CreatedAt:      baseTime.Add(offset),
		}
		err := s.Repo.Create(ctx, event)
		s.Require().NoError(err)
	}

	// Find events
	events, err := s.Repo.FindByPlanDocumentID(ctx, planDocID)
	s.Require().NoError(err)
	s.Require().Len(events, 5)

	// Verify events are in chronological order (ascending by CreatedAt)
	for i := 1; i < len(events); i++ {
		s.True(
			events[i-1].CreatedAt.Before(events[i].CreatedAt) || events[i-1].CreatedAt.Equal(events[i].CreatedAt),
			"Events should be in chronological order: event[%d].CreatedAt=%v should be <= event[%d].CreatedAt=%v",
			i-1, events[i-1].CreatedAt, i, events[i].CreatedAt,
		)
	}

	// Also verify the first and last events have the expected timestamps
	s.Equal(baseTime.Add(100*time.Millisecond).UnixNano(), events[0].CreatedAt.UnixNano(), "First event should have earliest timestamp")
	s.Equal(baseTime.Add(500*time.Millisecond).UnixNano(), events[4].CreatedAt.UnixNano(), "Last event should have latest timestamp")
}

func (s *PlanDocumentEventRepositorySuite) TestFindByClaudeSessionID() {
	ctx := context.Background()

	claudeSessionID := uuid.New().String()

	// Create events for claude session
	for i := 0; i < 3; i++ {
		planDocID := s.createTestPlanDocument("event-cs-" + string(rune('a'+i)))
		if planDocID == "" {
			s.T().Skip("PlanDocRepo not available, skipping test")
		}
		event := &domain.PlanDocumentEvent{
			PlanDocumentID:  planDocID,
			ClaudeSessionID: &claudeSessionID,
			EventType:       domain.PlanDocumentEventTypeBodyChange,
			Patch:           "Patch " + string(rune('a'+i)),
		}
		err := s.Repo.Create(ctx, event)
		s.Require().NoError(err)
	}

	// Create event for different claude session
	otherPlanDocID := s.createTestPlanDocument("event-cs-other")
	if otherPlanDocID == "" {
		s.T().Skip("PlanDocRepo not available, skipping test")
	}
	otherSessionID := uuid.New().String()
	otherEvent := &domain.PlanDocumentEvent{
		PlanDocumentID:  otherPlanDocID,
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

	planDocID := s.createTestPlanDocument("event-collab")
	userA := s.createTestUser("collab-a")
	userB := s.createTestUser("collab-b")
	userC := s.createTestUser("collab-c")
	if planDocID == "" {
		s.T().Skip("PlanDocRepo not available, skipping test")
	}

	// Create events from different users
	userIDs := []string{userA, userB, userA, userC} // user-a appears twice
	for i, userID := range userIDs {
		uid := userID
		event := &domain.PlanDocumentEvent{
			PlanDocumentID: planDocID,
			UserID:         &uid,
			EventType:      domain.PlanDocumentEventTypeBodyChange,
			Patch:          "Patch " + string(rune('a'+i)),
		}
		err := s.Repo.Create(ctx, event)
		s.Require().NoError(err)
	}

	collaborators, err := s.Repo.GetCollaboratorUserIDs(ctx, planDocID)
	s.Require().NoError(err)

	// Should return unique user IDs (3 unique users)
	s.Len(collaborators, 3)

	// Should contain all unique users
	collaboratorSet := make(map[string]bool)
	for _, c := range collaborators {
		collaboratorSet[c] = true
	}
	s.True(collaboratorSet[userA])
	s.True(collaboratorSet[userB])
	s.True(collaboratorSet[userC])
}

func (s *PlanDocumentEventRepositorySuite) TestGetPlanDocumentIDsByUserIDs() {
	ctx := context.Background()

	userX := s.createTestUser("plandoc-x")
	userY := s.createTestUser("plandoc-y")
	userZ := s.createTestUser("plandoc-z")

	// Create events for different users on different plans
	planU1 := s.createTestPlanDocument("plandoc-u1")
	planU2 := s.createTestPlanDocument("plandoc-u2")
	planU3 := s.createTestPlanDocument("plandoc-u3")
	planU4 := s.createTestPlanDocument("plandoc-u4")
	if planU1 == "" || planU2 == "" || planU3 == "" || planU4 == "" {
		s.T().Skip("PlanDocRepo not available, skipping test")
	}

	userData := []struct {
		planID string
		userID string
	}{
		{planU1, userX},
		{planU2, userX},
		{planU3, userY},
		{planU4, userZ},
		{planU1, userY}, // user-y also on plan-u1
	}

	for _, data := range userData {
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
	planIDs, err := s.Repo.GetPlanDocumentIDsByUserIDs(ctx, []string{userX, userY})
	s.Require().NoError(err)

	// Should return plan-u1, plan-u2, plan-u3 (unique plans for user-x and user-y)
	s.Len(planIDs, 3)

	planIDSet := make(map[string]bool)
	for _, id := range planIDs {
		planIDSet[id] = true
	}
	s.True(planIDSet[planU1])
	s.True(planIDSet[planU2])
	s.True(planIDSet[planU3])
	s.False(planIDSet[planU4]) // user-z only
}

func (s *PlanDocumentEventRepositorySuite) TestGetPlanDocumentIDsByUserIDs_Empty() {
	ctx := context.Background()

	planIDs, err := s.Repo.GetPlanDocumentIDsByUserIDs(ctx, []string{})
	s.Require().NoError(err)
	s.Empty(planIDs)
}
