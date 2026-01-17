package testsuite

import (
	"context"
	"time"

	"github.com/google/uuid"
	"github.com/satetsu888/agentrace/server/internal/domain"
	"github.com/satetsu888/agentrace/server/internal/repository"
	"github.com/stretchr/testify/suite"
)

// PlanDocumentRepositorySuite tests PlanDocumentRepository implementations
type PlanDocumentRepositorySuite struct {
	suite.Suite
	Repo        repository.PlanDocumentRepository
	ProjectRepo repository.ProjectRepository // Optional: for FK constraint support
	Cleanup     func()
}

// createTestProject creates a project for FK constraint tests and returns the auto-generated ID
func (s *PlanDocumentRepositorySuite) createTestProject(gitRepoSuffix string) string {
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

func (s *PlanDocumentRepositorySuite) TearDownTest() {
	if s.Cleanup != nil {
		s.Cleanup()
	}
}

func (s *PlanDocumentRepositorySuite) TestCreate() {
	ctx := context.Background()

	projectID := s.createTestProject("plandoc-create")
	if projectID == "" {
		s.T().Skip("ProjectRepo not available, skipping test")
	}

	doc := &domain.PlanDocument{
		ProjectID:   projectID,
		Description: "Test Plan",
		Body:        "# Plan\n\nThis is a test plan.",
		Status:      domain.PlanDocumentStatusPlanning,
	}

	err := s.Repo.Create(ctx, doc)
	s.Require().NoError(err)

	// ID should be auto-generated
	s.NotEmpty(doc.ID)

	// Timestamps should be set
	s.False(doc.CreatedAt.IsZero())
	s.False(doc.UpdatedAt.IsZero())
}

func (s *PlanDocumentRepositorySuite) TestFindByID() {
	ctx := context.Background()

	projectID := s.createTestProject("plandoc-findid")
	if projectID == "" {
		s.T().Skip("ProjectRepo not available, skipping test")
	}

	doc := &domain.PlanDocument{
		ProjectID:   projectID,
		Description: "Find By ID Plan",
		Body:        "Body content",
		Status:      domain.PlanDocumentStatusPlanning,
	}
	err := s.Repo.Create(ctx, doc)
	s.Require().NoError(err)

	found, err := s.Repo.FindByID(ctx, doc.ID)
	s.Require().NoError(err)
	s.Require().NotNil(found)
	s.Equal(doc.ID, found.ID)
	s.Equal(doc.Description, found.Description)
	s.Equal(doc.Body, found.Body)
}

func (s *PlanDocumentRepositorySuite) TestFindByID_NotFound() {
	ctx := context.Background()

	nonExistingID := uuid.New().String()
	found, err := s.Repo.FindByID(ctx, nonExistingID)
	s.NoError(err)
	s.Nil(found)
}

func (s *PlanDocumentRepositorySuite) TestFind_ByProjectID() {
	ctx := context.Background()

	projectID := s.createTestProject("plandoc-findproj")
	otherProjectID := s.createTestProject("plandoc-other")
	if projectID == "" || otherProjectID == "" {
		s.T().Skip("ProjectRepo not available, skipping test")
	}

	// Create docs for project
	for i := 0; i < 3; i++ {
		doc := &domain.PlanDocument{
			ProjectID:   projectID,
			Description: "Plan " + string(rune('A'+i)),
			Body:        "Body " + string(rune('a'+i)),
			Status:      domain.PlanDocumentStatusPlanning,
		}
		time.Sleep(1 * time.Millisecond)
		err := s.Repo.Create(ctx, doc)
		s.Require().NoError(err)
	}

	// Create doc for different project
	otherDoc := &domain.PlanDocument{
		ProjectID:   otherProjectID,
		Description: "Other Plan",
		Body:        "Other body",
		Status:      domain.PlanDocumentStatusPlanning,
	}
	err := s.Repo.Create(ctx, otherDoc)
	s.Require().NoError(err)

	docs, _, err := s.Repo.Find(ctx, domain.PlanDocumentQuery{ProjectID: projectID})
	s.Require().NoError(err)
	s.Len(docs, 3)

	for _, d := range docs {
		s.Equal(projectID, d.ProjectID)
	}
}

func (s *PlanDocumentRepositorySuite) TestFind_ByStatuses() {
	ctx := context.Background()

	projectID := s.createTestProject("plandoc-status")
	if projectID == "" {
		s.T().Skip("ProjectRepo not available, skipping test")
	}

	// Create docs with different statuses
	statuses := []domain.PlanDocumentStatus{
		domain.PlanDocumentStatusScratch,
		domain.PlanDocumentStatusPlanning,
		domain.PlanDocumentStatusImplementation,
		domain.PlanDocumentStatusComplete,
	}

	for i, status := range statuses {
		doc := &domain.PlanDocument{
			ProjectID:   projectID,
			Description: "Status Plan " + string(rune('A'+i)),
			Body:        "Body",
			Status:      status,
		}
		err := s.Repo.Create(ctx, doc)
		s.Require().NoError(err)
	}

	// Find only planning and implementation
	docs, _, err := s.Repo.Find(ctx, domain.PlanDocumentQuery{
		ProjectID: projectID,
		Statuses: []domain.PlanDocumentStatus{
			domain.PlanDocumentStatusPlanning,
			domain.PlanDocumentStatusImplementation,
		},
	})
	s.Require().NoError(err)
	s.Len(docs, 2)

	for _, d := range docs {
		s.True(d.Status == domain.PlanDocumentStatusPlanning || d.Status == domain.PlanDocumentStatusImplementation)
	}
}

func (s *PlanDocumentRepositorySuite) TestFind_ByDescriptionContains() {
	ctx := context.Background()

	projectID := s.createTestProject("plandoc-desc")
	if projectID == "" {
		s.T().Skip("ProjectRepo not available, skipping test")
	}

	// Create docs with different descriptions
	descriptions := []string{"Authentication feature", "Database migration", "API endpoint"}
	for i, desc := range descriptions {
		doc := &domain.PlanDocument{
			ProjectID:   projectID,
			Description: desc,
			Body:        "Body " + string(rune('a'+i)),
			Status:      domain.PlanDocumentStatusPlanning,
		}
		err := s.Repo.Create(ctx, doc)
		s.Require().NoError(err)
	}

	// Find by description containing "feature" (case-insensitive)
	docs, _, err := s.Repo.Find(ctx, domain.PlanDocumentQuery{
		ProjectID:           projectID,
		DescriptionContains: "feature",
	})
	s.Require().NoError(err)
	s.Len(docs, 1)
	s.Contains(docs[0].Description, "feature")
}

func (s *PlanDocumentRepositorySuite) TestFind_ByPlanDocumentIDs() {
	ctx := context.Background()

	projectID := s.createTestProject("plandoc-ids")
	if projectID == "" {
		s.T().Skip("ProjectRepo not available, skipping test")
	}

	var ids []string
	for i := 0; i < 5; i++ {
		doc := &domain.PlanDocument{
			ProjectID:   projectID,
			Description: "Plan " + string(rune('A'+i)),
			Body:        "Body",
			Status:      domain.PlanDocumentStatusPlanning,
		}
		err := s.Repo.Create(ctx, doc)
		s.Require().NoError(err)
		ids = append(ids, doc.ID)
	}

	// Find specific IDs
	targetIDs := []string{ids[0], ids[2], ids[4]}
	docs, _, err := s.Repo.Find(ctx, domain.PlanDocumentQuery{
		PlanDocumentIDs: targetIDs,
	})
	s.Require().NoError(err)
	s.Len(docs, 3)

	foundIDs := make(map[string]bool)
	for _, d := range docs {
		foundIDs[d.ID] = true
	}
	for _, id := range targetIDs {
		s.True(foundIDs[id])
	}
}

func (s *PlanDocumentRepositorySuite) TestFind_WithCursor() {
	ctx := context.Background()

	projectID := s.createTestProject("plandoc-pagination")
	if projectID == "" {
		s.T().Skip("ProjectRepo not available, skipping test")
	}

	for i := 0; i < 10; i++ {
		doc := &domain.PlanDocument{
			ProjectID:   projectID,
			Description: "Plan " + string(rune('A'+i)),
			Body:        "Body",
			Status:      domain.PlanDocumentStatusPlanning,
		}
		time.Sleep(1 * time.Millisecond)
		err := s.Repo.Create(ctx, doc)
		s.Require().NoError(err)
	}

	// First page (cursor-based pagination)
	docs, nextCursor, err := s.Repo.Find(ctx, domain.PlanDocumentQuery{
		ProjectID: projectID,
		Limit:     3,
	})
	s.Require().NoError(err)
	s.Len(docs, 3)
	s.NotEmpty(nextCursor) // More items available

	// Second page using cursor
	docs2, _, err := s.Repo.Find(ctx, domain.PlanDocumentQuery{
		ProjectID: projectID,
		Limit:     3,
		Cursor:    nextCursor,
	})
	s.Require().NoError(err)
	s.Len(docs2, 3)

	// No overlap
	for _, d1 := range docs {
		for _, d2 := range docs2 {
			s.NotEqual(d1.ID, d2.ID)
		}
	}
}

func (s *PlanDocumentRepositorySuite) TestFind_SortByCreatedAt() {
	ctx := context.Background()

	projectID := s.createTestProject("plandoc-sort")
	if projectID == "" {
		s.T().Skip("ProjectRepo not available, skipping test")
	}

	for i := 0; i < 5; i++ {
		doc := &domain.PlanDocument{
			ProjectID:   projectID,
			Description: "Plan " + string(rune('A'+i)),
			Body:        "Body",
			Status:      domain.PlanDocumentStatusPlanning,
		}
		time.Sleep(2 * time.Millisecond)
		err := s.Repo.Create(ctx, doc)
		s.Require().NoError(err)
	}

	docs, _, err := s.Repo.Find(ctx, domain.PlanDocumentQuery{
		ProjectID: projectID,
		SortBy:    "created_at",
	})
	s.Require().NoError(err)
	s.GreaterOrEqual(len(docs), 5)

	// Verify order (newest first)
	for i := 0; i < len(docs)-1; i++ {
		s.True(docs[i].CreatedAt.After(docs[i+1].CreatedAt) || docs[i].CreatedAt.Equal(docs[i+1].CreatedAt))
	}
}

func (s *PlanDocumentRepositorySuite) TestUpdate() {
	ctx := context.Background()

	projectID := s.createTestProject("plandoc-update")
	if projectID == "" {
		s.T().Skip("ProjectRepo not available, skipping test")
	}

	doc := &domain.PlanDocument{
		ProjectID:   projectID,
		Description: "Original Description",
		Body:        "Original Body",
		Status:      domain.PlanDocumentStatusPlanning,
	}
	err := s.Repo.Create(ctx, doc)
	s.Require().NoError(err)

	// Update
	doc.Description = "Updated Description"
	doc.Body = "Updated Body"
	err = s.Repo.Update(ctx, doc)
	s.Require().NoError(err)

	// Verify
	found, err := s.Repo.FindByID(ctx, doc.ID)
	s.Require().NoError(err)
	s.Equal("Updated Description", found.Description)
	s.Equal("Updated Body", found.Body)
}

func (s *PlanDocumentRepositorySuite) TestDelete() {
	ctx := context.Background()

	projectID := s.createTestProject("plandoc-delete")
	if projectID == "" {
		s.T().Skip("ProjectRepo not available, skipping test")
	}

	doc := &domain.PlanDocument{
		ProjectID:   projectID,
		Description: "Delete Me",
		Body:        "Body",
		Status:      domain.PlanDocumentStatusPlanning,
	}
	err := s.Repo.Create(ctx, doc)
	s.Require().NoError(err)

	err = s.Repo.Delete(ctx, doc.ID)
	s.Require().NoError(err)

	// Verify deleted
	found, err := s.Repo.FindByID(ctx, doc.ID)
	s.NoError(err)
	s.Nil(found)
}

func (s *PlanDocumentRepositorySuite) TestSetStatus() {
	ctx := context.Background()

	projectID := s.createTestProject("plandoc-setstatus")
	if projectID == "" {
		s.T().Skip("ProjectRepo not available, skipping test")
	}

	doc := &domain.PlanDocument{
		ProjectID:   projectID,
		Description: "Status Change Plan",
		Body:        "Body",
		Status:      domain.PlanDocumentStatusPlanning,
	}
	err := s.Repo.Create(ctx, doc)
	s.Require().NoError(err)

	// Change status
	err = s.Repo.SetStatus(ctx, doc.ID, domain.PlanDocumentStatusImplementation)
	s.Require().NoError(err)

	// Verify
	found, err := s.Repo.FindByID(ctx, doc.ID)
	s.Require().NoError(err)
	s.Equal(domain.PlanDocumentStatusImplementation, found.Status)
}
