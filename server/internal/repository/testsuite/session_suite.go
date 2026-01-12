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

// createTestUser creates a user for FK constraint tests
func (s *SessionRepositorySuite) createTestUser(id string) {
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

// createTestProject creates a project for FK constraint tests
func (s *SessionRepositorySuite) createTestProject(id string) {
	if s.ProjectRepo == nil {
		return
	}
	ctx := context.Background()
	project := &domain.Project{
		ID:                     id,
		CanonicalGitRepository: "https://github.com/test/" + id,
	}
	_ = s.ProjectRepo.Create(ctx, project)
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

	userID := "user-123"
	s.createTestUser(userID)

	session := &domain.Session{
		ClaudeSessionID: "claude-session-2",
		UserID:          &userID,
	}

	err := s.Repo.Create(ctx, session)
	s.Require().NoError(err)
	s.Require().NotNil(session.UserID)
	s.Equal("user-123", *session.UserID)
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

	found, err := s.Repo.FindByID(ctx, "non-existing-id")
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

	// Find all with limit, default sort (updated_at)
	sessions, err := s.Repo.FindAll(ctx, 3, 0, "")
	s.Require().NoError(err)
	s.Len(sessions, 3)
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

	// Find all sorted by created_at
	sessions, err := s.Repo.FindAll(ctx, 5, 0, "created_at")
	s.Require().NoError(err)
	s.GreaterOrEqual(len(sessions), 5)

	// Verify order (newest first)
	for i := 0; i < len(sessions)-1; i++ {
		s.True(sessions[i].CreatedAt.After(sessions[i+1].CreatedAt) || sessions[i].CreatedAt.Equal(sessions[i+1].CreatedAt))
	}
}

func (s *SessionRepositorySuite) TestFindByProjectID() {
	ctx := context.Background()

	projectID := "test-project-id"
	s.createTestProject(projectID)
	s.createTestProject("other-project-id")

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
		ProjectID:       "other-project-id",
	}
	err := s.Repo.Create(ctx, otherSession)
	s.Require().NoError(err)

	// Find by project ID
	sessions, err := s.Repo.FindByProjectID(ctx, projectID, 10, 0, "")
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

	userID := "user-456"
	s.createTestUser(userID)

	session, err := s.Repo.FindOrCreateByClaudeSessionID(ctx, "new-claude-session-with-user", &userID)
	s.Require().NoError(err)
	s.Require().NotNil(session)
	s.Require().NotNil(session.UserID)
	s.Equal("user-456", *session.UserID)
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

	userID := "new-user-id"
	s.createTestUser(userID)

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
	s.Equal("new-user-id", *found.UserID)
}

func (s *SessionRepositorySuite) TestUpdateUserID() {
	ctx := context.Background()

	s.createTestUser("updated-user-id")

	session := &domain.Session{
		ClaudeSessionID: "session-update-userid",
	}
	err := s.Repo.Create(ctx, session)
	s.Require().NoError(err)

	err = s.Repo.UpdateUserID(ctx, session.ID, "updated-user-id")
	s.Require().NoError(err)

	found, err := s.Repo.FindByID(ctx, session.ID)
	s.Require().NoError(err)
	s.Require().NotNil(found.UserID)
	s.Equal("updated-user-id", *found.UserID)
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

	s.createTestProject("new-project-id")

	session := &domain.Session{
		ClaudeSessionID: "session-update-projectid",
	}
	err := s.Repo.Create(ctx, session)
	s.Require().NoError(err)

	err = s.Repo.UpdateProjectID(ctx, session.ID, "new-project-id")
	s.Require().NoError(err)

	found, err := s.Repo.FindByID(ctx, session.ID)
	s.Require().NoError(err)
	s.Equal("new-project-id", found.ProjectID)
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
