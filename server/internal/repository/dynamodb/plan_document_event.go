package dynamodb

import (
	"context"
	"time"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/feature/dynamodb/attributevalue"
	"github.com/aws/aws-sdk-go-v2/feature/dynamodb/expression"
	"github.com/aws/aws-sdk-go-v2/service/dynamodb"
	"github.com/google/uuid"
	"github.com/satetsu888/agentrace/server/internal/domain"
)

type PlanDocumentEventRepository struct {
	db *DB
}

func NewPlanDocumentEventRepository(db *DB) *PlanDocumentEventRepository {
	return &PlanDocumentEventRepository{db: db}
}

type planDocumentEventItem struct {
	PlanDocumentID  string  `dynamodbav:"plan_document_id"`
	SortKey         string  `dynamodbav:"sort_key"` // created_at#id for chronological ordering
	ID              string  `dynamodbav:"id"`
	ClaudeSessionID *string `dynamodbav:"claude_session_id,omitempty"`
	ToolUseID       *string `dynamodbav:"tool_use_id,omitempty"`
	UserID          *string `dynamodbav:"user_id,omitempty"`
	EventType       string  `dynamodbav:"event_type"`
	Patch           string  `dynamodbav:"patch"`
	Message         string  `dynamodbav:"message,omitempty"`
	CreatedAt       string  `dynamodbav:"created_at"`
}

func (r *PlanDocumentEventRepository) Create(ctx context.Context, event *domain.PlanDocumentEvent) error {
	if event.ID == "" {
		event.ID = uuid.New().String()
	}
	if event.CreatedAt.IsZero() {
		event.CreatedAt = time.Now()
	}
	if event.EventType == "" {
		event.EventType = domain.PlanDocumentEventTypeBodyChange
	}

	createdAtStr := event.CreatedAt.Format(time.RFC3339Nano)
	sortKey := createdAtStr + "#" + event.ID // Chronological ordering with ID as tiebreaker

	item := planDocumentEventItem{
		PlanDocumentID:  event.PlanDocumentID,
		SortKey:         sortKey,
		ID:              event.ID,
		ClaudeSessionID: event.ClaudeSessionID,
		ToolUseID:       event.ToolUseID,
		UserID:          event.UserID,
		EventType:       string(event.EventType),
		Patch:           event.Patch,
		Message:         event.Message,
		CreatedAt:       createdAtStr,
	}

	av, err := attributevalue.MarshalMap(item)
	if err != nil {
		return err
	}

	_, err = r.db.Client.PutItem(ctx, &dynamodb.PutItemInput{
		TableName: aws.String(r.db.TableName("plan_document_events")),
		Item:      av,
	})
	return err
}

func (r *PlanDocumentEventRepository) FindByPlanDocumentID(ctx context.Context, planDocumentID string) ([]*domain.PlanDocumentEvent, error) {
	keyCond := expression.Key("plan_document_id").Equal(expression.Value(planDocumentID))
	expr, err := expression.NewBuilder().WithKeyCondition(keyCond).Build()
	if err != nil {
		return nil, err
	}

	result, err := r.db.Client.Query(ctx, &dynamodb.QueryInput{
		TableName:                 aws.String(r.db.TableName("plan_document_events")),
		KeyConditionExpression:    expr.KeyCondition(),
		ExpressionAttributeNames:  expr.Names(),
		ExpressionAttributeValues: expr.Values(),
		ScanIndexForward:          aws.Bool(true), // Ascending order (chronological)
	})
	if err != nil {
		return nil, err
	}

	var items []planDocumentEventItem
	if err := attributevalue.UnmarshalListOfMaps(result.Items, &items); err != nil {
		return nil, err
	}

	events := make([]*domain.PlanDocumentEvent, len(items))
	for i, item := range items {
		events[i] = r.itemToEvent(&item)
	}

	return events, nil
}

func (r *PlanDocumentEventRepository) FindByClaudeSessionID(ctx context.Context, claudeSessionID string) ([]*domain.PlanDocumentEvent, error) {
	keyCond := expression.Key("claude_session_id").Equal(expression.Value(claudeSessionID))
	expr, err := expression.NewBuilder().WithKeyCondition(keyCond).Build()
	if err != nil {
		return nil, err
	}

	result, err := r.db.Client.Query(ctx, &dynamodb.QueryInput{
		TableName:                 aws.String(r.db.TableName("plan_document_events")),
		IndexName:                 aws.String("claude_session_id-index"),
		KeyConditionExpression:    expr.KeyCondition(),
		ExpressionAttributeNames:  expr.Names(),
		ExpressionAttributeValues: expr.Values(),
	})
	if err != nil {
		return nil, err
	}

	var items []planDocumentEventItem
	if err := attributevalue.UnmarshalListOfMaps(result.Items, &items); err != nil {
		return nil, err
	}

	events := make([]*domain.PlanDocumentEvent, len(items))
	for i, item := range items {
		events[i] = r.itemToEvent(&item)
	}

	return events, nil
}

func (r *PlanDocumentEventRepository) GetCollaboratorUserIDs(ctx context.Context, planDocumentID string) ([]string, error) {
	events, err := r.FindByPlanDocumentID(ctx, planDocumentID)
	if err != nil {
		return nil, err
	}

	userIDSet := make(map[string]bool)
	for _, event := range events {
		if event.UserID != nil {
			userIDSet[*event.UserID] = true
		}
	}

	userIDs := make([]string, 0, len(userIDSet))
	for userID := range userIDSet {
		userIDs = append(userIDs, userID)
	}

	return userIDs, nil
}

func (r *PlanDocumentEventRepository) GetPlanDocumentIDsByUserIDs(ctx context.Context, userIDs []string) ([]string, error) {
	if len(userIDs) == 0 {
		return []string{}, nil
	}

	planDocIDSet := make(map[string]bool)

	for _, userID := range userIDs {
		keyCond := expression.Key("user_id").Equal(expression.Value(userID))
		expr, err := expression.NewBuilder().WithKeyCondition(keyCond).Build()
		if err != nil {
			return nil, err
		}

		result, err := r.db.Client.Query(ctx, &dynamodb.QueryInput{
			TableName:                 aws.String(r.db.TableName("plan_document_events")),
			IndexName:                 aws.String("user_id-index"),
			KeyConditionExpression:    expr.KeyCondition(),
			ExpressionAttributeNames:  expr.Names(),
			ExpressionAttributeValues: expr.Values(),
		})
		if err != nil {
			return nil, err
		}

		var items []planDocumentEventItem
		if err := attributevalue.UnmarshalListOfMaps(result.Items, &items); err != nil {
			return nil, err
		}

		for _, item := range items {
			planDocIDSet[item.PlanDocumentID] = true
		}
	}

	planDocIDs := make([]string, 0, len(planDocIDSet))
	for id := range planDocIDSet {
		planDocIDs = append(planDocIDs, id)
	}

	return planDocIDs, nil
}

func (r *PlanDocumentEventRepository) itemToEvent(item *planDocumentEventItem) *domain.PlanDocumentEvent {
	createdAt, _ := time.Parse(time.RFC3339Nano, item.CreatedAt)

	return &domain.PlanDocumentEvent{
		ID:              item.ID,
		PlanDocumentID:  item.PlanDocumentID,
		ClaudeSessionID: item.ClaudeSessionID,
		ToolUseID:       item.ToolUseID,
		UserID:          item.UserID,
		EventType:       domain.PlanDocumentEventType(item.EventType),
		Patch:           item.Patch,
		Message:         item.Message,
		CreatedAt:       createdAt,
	}
}
