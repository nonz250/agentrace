package dynamodb

import (
	"context"
	"time"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/feature/dynamodb/attributevalue"
	"github.com/aws/aws-sdk-go-v2/feature/dynamodb/expression"
	"github.com/aws/aws-sdk-go-v2/service/dynamodb"
	"github.com/aws/aws-sdk-go-v2/service/dynamodb/types"
	"github.com/google/uuid"
	"github.com/satetsu888/agentrace/server/internal/domain"
	"github.com/satetsu888/agentrace/server/internal/repository"
)

type ProjectRepository struct {
	db *DB
}

func NewProjectRepository(db *DB) *ProjectRepository {
	return &ProjectRepository{db: db}
}

type projectItem struct {
	ID                     string `dynamodbav:"id"`
	CanonicalGitRepository string `dynamodbav:"canonical_git_repository,omitempty"`
	CreatedAt              string `dynamodbav:"created_at"`
	GSIPK                  string `dynamodbav:"_gsi_pk"`
}

func (r *ProjectRepository) Create(ctx context.Context, project *domain.Project) error {
	if project.ID == "" {
		project.ID = uuid.New().String()
	}
	if project.CreatedAt.IsZero() {
		project.CreatedAt = time.Now()
	}

	item := projectItem{
		ID:                     project.ID,
		CanonicalGitRepository: project.CanonicalGitRepository,
		CreatedAt:              project.CreatedAt.Format(time.RFC3339Nano),
		GSIPK:                  projectGSIPK,
	}

	av, err := attributevalue.MarshalMap(item)
	if err != nil {
		return err
	}

	_, err = r.db.Client.PutItem(ctx, &dynamodb.PutItemInput{
		TableName: aws.String(r.db.TableName("projects")),
		Item:      av,
	})
	return err
}

func (r *ProjectRepository) FindByID(ctx context.Context, id string) (*domain.Project, error) {
	result, err := r.db.Client.GetItem(ctx, &dynamodb.GetItemInput{
		TableName: aws.String(r.db.TableName("projects")),
		Key: map[string]types.AttributeValue{
			"id": &types.AttributeValueMemberS{Value: id},
		},
	})
	if err != nil {
		return nil, err
	}
	if result.Item == nil {
		return nil, nil
	}

	var item projectItem
	if err := attributevalue.UnmarshalMap(result.Item, &item); err != nil {
		return nil, err
	}

	return r.itemToProject(&item), nil
}

func (r *ProjectRepository) FindByCanonicalGitRepository(ctx context.Context, canonicalGitRepo string) (*domain.Project, error) {
	keyCond := expression.Key("canonical_git_repository").Equal(expression.Value(canonicalGitRepo))
	expr, err := expression.NewBuilder().WithKeyCondition(keyCond).Build()
	if err != nil {
		return nil, err
	}

	result, err := r.db.Client.Query(ctx, &dynamodb.QueryInput{
		TableName:                 aws.String(r.db.TableName("projects")),
		IndexName:                 aws.String("canonical_git_repository-index"),
		KeyConditionExpression:    expr.KeyCondition(),
		ExpressionAttributeNames:  expr.Names(),
		ExpressionAttributeValues: expr.Values(),
		Limit:                     aws.Int32(1),
	})
	if err != nil {
		return nil, err
	}
	if len(result.Items) == 0 {
		return nil, nil
	}

	var item projectItem
	if err := attributevalue.UnmarshalMap(result.Items[0], &item); err != nil {
		return nil, err
	}

	return r.itemToProject(&item), nil
}

func (r *ProjectRepository) FindOrCreateByCanonicalGitRepository(ctx context.Context, canonicalGitRepo string) (*domain.Project, error) {
	project, err := r.FindByCanonicalGitRepository(ctx, canonicalGitRepo)
	if err != nil {
		return nil, err
	}
	if project != nil {
		return project, nil
	}

	newProject := &domain.Project{
		ID:                     uuid.New().String(),
		CanonicalGitRepository: canonicalGitRepo,
		CreatedAt:              time.Now(),
	}

	if err := r.Create(ctx, newProject); err != nil {
		// Handle race condition
		existingProject, findErr := r.FindByCanonicalGitRepository(ctx, canonicalGitRepo)
		if findErr != nil {
			return nil, err
		}
		if existingProject != nil {
			return existingProject, nil
		}
		return nil, err
	}

	return newProject, nil
}

func (r *ProjectRepository) FindAll(ctx context.Context, limit int, cursor string) ([]*domain.Project, string, error) {
	keyCond := expression.Key("_gsi_pk").Equal(expression.Value(projectGSIPK))
	expr, err := expression.NewBuilder().WithKeyCondition(keyCond).Build()
	if err != nil {
		return nil, "", err
	}

	input := &dynamodb.QueryInput{
		TableName:                 aws.String(r.db.TableName("projects")),
		IndexName:                 aws.String("gsi-created_at-index"),
		KeyConditionExpression:    expr.KeyCondition(),
		ExpressionAttributeNames:  expr.Names(),
		ExpressionAttributeValues: expr.Values(),
		ScanIndexForward:          aws.Bool(false), // DESC order
	}
	if limit > 0 {
		input.Limit = aws.Int32(int32(limit + 1))
	}

	// Use ExclusiveStartKey for cursor-based pagination
	if cursor != "" {
		cursorInfo := repository.DecodeCursor(cursor)
		if cursorInfo != nil {
			input.ExclusiveStartKey = map[string]types.AttributeValue{
				"id":         &types.AttributeValueMemberS{Value: cursorInfo.ID},
				"_gsi_pk":    &types.AttributeValueMemberS{Value: projectGSIPK},
				"created_at": &types.AttributeValueMemberS{Value: cursorInfo.SortValue},
			}
		}
	}

	result, err := r.db.Client.Query(ctx, input)
	if err != nil {
		return nil, "", err
	}

	var items []projectItem
	if err := attributevalue.UnmarshalListOfMaps(result.Items, &items); err != nil {
		return nil, "", err
	}

	projects := make([]*domain.Project, 0, len(items))
	for _, item := range items {
		projects = append(projects, r.itemToProject(&item))
	}

	// Generate cursor
	var nextCursor string
	if limit > 0 && len(projects) > limit {
		projects = projects[:limit]
		lastItem := projects[limit-1]
		nextCursor = repository.EncodeCursor(lastItem.CreatedAt, lastItem.ID)
	}

	return projects, nextCursor, nil
}

func (r *ProjectRepository) GetDefaultProject(ctx context.Context) (*domain.Project, error) {
	return r.FindByID(ctx, domain.DefaultProjectID)
}

func (r *ProjectRepository) itemToProject(item *projectItem) *domain.Project {
	createdAt, _ := time.Parse(time.RFC3339Nano, item.CreatedAt)
	return &domain.Project{
		ID:                     item.ID,
		CanonicalGitRepository: item.CanonicalGitRepository,
		CreatedAt:              createdAt,
	}
}

