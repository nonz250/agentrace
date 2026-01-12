package testsuite

import (
	"context"
	"time"

	"github.com/satetsu888/agentrace/server/internal/domain"
	"github.com/satetsu888/agentrace/server/internal/repository"
	"github.com/stretchr/testify/suite"
)

// ProjectRepositorySuite tests ProjectRepository implementations
type ProjectRepositorySuite struct {
	suite.Suite
	Repo    repository.ProjectRepository
	Cleanup func()
}

func (s *ProjectRepositorySuite) SetupTest() {
	// Cleanup is optional - some implementations may need it
}

func (s *ProjectRepositorySuite) TearDownTest() {
	if s.Cleanup != nil {
		s.Cleanup()
	}
}

func (s *ProjectRepositorySuite) TestCreate() {
	ctx := context.Background()

	project := &domain.Project{
		CanonicalGitRepository: "https://github.com/example/repo1",
	}

	err := s.Repo.Create(ctx, project)
	s.Require().NoError(err)

	// ID should be auto-generated
	s.NotEmpty(project.ID)

	// CreatedAt should be set
	s.False(project.CreatedAt.IsZero())

	// Verify by finding
	found, err := s.Repo.FindByID(ctx, project.ID)
	s.Require().NoError(err)
	s.Require().NotNil(found)
	s.Equal(project.CanonicalGitRepository, found.CanonicalGitRepository)
}

func (s *ProjectRepositorySuite) TestCreate_WithID() {
	ctx := context.Background()

	project := &domain.Project{
		ID:                     "custom-id-123",
		CanonicalGitRepository: "https://github.com/example/repo2",
	}

	err := s.Repo.Create(ctx, project)
	s.Require().NoError(err)

	// ID should remain as specified
	s.Equal("custom-id-123", project.ID)

	// Verify by finding
	found, err := s.Repo.FindByID(ctx, project.ID)
	s.Require().NoError(err)
	s.Require().NotNil(found)
	s.Equal("custom-id-123", found.ID)
}

func (s *ProjectRepositorySuite) TestFindByID() {
	ctx := context.Background()

	project := &domain.Project{
		CanonicalGitRepository: "https://github.com/example/repo3",
	}
	err := s.Repo.Create(ctx, project)
	s.Require().NoError(err)

	// Find existing
	found, err := s.Repo.FindByID(ctx, project.ID)
	s.Require().NoError(err)
	s.Require().NotNil(found)
	s.Equal(project.ID, found.ID)
	s.Equal(project.CanonicalGitRepository, found.CanonicalGitRepository)
}

func (s *ProjectRepositorySuite) TestFindByID_NotFound() {
	ctx := context.Background()

	// Find non-existing
	found, err := s.Repo.FindByID(ctx, "non-existing-id")
	s.NoError(err)
	s.Nil(found)
}

func (s *ProjectRepositorySuite) TestFindByCanonicalGitRepository() {
	ctx := context.Background()

	project := &domain.Project{
		CanonicalGitRepository: "https://github.com/example/repo4",
	}
	err := s.Repo.Create(ctx, project)
	s.Require().NoError(err)

	// Find by canonical git repo
	found, err := s.Repo.FindByCanonicalGitRepository(ctx, "https://github.com/example/repo4")
	s.Require().NoError(err)
	s.Require().NotNil(found)
	s.Equal(project.ID, found.ID)
}

func (s *ProjectRepositorySuite) TestFindByCanonicalGitRepository_NotFound() {
	ctx := context.Background()

	found, err := s.Repo.FindByCanonicalGitRepository(ctx, "https://github.com/non-existing/repo")
	s.NoError(err)
	s.Nil(found)
}

func (s *ProjectRepositorySuite) TestFindOrCreateByCanonicalGitRepository_Create() {
	ctx := context.Background()

	// Should create new project
	project, err := s.Repo.FindOrCreateByCanonicalGitRepository(ctx, "https://github.com/example/new-repo")
	s.Require().NoError(err)
	s.Require().NotNil(project)
	s.NotEmpty(project.ID)
	s.Equal("https://github.com/example/new-repo", project.CanonicalGitRepository)
}

func (s *ProjectRepositorySuite) TestFindOrCreateByCanonicalGitRepository_Find() {
	ctx := context.Background()

	// Create first
	original := &domain.Project{
		CanonicalGitRepository: "https://github.com/example/existing-repo",
	}
	err := s.Repo.Create(ctx, original)
	s.Require().NoError(err)

	// FindOrCreate should return existing
	found, err := s.Repo.FindOrCreateByCanonicalGitRepository(ctx, "https://github.com/example/existing-repo")
	s.Require().NoError(err)
	s.Require().NotNil(found)
	s.Equal(original.ID, found.ID)
}

func (s *ProjectRepositorySuite) TestFindAll() {
	ctx := context.Background()

	// Create multiple projects
	for i := 0; i < 5; i++ {
		project := &domain.Project{
			CanonicalGitRepository: "https://github.com/example/findall-repo-" + string(rune('a'+i)),
		}
		time.Sleep(1 * time.Millisecond) // Ensure different CreatedAt
		err := s.Repo.Create(ctx, project)
		s.Require().NoError(err)
	}

	// Find all with limit
	projects, err := s.Repo.FindAll(ctx, 3, 0)
	s.Require().NoError(err)
	s.Len(projects, 3)

	// Verify order (newest first)
	for i := 0; i < len(projects)-1; i++ {
		s.True(projects[i].CreatedAt.After(projects[i+1].CreatedAt) || projects[i].CreatedAt.Equal(projects[i+1].CreatedAt))
	}
}

func (s *ProjectRepositorySuite) TestFindAll_WithOffset() {
	ctx := context.Background()

	// Create multiple projects
	for i := 0; i < 5; i++ {
		project := &domain.Project{
			CanonicalGitRepository: "https://github.com/example/offset-repo-" + string(rune('a'+i)),
		}
		time.Sleep(1 * time.Millisecond)
		err := s.Repo.Create(ctx, project)
		s.Require().NoError(err)
	}

	// Find with offset
	projects, err := s.Repo.FindAll(ctx, 10, 2)
	s.Require().NoError(err)
	s.GreaterOrEqual(len(projects), 3) // At least 3 from our 5 created
}

func (s *ProjectRepositorySuite) TestGetDefaultProject() {
	ctx := context.Background()

	defaultProject, err := s.Repo.GetDefaultProject(ctx)
	s.Require().NoError(err)
	s.Require().NotNil(defaultProject)
	s.Equal(domain.DefaultProjectID, defaultProject.ID)
	s.Empty(defaultProject.CanonicalGitRepository)
}
