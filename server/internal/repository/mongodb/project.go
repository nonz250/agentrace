package mongodb

import (
	"context"
	"time"

	"github.com/google/uuid"
	"github.com/satetsu888/agentrace/server/internal/domain"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

type ProjectRepository struct {
	collection *mongo.Collection
}

func NewProjectRepository(db *DB) *ProjectRepository {
	return &ProjectRepository{
		collection: db.Collection("projects"),
	}
}

type projectDocument struct {
	ID                     string    `bson:"_id"`
	CanonicalGitRepository string    `bson:"canonical_git_repository"`
	CreatedAt              time.Time `bson:"created_at"`
}

func (r *ProjectRepository) Create(ctx context.Context, project *domain.Project) error {
	if project.ID == "" {
		project.ID = uuid.New().String()
	}
	if project.CreatedAt.IsZero() {
		project.CreatedAt = time.Now()
	}

	doc := projectDocument{
		ID:                     project.ID,
		CanonicalGitRepository: project.CanonicalGitRepository,
		CreatedAt:              project.CreatedAt,
	}

	_, err := r.collection.InsertOne(ctx, doc)
	return err
}

func (r *ProjectRepository) FindByID(ctx context.Context, id string) (*domain.Project, error) {
	var doc projectDocument
	err := r.collection.FindOne(ctx, bson.M{"_id": id}).Decode(&doc)
	if err == mongo.ErrNoDocuments {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}

	return docToProject(&doc), nil
}

func (r *ProjectRepository) FindByCanonicalGitRepository(ctx context.Context, canonicalGitRepo string) (*domain.Project, error) {
	var doc projectDocument
	err := r.collection.FindOne(ctx, bson.M{"canonical_git_repository": canonicalGitRepo}).Decode(&doc)
	if err == mongo.ErrNoDocuments {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}

	return docToProject(&doc), nil
}

func (r *ProjectRepository) FindOrCreateByCanonicalGitRepository(ctx context.Context, canonicalGitRepo string) (*domain.Project, error) {
	// First try to find existing project
	project, err := r.FindByCanonicalGitRepository(ctx, canonicalGitRepo)
	if err != nil {
		return nil, err
	}

	if project != nil {
		return project, nil
	}

	// Create new project
	newProject := &domain.Project{
		ID:                     uuid.New().String(),
		CanonicalGitRepository: canonicalGitRepo,
		CreatedAt:              time.Now(),
	}

	if err := r.Create(ctx, newProject); err != nil {
		// Handle race condition - another process may have created it
		existingProject, findErr := r.FindByCanonicalGitRepository(ctx, canonicalGitRepo)
		if findErr != nil {
			return nil, err // Return original error
		}
		if existingProject != nil {
			return existingProject, nil
		}
		return nil, err
	}

	return newProject, nil
}

func (r *ProjectRepository) FindAll(ctx context.Context, limit int, offset int) ([]*domain.Project, error) {
	opts := options.Find().SetSort(bson.D{{Key: "created_at", Value: -1}})
	if limit > 0 {
		opts.SetLimit(int64(limit))
		opts.SetSkip(int64(offset))
	}

	cursor, err := r.collection.Find(ctx, bson.M{}, opts)
	if err != nil {
		return nil, err
	}
	defer cursor.Close(ctx)

	var projects []*domain.Project
	for cursor.Next(ctx) {
		var doc projectDocument
		if err := cursor.Decode(&doc); err != nil {
			return nil, err
		}
		projects = append(projects, docToProject(&doc))
	}

	return projects, cursor.Err()
}

func (r *ProjectRepository) GetDefaultProject(ctx context.Context) (*domain.Project, error) {
	return r.FindByID(ctx, domain.DefaultProjectID)
}

func docToProject(doc *projectDocument) *domain.Project {
	return &domain.Project{
		ID:                     doc.ID,
		CanonicalGitRepository: doc.CanonicalGitRepository,
		CreatedAt:              doc.CreatedAt,
	}
}
