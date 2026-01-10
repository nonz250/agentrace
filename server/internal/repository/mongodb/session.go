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

type SessionRepository struct {
	collection *mongo.Collection
}

func NewSessionRepository(db *DB) *SessionRepository {
	return &SessionRepository{
		collection: db.Collection("sessions"),
	}
}

type sessionDocument struct {
	ID              string     `bson:"_id"`
	UserID          *string    `bson:"user_id,omitempty"`
	ProjectID       string     `bson:"project_id"`
	ClaudeSessionID string     `bson:"claude_session_id"`
	ProjectPath     string     `bson:"project_path"`
	GitBranch       string     `bson:"git_branch"`
	Title           *string    `bson:"title,omitempty"`
	StartedAt       time.Time  `bson:"started_at"`
	EndedAt         *time.Time `bson:"ended_at,omitempty"`
	UpdatedAt       time.Time  `bson:"updated_at"`
	CreatedAt       time.Time  `bson:"created_at"`
}

func (r *SessionRepository) Create(ctx context.Context, session *domain.Session) error {
	if session.ID == "" {
		session.ID = uuid.New().String()
	}
	if session.CreatedAt.IsZero() {
		session.CreatedAt = time.Now()
	}
	if session.StartedAt.IsZero() {
		session.StartedAt = time.Now()
	}
	if session.UpdatedAt.IsZero() {
		session.UpdatedAt = session.StartedAt
	}
	if session.ProjectID == "" {
		session.ProjectID = domain.DefaultProjectID
	}

	doc := sessionDocument{
		ID:              session.ID,
		UserID:          session.UserID,
		ProjectID:       session.ProjectID,
		ClaudeSessionID: session.ClaudeSessionID,
		ProjectPath:     session.ProjectPath,
		GitBranch:       session.GitBranch,
		Title:           session.Title,
		StartedAt:       session.StartedAt,
		EndedAt:         session.EndedAt,
		UpdatedAt:       session.UpdatedAt,
		CreatedAt:       session.CreatedAt,
	}

	_, err := r.collection.InsertOne(ctx, doc)
	return err
}

func (r *SessionRepository) FindByID(ctx context.Context, id string) (*domain.Session, error) {
	var doc sessionDocument
	err := r.collection.FindOne(ctx, bson.M{"_id": id}).Decode(&doc)
	if err == mongo.ErrNoDocuments {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}

	return docToSession(&doc), nil
}

func (r *SessionRepository) FindByClaudeSessionID(ctx context.Context, claudeSessionID string) (*domain.Session, error) {
	var doc sessionDocument
	err := r.collection.FindOne(ctx, bson.M{"claude_session_id": claudeSessionID}).Decode(&doc)
	if err == mongo.ErrNoDocuments {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}

	return docToSession(&doc), nil
}

func (r *SessionRepository) FindAll(ctx context.Context, limit int, offset int, sortBy string) ([]*domain.Session, error) {
	// Validate sortBy
	sortField := "updated_at"
	if sortBy == "created_at" {
		sortField = "created_at"
	}

	opts := options.Find().SetSort(bson.D{{Key: sortField, Value: -1}})
	if limit > 0 {
		opts.SetLimit(int64(limit))
		opts.SetSkip(int64(offset))
	}

	cursor, err := r.collection.Find(ctx, bson.M{}, opts)
	if err != nil {
		return nil, err
	}
	defer cursor.Close(ctx)

	var sessions []*domain.Session
	for cursor.Next(ctx) {
		var doc sessionDocument
		if err := cursor.Decode(&doc); err != nil {
			return nil, err
		}
		sessions = append(sessions, docToSession(&doc))
	}

	return sessions, cursor.Err()
}

func (r *SessionRepository) FindByProjectID(ctx context.Context, projectID string, limit int, offset int, sortBy string) ([]*domain.Session, error) {
	// Validate sortBy
	sortField := "updated_at"
	if sortBy == "created_at" {
		sortField = "created_at"
	}

	opts := options.Find().SetSort(bson.D{{Key: sortField, Value: -1}})
	if limit > 0 {
		opts.SetLimit(int64(limit))
		opts.SetSkip(int64(offset))
	}

	cursor, err := r.collection.Find(ctx, bson.M{"project_id": projectID}, opts)
	if err != nil {
		return nil, err
	}
	defer cursor.Close(ctx)

	var sessions []*domain.Session
	for cursor.Next(ctx) {
		var doc sessionDocument
		if err := cursor.Decode(&doc); err != nil {
			return nil, err
		}
		sessions = append(sessions, docToSession(&doc))
	}

	return sessions, cursor.Err()
}

func (r *SessionRepository) FindOrCreateByClaudeSessionID(ctx context.Context, claudeSessionID string, userID *string) (*domain.Session, error) {
	// First try to find existing session
	var doc sessionDocument
	err := r.collection.FindOne(ctx, bson.M{"claude_session_id": claudeSessionID}).Decode(&doc)

	if err == nil {
		session := docToSession(&doc)
		// Update UserID if provided and not already set
		if userID != nil && session.UserID == nil {
			_, err := r.collection.UpdateOne(ctx,
				bson.M{"_id": session.ID},
				bson.M{"$set": bson.M{"user_id": *userID}},
			)
			if err != nil {
				return nil, err
			}
			session.UserID = userID
		}
		return session, nil
	}

	if err != mongo.ErrNoDocuments {
		return nil, err
	}

	// Create new session
	newSession := &domain.Session{
		ID:              uuid.New().String(),
		UserID:          userID,
		ProjectID:       domain.DefaultProjectID,
		ClaudeSessionID: claudeSessionID,
		StartedAt:       time.Now(),
		CreatedAt:       time.Now(),
	}

	if err := r.Create(ctx, newSession); err != nil {
		return nil, err
	}

	return newSession, nil
}

func (r *SessionRepository) UpdateUserID(ctx context.Context, id string, userID string) error {
	_, err := r.collection.UpdateOne(ctx,
		bson.M{"_id": id},
		bson.M{"$set": bson.M{"user_id": userID}},
	)
	return err
}

func (r *SessionRepository) UpdateProjectPath(ctx context.Context, id string, projectPath string) error {
	_, err := r.collection.UpdateOne(ctx,
		bson.M{"_id": id},
		bson.M{"$set": bson.M{"project_path": projectPath}},
	)
	return err
}

func (r *SessionRepository) UpdateProjectID(ctx context.Context, id string, projectID string) error {
	_, err := r.collection.UpdateOne(ctx,
		bson.M{"_id": id},
		bson.M{"$set": bson.M{"project_id": projectID}},
	)
	return err
}

func (r *SessionRepository) UpdateGitBranch(ctx context.Context, id string, gitBranch string) error {
	_, err := r.collection.UpdateOne(ctx,
		bson.M{"_id": id},
		bson.M{"$set": bson.M{"git_branch": gitBranch}},
	)
	return err
}

func (r *SessionRepository) UpdateTitle(ctx context.Context, id string, title string) error {
	_, err := r.collection.UpdateOne(ctx,
		bson.M{"_id": id},
		bson.M{"$set": bson.M{"title": title}},
	)
	return err
}

func (r *SessionRepository) UpdateUpdatedAt(ctx context.Context, id string, updatedAt time.Time) error {
	_, err := r.collection.UpdateOne(ctx,
		bson.M{"_id": id},
		bson.M{"$set": bson.M{"updated_at": updatedAt}},
	)
	return err
}

func docToSession(doc *sessionDocument) *domain.Session {
	projectID := doc.ProjectID
	if projectID == "" {
		projectID = domain.DefaultProjectID
	}

	return &domain.Session{
		ID:              doc.ID,
		UserID:          doc.UserID,
		ProjectID:       projectID,
		ClaudeSessionID: doc.ClaudeSessionID,
		ProjectPath:     doc.ProjectPath,
		GitBranch:       doc.GitBranch,
		Title:           doc.Title,
		StartedAt:       doc.StartedAt,
		EndedAt:         doc.EndedAt,
		UpdatedAt:       doc.UpdatedAt,
		CreatedAt:       doc.CreatedAt,
	}
}
