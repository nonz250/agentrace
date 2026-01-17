package testsuite

import (
	"context"
	"time"

	"github.com/satetsu888/agentrace/server/internal/domain"
	"github.com/satetsu888/agentrace/server/internal/repository"
	"github.com/stretchr/testify/suite"
)

// SessionRepositorySuite tests SessionRepository implementations
type SessionRepositorySuite struct {
	suite.Suite
	Repo        repository.SessionRepository
	UserRepo    repository.UserRepository    // Optional: for FK constraint support
	ProjectRepo repository.ProjectRepository // Optional: for FK constraint support
	Cleanup     func()
}

// createTestUserWithID creates a user with a specific ID for FK constraint tests
// Returns the created user ID, or empty string if UserRepo is nil
func (s *SessionRepositorySuite) createTestUserWithID(id string) string {
	if s.UserRepo == nil {
		return ""
	}
	ctx := context.Background()
	user := &domain.User{
		ID:    id,
		Email: id + "@example.com",
	}
	_ = s.UserRepo.Create(ctx, user)
	return id
}

// createTestUser creates a user and returns the auto-generated ID
// Returns empty string if UserRepo is nil
func (s *SessionRepositorySuite) createTestUser(emailPrefix string) string {
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

// createTestProjectWithID creates a project with a specific ID for FK constraint tests
// Returns the created project ID, or empty string if ProjectRepo is nil
func (s *SessionRepositorySuite) createTestProjectWithID(id string) string {
	if s.ProjectRepo == nil {
		return ""
	}
	ctx := context.Background()
	project := &domain.Project{
		ID:                     id,
		CanonicalGitRepository: "https://github.com/test/" + id,
	}
	_ = s.ProjectRepo.Create(ctx, project)
	return id
}

// createTestProject creates a project and returns the auto-generated ID
// Returns empty string if ProjectRepo is nil
func (s *SessionRepositorySuite) createTestProject(gitRepoSuffix string) string {
	if s.ProjectRepo == nil {
		return ""
	}
	ctx := context.Background()
	project := &domain.Project{
		CanonicalGitRepository: "https://github.com/test/" + gitRepoSuffix,
	}
	err := s.ProjectRepo.Create(ctx, project)
	if err != nil {
		return ""
	}
	return project.ID
}

func (s *SessionRepositorySuite) TearDownTest() {
	if s.Cleanup != nil {
		s.Cleanup()
	}
}

func (s *SessionRepositorySuite) TestCreate() {
	ctx := context.Background()

	session := &domain.Session{
		ClaudeSessionID: "claude-session-1",
		ProjectPath:     "/path/to/project",
	}

	err := s.Repo.Create(ctx, session)
	s.Require().NoError(err)

	// ID should be auto-generated
	s.NotEmpty(session.ID)

	// Timestamps should be set
	s.False(session.CreatedAt.IsZero())
	s.False(session.StartedAt.IsZero())
	s.False(session.UpdatedAt.IsZero())

	// Default ProjectID should be set
	s.Equal(domain.DefaultProjectID, session.ProjectID)
}

func (s *SessionRepositorySuite) TestCreate_WithUserID() {
	ctx := context.Background()

	userID := s.createTestUser("user-for-create")
	if userID == "" {
		s.T().Skip("UserRepo not available, skipping test")
	}

	session := &domain.Session{
		ClaudeSessionID: "claude-session-2",
		UserID:          &userID,
	}

	err := s.Repo.Create(ctx, session)
	s.Require().NoError(err)
	s.Require().NotNil(session.UserID)
	s.Equal(userID, *session.UserID)
}

func (s *SessionRepositorySuite) TestFindByID() {
	ctx := context.Background()

	session := &domain.Session{
		ClaudeSessionID: "claude-session-3",
	}
	err := s.Repo.Create(ctx, session)
	s.Require().NoError(err)

	found, err := s.Repo.FindByID(ctx, session.ID)
	s.Require().NoError(err)
	s.Require().NotNil(found)
	s.Equal(session.ID, found.ID)
	s.Equal(session.ClaudeSessionID, found.ClaudeSessionID)
}

func (s *SessionRepositorySuite) TestFindByID_NotFound() {
	ctx := context.Background()

	// Create a session to get a valid ID format, then modify it to create a non-existent ID
	session := &domain.Session{
		ClaudeSessionID: "session-for-notfound-id",
	}
	err := s.Repo.Create(ctx, session)
	s.Require().NoError(err)

	// Modify the last character to create a valid but non-existent ID
	nonExistentID := session.ID[:len(session.ID)-1] + "0"
	if session.ID[len(session.ID)-1] == '0' {
		nonExistentID = session.ID[:len(session.ID)-1] + "1"
	}

	found, err := s.Repo.FindByID(ctx, nonExistentID)
	s.NoError(err)
	s.Nil(found)
}

func (s *SessionRepositorySuite) TestFindByClaudeSessionID() {
	ctx := context.Background()

	session := &domain.Session{
		ClaudeSessionID: "claude-session-unique",
	}
	err := s.Repo.Create(ctx, session)
	s.Require().NoError(err)

	found, err := s.Repo.FindByClaudeSessionID(ctx, "claude-session-unique")
	s.Require().NoError(err)
	s.Require().NotNil(found)
	s.Equal(session.ID, found.ID)
}

func (s *SessionRepositorySuite) TestFindByClaudeSessionID_NotFound() {
	ctx := context.Background()

	found, err := s.Repo.FindByClaudeSessionID(ctx, "non-existing-claude-session")
	s.NoError(err)
	s.Nil(found)
}

func (s *SessionRepositorySuite) TestFindAll() {
	ctx := context.Background()

	// Create multiple sessions
	for i := 0; i < 5; i++ {
		session := &domain.Session{
			ClaudeSessionID: "findall-session-" + string(rune('a'+i)),
		}
		time.Sleep(1 * time.Millisecond)
		err := s.Repo.Create(ctx, session)
		s.Require().NoError(err)
	}

	// Find all with limit, default sort (updated_at), cursor-based pagination
	sessions, nextCursor, err := s.Repo.FindAll(ctx, 3, "", "")
	s.Require().NoError(err)
	s.Len(sessions, 3)
	s.NotEmpty(nextCursor) // More items available
}

func (s *SessionRepositorySuite) TestFindAll_SortByCreatedAt() {
	ctx := context.Background()

	// Create multiple sessions
	for i := 0; i < 5; i++ {
		session := &domain.Session{
			ClaudeSessionID: "sort-created-session-" + string(rune('a'+i)),
		}
		time.Sleep(1 * time.Millisecond)
		err := s.Repo.Create(ctx, session)
		s.Require().NoError(err)
	}

	// Find all sorted by created_at (cursor-based pagination)
	sessions, _, err := s.Repo.FindAll(ctx, 5, "", "created_at")
	s.Require().NoError(err)
	s.GreaterOrEqual(len(sessions), 5)

	// Verify order (newest first)
	for i := 0; i < len(sessions)-1; i++ {
		s.True(sessions[i].CreatedAt.After(sessions[i+1].CreatedAt) || sessions[i].CreatedAt.Equal(sessions[i+1].CreatedAt))
	}
}

func (s *SessionRepositorySuite) TestFindByProjectID() {
	ctx := context.Background()

	projectID := s.createTestProject("test-project")
	otherProjectID := s.createTestProject("other-project")
	if projectID == "" || otherProjectID == "" {
		s.T().Skip("ProjectRepo not available, skipping test")
	}

	// Create sessions for different projects
	for i := 0; i < 3; i++ {
		session := &domain.Session{
			ClaudeSessionID: "project-session-" + string(rune('a'+i)),
			ProjectID:       projectID,
		}
		time.Sleep(1 * time.Millisecond)
		err := s.Repo.Create(ctx, session)
		s.Require().NoError(err)
	}

	// Create session for different project
	otherSession := &domain.Session{
		ClaudeSessionID: "other-project-session",
		ProjectID:       otherProjectID,
	}
	err := s.Repo.Create(ctx, otherSession)
	s.Require().NoError(err)

	// Find by project ID (cursor-based pagination)
	sessions, _, err := s.Repo.FindByProjectID(ctx, projectID, 10, "", "")
	s.Require().NoError(err)
	s.Len(sessions, 3)

	for _, sess := range sessions {
		s.Equal(projectID, sess.ProjectID)
	}
}

func (s *SessionRepositorySuite) TestFindOrCreateByClaudeSessionID_Create() {
	ctx := context.Background()

	session, err := s.Repo.FindOrCreateByClaudeSessionID(ctx, "new-claude-session", nil)
	s.Require().NoError(err)
	s.Require().NotNil(session)
	s.NotEmpty(session.ID)
	s.Equal("new-claude-session", session.ClaudeSessionID)
	s.Nil(session.UserID)
}

func (s *SessionRepositorySuite) TestFindOrCreateByClaudeSessionID_CreateWithUserID() {
	ctx := context.Background()

	userID := s.createTestUser("user-for-findorcreate")
	if userID == "" {
		s.T().Skip("UserRepo not available, skipping test")
	}

	session, err := s.Repo.FindOrCreateByClaudeSessionID(ctx, "new-claude-session-with-user", &userID)
	s.Require().NoError(err)
	s.Require().NotNil(session)
	s.Require().NotNil(session.UserID)
	s.Equal(userID, *session.UserID)
}

func (s *SessionRepositorySuite) TestFindOrCreateByClaudeSessionID_Find() {
	ctx := context.Background()

	// Create first
	original := &domain.Session{
		ClaudeSessionID: "existing-claude-session",
	}
	err := s.Repo.Create(ctx, original)
	s.Require().NoError(err)

	// FindOrCreate should return existing
	found, err := s.Repo.FindOrCreateByClaudeSessionID(ctx, "existing-claude-session", nil)
	s.Require().NoError(err)
	s.Require().NotNil(found)
	s.Equal(original.ID, found.ID)
}

func (s *SessionRepositorySuite) TestFindOrCreateByClaudeSessionID_FindAndUpdateUserID() {
	ctx := context.Background()

	userID := s.createTestUser("user-for-update")
	if userID == "" {
		s.T().Skip("UserRepo not available, skipping test")
	}

	// Create without UserID
	original := &domain.Session{
		ClaudeSessionID: "session-to-update-user",
	}
	err := s.Repo.Create(ctx, original)
	s.Require().NoError(err)
	s.Nil(original.UserID)

	// FindOrCreate with UserID should update
	found, err := s.Repo.FindOrCreateByClaudeSessionID(ctx, "session-to-update-user", &userID)
	s.Require().NoError(err)
	s.Require().NotNil(found)
	s.Equal(original.ID, found.ID)
	s.Require().NotNil(found.UserID)
	s.Equal(userID, *found.UserID)
}

func (s *SessionRepositorySuite) TestUpdateUserID() {
	ctx := context.Background()

	userID := s.createTestUser("user-for-updateuserid")
	if userID == "" {
		s.T().Skip("UserRepo not available, skipping test")
	}

	session := &domain.Session{
		ClaudeSessionID: "session-update-userid",
	}
	err := s.Repo.Create(ctx, session)
	s.Require().NoError(err)

	err = s.Repo.UpdateUserID(ctx, session.ID, userID)
	s.Require().NoError(err)

	found, err := s.Repo.FindByID(ctx, session.ID)
	s.Require().NoError(err)
	s.Require().NotNil(found.UserID)
	s.Equal(userID, *found.UserID)
}

func (s *SessionRepositorySuite) TestUpdateProjectPath() {
	ctx := context.Background()

	session := &domain.Session{
		ClaudeSessionID: "session-update-path",
	}
	err := s.Repo.Create(ctx, session)
	s.Require().NoError(err)

	err = s.Repo.UpdateProjectPath(ctx, session.ID, "/new/project/path")
	s.Require().NoError(err)

	found, err := s.Repo.FindByID(ctx, session.ID)
	s.Require().NoError(err)
	s.Equal("/new/project/path", found.ProjectPath)
}

func (s *SessionRepositorySuite) TestUpdateProjectID() {
	ctx := context.Background()

	projectID := s.createTestProject("project-for-updateprojectid")
	if projectID == "" {
		s.T().Skip("ProjectRepo not available, skipping test")
	}

	session := &domain.Session{
		ClaudeSessionID: "session-update-projectid",
	}
	err := s.Repo.Create(ctx, session)
	s.Require().NoError(err)

	err = s.Repo.UpdateProjectID(ctx, session.ID, projectID)
	s.Require().NoError(err)

	found, err := s.Repo.FindByID(ctx, session.ID)
	s.Require().NoError(err)
	s.Equal(projectID, found.ProjectID)
}

func (s *SessionRepositorySuite) TestUpdateGitBranch() {
	ctx := context.Background()

	session := &domain.Session{
		ClaudeSessionID: "session-update-branch",
	}
	err := s.Repo.Create(ctx, session)
	s.Require().NoError(err)

	err = s.Repo.UpdateGitBranch(ctx, session.ID, "feature/new-branch")
	s.Require().NoError(err)

	found, err := s.Repo.FindByID(ctx, session.ID)
	s.Require().NoError(err)
	s.Equal("feature/new-branch", found.GitBranch)
}

func (s *SessionRepositorySuite) TestUpdateTitle() {
	ctx := context.Background()

	session := &domain.Session{
		ClaudeSessionID: "session-update-title",
	}
	err := s.Repo.Create(ctx, session)
	s.Require().NoError(err)

	err = s.Repo.UpdateTitle(ctx, session.ID, "New Session Title")
	s.Require().NoError(err)

	found, err := s.Repo.FindByID(ctx, session.ID)
	s.Require().NoError(err)
	s.Require().NotNil(found.Title)
	s.Equal("New Session Title", *found.Title)
}

func (s *SessionRepositorySuite) TestUpdateUpdatedAt() {
	ctx := context.Background()

	session := &domain.Session{
		ClaudeSessionID: "session-update-updatedat",
	}
	err := s.Repo.Create(ctx, session)
	s.Require().NoError(err)

	newTime := time.Now().Add(1 * time.Hour)
	err = s.Repo.UpdateUpdatedAt(ctx, session.ID, newTime)
	s.Require().NoError(err)

	found, err := s.Repo.FindByID(ctx, session.ID)
	s.Require().NoError(err)
	s.WithinDuration(newTime, found.UpdatedAt, time.Second)
}

// Subagent (Task tool) related tests

func (s *SessionRepositorySuite) TestCreate_WithSubagentFields() {
	ctx := context.Background()

	// Create parent session first
	parentSession := &domain.Session{
		ClaudeSessionID: "parent-session-for-subagent",
	}
	err := s.Repo.Create(ctx, parentSession)
	s.Require().NoError(err)

	// Create subagent session
	agentID := "a5d7f46"
	subagentSession := &domain.Session{
		ClaudeSessionID: "subagent-session-1",
		ParentSessionID: &parentSession.ID,
		AgentID:         &agentID,
		IsSidechain:     true,
	}
	err = s.Repo.Create(ctx, subagentSession)
	s.Require().NoError(err)

	// Verify fields are persisted
	found, err := s.Repo.FindByID(ctx, subagentSession.ID)
	s.Require().NoError(err)
	s.Require().NotNil(found)
	s.Require().NotNil(found.ParentSessionID)
	s.Equal(parentSession.ID, *found.ParentSessionID)
	s.Require().NotNil(found.AgentID)
	s.Equal("a5d7f46", *found.AgentID)
	s.True(found.IsSidechain)
}

func (s *SessionRepositorySuite) TestFindSubagentsByParentID() {
	ctx := context.Background()

	// Create parent session
	parentSession := &domain.Session{
		ClaudeSessionID: "parent-session-subagents",
	}
	err := s.Repo.Create(ctx, parentSession)
	s.Require().NoError(err)

	// Create multiple subagent sessions
	for i := 0; i < 3; i++ {
		agentID := "agent-" + string(rune('a'+i))
		subagent := &domain.Session{
			ClaudeSessionID: "subagent-" + string(rune('a'+i)),
			ParentSessionID: &parentSession.ID,
			AgentID:         &agentID,
			IsSidechain:     true,
		}
		time.Sleep(1 * time.Millisecond)
		err := s.Repo.Create(ctx, subagent)
		s.Require().NoError(err)
	}

	// Find subagents
	subagents, err := s.Repo.FindSubagentsByParentID(ctx, parentSession.ID)
	s.Require().NoError(err)
	s.Len(subagents, 3)

	// Verify all are subagents of the parent
	for _, sub := range subagents {
		s.Require().NotNil(sub.ParentSessionID)
		s.Equal(parentSession.ID, *sub.ParentSessionID)
		s.True(sub.IsSidechain)
	}
}

func (s *SessionRepositorySuite) TestFindSubagentsByParentID_Empty() {
	ctx := context.Background()

	// Create session with no subagents
	session := &domain.Session{
		ClaudeSessionID: "session-no-subagents",
	}
	err := s.Repo.Create(ctx, session)
	s.Require().NoError(err)

	// Find subagents - should be empty
	subagents, err := s.Repo.FindSubagentsByParentID(ctx, session.ID)
	s.Require().NoError(err)
	s.Empty(subagents)
}

func (s *SessionRepositorySuite) TestFindSubagentsByParentID_NonExistentParent() {
	ctx := context.Background()

	// Create a session to get a valid ID format, then use a different valid-format ID
	session := &domain.Session{
		ClaudeSessionID: "session-for-id-format",
	}
	err := s.Repo.Create(ctx, session)
	s.Require().NoError(err)

	// Find subagents for non-existent parent (use the session ID prefix with different suffix)
	// This ensures the ID format is valid for databases that use UUIDs
	nonExistentID := session.ID[:len(session.ID)-1] + "0"
	if session.ID[len(session.ID)-1] == '0' {
		nonExistentID = session.ID[:len(session.ID)-1] + "1"
	}

	subagents, err := s.Repo.FindSubagentsByParentID(ctx, nonExistentID)
	s.Require().NoError(err)
	s.Empty(subagents)
}

func (s *SessionRepositorySuite) TestFindAll_ExcludesSubagents() {
	ctx := context.Background()

	// Create regular sessions
	for i := 0; i < 3; i++ {
		session := &domain.Session{
			ClaudeSessionID: "regular-session-exclude-" + string(rune('a'+i)),
		}
		time.Sleep(1 * time.Millisecond)
		err := s.Repo.Create(ctx, session)
		s.Require().NoError(err)
	}

	// Create parent session
	parentSession := &domain.Session{
		ClaudeSessionID: "parent-session-exclude",
	}
	err := s.Repo.Create(ctx, parentSession)
	s.Require().NoError(err)

	// Create subagent sessions (should be excluded from FindAll)
	for i := 0; i < 2; i++ {
		agentID := "exclude-agent-" + string(rune('a'+i))
		subagent := &domain.Session{
			ClaudeSessionID: "subagent-exclude-" + string(rune('a'+i)),
			ParentSessionID: &parentSession.ID,
			AgentID:         &agentID,
			IsSidechain:     true,
		}
		err := s.Repo.Create(ctx, subagent)
		s.Require().NoError(err)
	}

	// FindAll should return only regular sessions (not subagents)
	sessions, _, err := s.Repo.FindAll(ctx, 100, "", "")
	s.Require().NoError(err)

	// Verify no subagents in the result
	for _, sess := range sessions {
		s.False(sess.IsSidechain, "FindAll should not return subagent sessions")
	}
}

func (s *SessionRepositorySuite) TestFindByProjectID_ExcludesSubagents() {
	ctx := context.Background()

	// Skip if ProjectRepo is not available (can't create unique project)
	if s.ProjectRepo == nil {
		s.T().Skip("ProjectRepo not available, skipping test")
	}

	// Create a unique project for this test
	uniqueProject := &domain.Project{
		CanonicalGitRepository: "https://github.com/test/subagent-exclude-test-" + time.Now().Format("20060102150405"),
	}
	err := s.ProjectRepo.Create(ctx, uniqueProject)
	s.Require().NoError(err)
	projectID := uniqueProject.ID

	// Create regular sessions for the project
	for i := 0; i < 2; i++ {
		session := &domain.Session{
			ClaudeSessionID: "project-regular-" + string(rune('a'+i)),
			ProjectID:       projectID,
		}
		time.Sleep(1 * time.Millisecond)
		err := s.Repo.Create(ctx, session)
		s.Require().NoError(err)
	}

	// Create parent session
	parentSession := &domain.Session{
		ClaudeSessionID: "project-parent-exclude",
		ProjectID:       projectID,
	}
	err = s.Repo.Create(ctx, parentSession)
	s.Require().NoError(err)

	// Create subagent session for the same project
	agentID := "project-agent-1"
	subagent := &domain.Session{
		ClaudeSessionID: "project-subagent-exclude",
		ProjectID:       projectID,
		ParentSessionID: &parentSession.ID,
		AgentID:         &agentID,
		IsSidechain:     true,
	}
	err = s.Repo.Create(ctx, subagent)
	s.Require().NoError(err)

	// FindByProjectID should exclude subagents
	sessions, _, err := s.Repo.FindByProjectID(ctx, projectID, 100, "", "")
	s.Require().NoError(err)

	// Verify no subagents in the result
	for _, sess := range sessions {
		s.False(sess.IsSidechain, "FindByProjectID should not return subagent sessions")
	}

	// Should have 3 regular sessions (2 regular + 1 parent)
	s.Len(sessions, 3)
}
