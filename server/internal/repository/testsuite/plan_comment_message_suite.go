package testsuite

import (
	"context"

	"github.com/satetsu888/agentrace/server/internal/domain"
	"github.com/satetsu888/agentrace/server/internal/repository"
	"github.com/stretchr/testify/suite"
)

// PlanCommentMessageRepositorySuite tests PlanCommentMessageRepository implementations
type PlanCommentMessageRepositorySuite struct {
	suite.Suite
	Repo           repository.PlanCommentMessageRepository
	ThreadRepo     repository.PlanCommentThreadRepository
	PlanDocRepo    repository.PlanDocumentRepository
	ProjectRepo    repository.ProjectRepository
	UserRepo       repository.UserRepository
	Cleanup        func()
	projectCreated map[string]string
	planDocCreated map[string]string
	threadCreated  map[string]string
	userCreated    map[string]string
}

func (s *PlanCommentMessageRepositorySuite) createTestProject(gitRepoSuffix string) string {
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

func (s *PlanCommentMessageRepositorySuite) createTestUser(emailPrefix string) string {
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
		Email:       emailPrefix + "@example.com",
		DisplayName: emailPrefix,
	}
	err := s.UserRepo.Create(ctx, user)
	if err != nil {
		return ""
	}
	s.userCreated[emailPrefix] = user.ID
	return user.ID
}

func (s *PlanCommentMessageRepositorySuite) createTestPlanDocument(planDocSuffix string) string {
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
		Body:        "Test body content for " + planDocSuffix,
		Status:      domain.PlanDocumentStatusPlanning,
	}
	err := s.PlanDocRepo.Create(ctx, doc)
	if err != nil {
		return ""
	}
	s.planDocCreated[planDocSuffix] = doc.ID
	return doc.ID
}

func (s *PlanCommentMessageRepositorySuite) createTestThread(threadSuffix string) string {
	if s.ThreadRepo == nil {
		return ""
	}
	if s.threadCreated == nil {
		s.threadCreated = make(map[string]string)
	}
	if id, ok := s.threadCreated[threadSuffix]; ok {
		return id
	}
	ctx := context.Background()
	planDocID := s.createTestPlanDocument("plandoc-" + threadSuffix)
	if planDocID == "" {
		return ""
	}
	thread := &domain.PlanCommentThread{
		PlanDocumentID: planDocID,
		TargetText:     "target text for " + threadSuffix,
		Status:         domain.PlanCommentThreadStatusActive,
	}
	err := s.ThreadRepo.Create(ctx, thread)
	if err != nil {
		return ""
	}
	s.threadCreated[threadSuffix] = thread.ID
	return thread.ID
}

func (s *PlanCommentMessageRepositorySuite) TearDownTest() {
	if s.Cleanup != nil {
		s.Cleanup()
	}
}

func (s *PlanCommentMessageRepositorySuite) TestCreate() {
	ctx := context.Background()

	threadID := s.createTestThread("msg-create")
	userID := s.createTestUser("msg-create")
	if threadID == "" {
		s.T().Skip("ThreadRepo not available, skipping test")
	}

	message := &domain.PlanCommentMessage{
		ThreadID: threadID,
		UserID:   userID,
		Content:  "test message content",
	}

	err := s.Repo.Create(ctx, message)
	s.Require().NoError(err)
	s.NotEmpty(message.ID)
	s.False(message.CreatedAt.IsZero())
	s.False(message.UpdatedAt.IsZero())
}

func (s *PlanCommentMessageRepositorySuite) TestFindByID() {
	ctx := context.Background()

	threadID := s.createTestThread("msg-findbyid")
	userID := s.createTestUser("msg-findbyid")
	if threadID == "" {
		s.T().Skip("ThreadRepo not available, skipping test")
	}

	message := &domain.PlanCommentMessage{
		ThreadID: threadID,
		UserID:   userID,
		Content:  "findbyid message content",
	}
	err := s.Repo.Create(ctx, message)
	s.Require().NoError(err)

	found, err := s.Repo.FindByID(ctx, message.ID)
	s.Require().NoError(err)
	s.Equal(message.ID, found.ID)
	s.Equal(threadID, found.ThreadID)
	s.Equal("findbyid message content", found.Content)
}

func (s *PlanCommentMessageRepositorySuite) TestFindByID_NotFound() {
	ctx := context.Background()

	// Use valid UUID format for PostgreSQL compatibility
	found, err := s.Repo.FindByID(ctx, "00000000-0000-0000-0000-000000000000")
	s.NoError(err)
	s.Nil(found)
}

func (s *PlanCommentMessageRepositorySuite) TestFindByThreadID() {
	ctx := context.Background()

	threadID := s.createTestThread("msg-findbythread")
	otherThreadID := s.createTestThread("msg-findbythread-other")
	userID := s.createTestUser("msg-findbythread")
	if threadID == "" || otherThreadID == "" {
		s.T().Skip("ThreadRepo not available, skipping test")
	}

	// Create messages for target thread
	for i := 0; i < 3; i++ {
		message := &domain.PlanCommentMessage{
			ThreadID: threadID,
			UserID:   userID,
			Content:  "message " + string(rune('a'+i)),
		}
		err := s.Repo.Create(ctx, message)
		s.Require().NoError(err)
	}

	// Create message for other thread
	otherMessage := &domain.PlanCommentMessage{
		ThreadID: otherThreadID,
		UserID:   userID,
		Content:  "other message",
	}
	err := s.Repo.Create(ctx, otherMessage)
	s.Require().NoError(err)

	messages, err := s.Repo.FindByThreadID(ctx, threadID)
	s.Require().NoError(err)
	s.Len(messages, 3)

	for _, m := range messages {
		s.Equal(threadID, m.ThreadID)
	}
}

func (s *PlanCommentMessageRepositorySuite) TestFindByThreadID_OrderedByCreatedAt() {
	ctx := context.Background()

	threadID := s.createTestThread("msg-ordered")
	userID := s.createTestUser("msg-ordered")
	if threadID == "" {
		s.T().Skip("ThreadRepo not available, skipping test")
	}

	// Create messages in sequence
	contents := []string{"first", "second", "third"}
	for _, content := range contents {
		message := &domain.PlanCommentMessage{
			ThreadID: threadID,
			UserID:   userID,
			Content:  content,
		}
		err := s.Repo.Create(ctx, message)
		s.Require().NoError(err)
	}

	messages, err := s.Repo.FindByThreadID(ctx, threadID)
	s.Require().NoError(err)
	s.Require().Len(messages, 3)

	// Verify order
	for i := 1; i < len(messages); i++ {
		s.True(
			messages[i-1].CreatedAt.Before(messages[i].CreatedAt) || messages[i-1].CreatedAt.Equal(messages[i].CreatedAt),
			"Messages should be in chronological order",
		)
	}
}

func (s *PlanCommentMessageRepositorySuite) TestUpdate() {
	ctx := context.Background()

	threadID := s.createTestThread("msg-update")
	userID := s.createTestUser("msg-update")
	if threadID == "" {
		s.T().Skip("ThreadRepo not available, skipping test")
	}

	message := &domain.PlanCommentMessage{
		ThreadID: threadID,
		UserID:   userID,
		Content:  "original content",
	}
	err := s.Repo.Create(ctx, message)
	s.Require().NoError(err)

	// Update content
	message.Content = "updated content"
	err = s.Repo.Update(ctx, message)
	s.Require().NoError(err)

	// Verify update
	found, err := s.Repo.FindByID(ctx, message.ID)
	s.Require().NoError(err)
	s.Equal("updated content", found.Content)
}

func (s *PlanCommentMessageRepositorySuite) TestDelete() {
	ctx := context.Background()

	threadID := s.createTestThread("msg-delete")
	userID := s.createTestUser("msg-delete")
	if threadID == "" {
		s.T().Skip("ThreadRepo not available, skipping test")
	}

	message := &domain.PlanCommentMessage{
		ThreadID: threadID,
		UserID:   userID,
		Content:  "delete message",
	}
	err := s.Repo.Create(ctx, message)
	s.Require().NoError(err)

	err = s.Repo.Delete(ctx, message.ID)
	s.Require().NoError(err)

	found, err := s.Repo.FindByID(ctx, message.ID)
	s.NoError(err)
	s.Nil(found)
}

func (s *PlanCommentMessageRepositorySuite) TestDeleteByThreadID() {
	ctx := context.Background()

	threadID := s.createTestThread("msg-deletebythread")
	otherThreadID := s.createTestThread("msg-deletebythread-other")
	userID := s.createTestUser("msg-deletebythread")
	if threadID == "" || otherThreadID == "" {
		s.T().Skip("ThreadRepo not available, skipping test")
	}

	// Create messages for target thread
	for i := 0; i < 3; i++ {
		message := &domain.PlanCommentMessage{
			ThreadID: threadID,
			UserID:   userID,
			Content:  "message " + string(rune('a'+i)),
		}
		err := s.Repo.Create(ctx, message)
		s.Require().NoError(err)
	}

	// Create message for other thread
	otherMessage := &domain.PlanCommentMessage{
		ThreadID: otherThreadID,
		UserID:   userID,
		Content:  "other message",
	}
	err := s.Repo.Create(ctx, otherMessage)
	s.Require().NoError(err)

	// Delete messages by thread ID
	err = s.Repo.DeleteByThreadID(ctx, threadID)
	s.Require().NoError(err)

	// Verify target thread messages are deleted
	messages, err := s.Repo.FindByThreadID(ctx, threadID)
	s.Require().NoError(err)
	s.Len(messages, 0)

	// Verify other thread messages are not affected
	otherMessages, err := s.Repo.FindByThreadID(ctx, otherThreadID)
	s.Require().NoError(err)
	s.Len(otherMessages, 1)
}
