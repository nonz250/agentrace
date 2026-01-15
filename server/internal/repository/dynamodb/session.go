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

type SessionRepository struct {
	db *DB
}

func NewSessionRepository(db *DB) *SessionRepository {
	return &SessionRepository{db: db}
}

type sessionItem struct {
	ID              string  `dynamodbav:"id"`
	UserID          *string `dynamodbav:"user_id,omitempty"`
	ProjectID       string  `dynamodbav:"project_id"`
	ClaudeSessionID string  `dynamodbav:"claude_session_id"`
	ProjectPath     string  `dynamodbav:"project_path,omitempty"`
	GitBranch       string  `dynamodbav:"git_branch,omitempty"`
	Title           *string `dynamodbav:"title,omitempty"`
	StartedAt       string  `dynamodbav:"started_at"`
	EndedAt         *string `dynamodbav:"ended_at,omitempty"`
	UpdatedAt       string  `dynamodbav:"updated_at"`
	CreatedAt       string  `dynamodbav:"created_at"`
	GSIPK           string  `dynamodbav:"_gsi_pk"` // Fixed value for global queries
}

const sessionGSIPK = "SESSION"

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

	var endedAt *string
	if session.EndedAt != nil {
		s := session.EndedAt.Format(time.RFC3339Nano)
		endedAt = &s
	}

	item := sessionItem{
		ID:              session.ID,
		UserID:          session.UserID,
		ProjectID:       session.ProjectID,
		ClaudeSessionID: session.ClaudeSessionID,
		ProjectPath:     session.ProjectPath,
		GitBranch:       session.GitBranch,
		Title:           session.Title,
		StartedAt:       session.StartedAt.Format(time.RFC3339Nano),
		EndedAt:         endedAt,
		UpdatedAt:       session.UpdatedAt.Format(time.RFC3339Nano),
		CreatedAt:       session.CreatedAt.Format(time.RFC3339Nano),
		GSIPK:           sessionGSIPK,
	}

	av, err := attributevalue.MarshalMap(item)
	if err != nil {
		return err
	}

	_, err = r.db.Client.PutItem(ctx, &dynamodb.PutItemInput{
		TableName: aws.String(r.db.TableName("sessions")),
		Item:      av,
	})
	return err
}

func (r *SessionRepository) FindByID(ctx context.Context, id string) (*domain.Session, error) {
	result, err := r.db.Client.GetItem(ctx, &dynamodb.GetItemInput{
		TableName: aws.String(r.db.TableName("sessions")),
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

	var item sessionItem
	if err := attributevalue.UnmarshalMap(result.Item, &item); err != nil {
		return nil, err
	}

	return r.itemToSession(&item), nil
}

func (r *SessionRepository) FindByClaudeSessionID(ctx context.Context, claudeSessionID string) (*domain.Session, error) {
	keyCond := expression.Key("claude_session_id").Equal(expression.Value(claudeSessionID))
	expr, err := expression.NewBuilder().WithKeyCondition(keyCond).Build()
	if err != nil {
		return nil, err
	}

	result, err := r.db.Client.Query(ctx, &dynamodb.QueryInput{
		TableName:                 aws.String(r.db.TableName("sessions")),
		IndexName:                 aws.String("claude_session_id-index"),
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

	var item sessionItem
	if err := attributevalue.UnmarshalMap(result.Items[0], &item); err != nil {
		return nil, err
	}

	return r.itemToSession(&item), nil
}

func (r *SessionRepository) FindAll(ctx context.Context, limit int, cursor string, sortBy string) ([]*domain.Session, string, error) {
	indexName := "gsi-updated_at-index"
	if sortBy == "created_at" {
		indexName = "gsi-created_at-index"
	}

	keyCond := expression.Key("_gsi_pk").Equal(expression.Value(sessionGSIPK))
	builder := expression.NewBuilder().WithKeyCondition(keyCond)

	if cursor != "" {
		cursorInfo := repository.DecodeCursor(cursor)
		if cursorInfo != nil {
			// Add filter for cursor
			sortAttr := "updated_at"
			if sortBy == "created_at" {
				sortAttr = "created_at"
			}
			filterExpr := expression.Or(
				expression.Name(sortAttr).LessThan(expression.Value(cursorInfo.SortValue)),
				expression.And(
					expression.Name(sortAttr).Equal(expression.Value(cursorInfo.SortValue)),
					expression.Name("id").LessThan(expression.Value(cursorInfo.ID)),
				),
			)
			builder = builder.WithFilter(filterExpr)
		}
	}

	expr, err := builder.Build()
	if err != nil {
		return nil, "", err
	}

	input := &dynamodb.QueryInput{
		TableName:                 aws.String(r.db.TableName("sessions")),
		IndexName:                 aws.String(indexName),
		KeyConditionExpression:    expr.KeyCondition(),
		ExpressionAttributeNames:  expr.Names(),
		ExpressionAttributeValues: expr.Values(),
		ScanIndexForward:          aws.Bool(false), // DESC order
	}
	if expr.Filter() != nil {
		input.FilterExpression = expr.Filter()
	}
	if limit > 0 {
		input.Limit = aws.Int32(int32(limit + 1))
	}

	result, err := r.db.Client.Query(ctx, input)
	if err != nil {
		return nil, "", err
	}

	var items []sessionItem
	if err := attributevalue.UnmarshalListOfMaps(result.Items, &items); err != nil {
		return nil, "", err
	}

	sessions := make([]*domain.Session, 0, len(items))
	for _, item := range items {
		sessions = append(sessions, r.itemToSession(&item))
	}

	// Generate cursor
	var nextCursor string
	if limit > 0 && len(sessions) > limit {
		sessions = sessions[:limit]
		lastItem := sessions[limit-1]
		var sortTime time.Time
		if sortBy == "created_at" {
			sortTime = lastItem.CreatedAt
		} else {
			sortTime = lastItem.UpdatedAt
		}
		nextCursor = repository.EncodeCursor(sortTime, lastItem.ID)
	}

	return sessions, nextCursor, nil
}

func (r *SessionRepository) FindByProjectID(ctx context.Context, projectID string, limit int, cursor string, sortBy string) ([]*domain.Session, string, error) {
	indexName := "project_id-updated_at-index"
	sortAttr := "updated_at"
	if sortBy == "created_at" {
		indexName = "project_id-created_at-index"
		sortAttr = "created_at"
	}

	keyCond := expression.Key("project_id").Equal(expression.Value(projectID))
	builder := expression.NewBuilder().WithKeyCondition(keyCond)

	if cursor != "" {
		cursorInfo := repository.DecodeCursor(cursor)
		if cursorInfo != nil {
			filterExpr := expression.Or(
				expression.Name(sortAttr).LessThan(expression.Value(cursorInfo.SortValue)),
				expression.And(
					expression.Name(sortAttr).Equal(expression.Value(cursorInfo.SortValue)),
					expression.Name("id").LessThan(expression.Value(cursorInfo.ID)),
				),
			)
			builder = builder.WithFilter(filterExpr)
		}
	}

	expr, err := builder.Build()
	if err != nil {
		return nil, "", err
	}

	input := &dynamodb.QueryInput{
		TableName:                 aws.String(r.db.TableName("sessions")),
		IndexName:                 aws.String(indexName),
		KeyConditionExpression:    expr.KeyCondition(),
		ExpressionAttributeNames:  expr.Names(),
		ExpressionAttributeValues: expr.Values(),
		ScanIndexForward:          aws.Bool(false),
	}
	if expr.Filter() != nil {
		input.FilterExpression = expr.Filter()
	}
	if limit > 0 {
		input.Limit = aws.Int32(int32(limit + 1))
	}

	result, err := r.db.Client.Query(ctx, input)
	if err != nil {
		return nil, "", err
	}

	var items []sessionItem
	if err := attributevalue.UnmarshalListOfMaps(result.Items, &items); err != nil {
		return nil, "", err
	}

	sessions := make([]*domain.Session, 0, len(items))
	for _, item := range items {
		sessions = append(sessions, r.itemToSession(&item))
	}

	var nextCursor string
	if limit > 0 && len(sessions) > limit {
		sessions = sessions[:limit]
		lastItem := sessions[limit-1]
		var sortTime time.Time
		if sortBy == "created_at" {
			sortTime = lastItem.CreatedAt
		} else {
			sortTime = lastItem.UpdatedAt
		}
		nextCursor = repository.EncodeCursor(sortTime, lastItem.ID)
	}

	return sessions, nextCursor, nil
}

func (r *SessionRepository) FindOrCreateByClaudeSessionID(ctx context.Context, claudeSessionID string, userID *string) (*domain.Session, error) {
	session, err := r.FindByClaudeSessionID(ctx, claudeSessionID)
	if err != nil {
		return nil, err
	}

	if session != nil {
		if userID != nil && session.UserID == nil {
			if err := r.UpdateUserID(ctx, session.ID, *userID); err != nil {
				return nil, err
			}
			session.UserID = userID
		}
		return session, nil
	}

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
	update := expression.Set(expression.Name("user_id"), expression.Value(userID))
	expr, err := expression.NewBuilder().WithUpdate(update).Build()
	if err != nil {
		return err
	}

	_, err = r.db.Client.UpdateItem(ctx, &dynamodb.UpdateItemInput{
		TableName: aws.String(r.db.TableName("sessions")),
		Key: map[string]types.AttributeValue{
			"id": &types.AttributeValueMemberS{Value: id},
		},
		UpdateExpression:          expr.Update(),
		ExpressionAttributeNames:  expr.Names(),
		ExpressionAttributeValues: expr.Values(),
	})
	return err
}

func (r *SessionRepository) UpdateProjectPath(ctx context.Context, id string, projectPath string) error {
	update := expression.Set(expression.Name("project_path"), expression.Value(projectPath))
	expr, err := expression.NewBuilder().WithUpdate(update).Build()
	if err != nil {
		return err
	}

	_, err = r.db.Client.UpdateItem(ctx, &dynamodb.UpdateItemInput{
		TableName: aws.String(r.db.TableName("sessions")),
		Key: map[string]types.AttributeValue{
			"id": &types.AttributeValueMemberS{Value: id},
		},
		UpdateExpression:          expr.Update(),
		ExpressionAttributeNames:  expr.Names(),
		ExpressionAttributeValues: expr.Values(),
	})
	return err
}

func (r *SessionRepository) UpdateProjectID(ctx context.Context, id string, projectID string) error {
	update := expression.Set(expression.Name("project_id"), expression.Value(projectID))
	expr, err := expression.NewBuilder().WithUpdate(update).Build()
	if err != nil {
		return err
	}

	_, err = r.db.Client.UpdateItem(ctx, &dynamodb.UpdateItemInput{
		TableName: aws.String(r.db.TableName("sessions")),
		Key: map[string]types.AttributeValue{
			"id": &types.AttributeValueMemberS{Value: id},
		},
		UpdateExpression:          expr.Update(),
		ExpressionAttributeNames:  expr.Names(),
		ExpressionAttributeValues: expr.Values(),
	})
	return err
}

func (r *SessionRepository) UpdateGitBranch(ctx context.Context, id string, gitBranch string) error {
	update := expression.Set(expression.Name("git_branch"), expression.Value(gitBranch))
	expr, err := expression.NewBuilder().WithUpdate(update).Build()
	if err != nil {
		return err
	}

	_, err = r.db.Client.UpdateItem(ctx, &dynamodb.UpdateItemInput{
		TableName: aws.String(r.db.TableName("sessions")),
		Key: map[string]types.AttributeValue{
			"id": &types.AttributeValueMemberS{Value: id},
		},
		UpdateExpression:          expr.Update(),
		ExpressionAttributeNames:  expr.Names(),
		ExpressionAttributeValues: expr.Values(),
	})
	return err
}

func (r *SessionRepository) UpdateTitle(ctx context.Context, id string, title string) error {
	update := expression.Set(expression.Name("title"), expression.Value(title))
	expr, err := expression.NewBuilder().WithUpdate(update).Build()
	if err != nil {
		return err
	}

	_, err = r.db.Client.UpdateItem(ctx, &dynamodb.UpdateItemInput{
		TableName: aws.String(r.db.TableName("sessions")),
		Key: map[string]types.AttributeValue{
			"id": &types.AttributeValueMemberS{Value: id},
		},
		UpdateExpression:          expr.Update(),
		ExpressionAttributeNames:  expr.Names(),
		ExpressionAttributeValues: expr.Values(),
	})
	return err
}

func (r *SessionRepository) UpdateUpdatedAt(ctx context.Context, id string, updatedAt time.Time) error {
	update := expression.Set(expression.Name("updated_at"), expression.Value(updatedAt.Format(time.RFC3339Nano)))
	expr, err := expression.NewBuilder().WithUpdate(update).Build()
	if err != nil {
		return err
	}

	_, err = r.db.Client.UpdateItem(ctx, &dynamodb.UpdateItemInput{
		TableName: aws.String(r.db.TableName("sessions")),
		Key: map[string]types.AttributeValue{
			"id": &types.AttributeValueMemberS{Value: id},
		},
		UpdateExpression:          expr.Update(),
		ExpressionAttributeNames:  expr.Names(),
		ExpressionAttributeValues: expr.Values(),
	})
	return err
}

func (r *SessionRepository) itemToSession(item *sessionItem) *domain.Session {
	startedAt, _ := time.Parse(time.RFC3339Nano, item.StartedAt)
	updatedAt, _ := time.Parse(time.RFC3339Nano, item.UpdatedAt)
	createdAt, _ := time.Parse(time.RFC3339Nano, item.CreatedAt)

	var endedAt *time.Time
	if item.EndedAt != nil {
		t, _ := time.Parse(time.RFC3339Nano, *item.EndedAt)
		endedAt = &t
	}

	return &domain.Session{
		ID:              item.ID,
		UserID:          item.UserID,
		ProjectID:       item.ProjectID,
		ClaudeSessionID: item.ClaudeSessionID,
		ProjectPath:     item.ProjectPath,
		GitBranch:       item.GitBranch,
		Title:           item.Title,
		StartedAt:       startedAt,
		EndedAt:         endedAt,
		UpdatedAt:       updatedAt,
		CreatedAt:       createdAt,
	}
}
