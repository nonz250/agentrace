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
)

type PlanCommentMessageRepository struct {
	db *DB
}

func NewPlanCommentMessageRepository(db *DB) *PlanCommentMessageRepository {
	return &PlanCommentMessageRepository{db: db}
}

type planCommentMessageItem struct {
	ID        string `dynamodbav:"id"`
	ThreadID  string `dynamodbav:"thread_id"`
	UserID    string `dynamodbav:"user_id"`
	Content   string `dynamodbav:"content"`
	CreatedAt string `dynamodbav:"created_at"`
	UpdatedAt string `dynamodbav:"updated_at"`
}

func (r *PlanCommentMessageRepository) Create(ctx context.Context, message *domain.PlanCommentMessage) error {
	if message.ID == "" {
		message.ID = uuid.New().String()
	}
	now := time.Now()
	if message.CreatedAt.IsZero() {
		message.CreatedAt = now
	}
	if message.UpdatedAt.IsZero() {
		message.UpdatedAt = now
	}

	item := planCommentMessageItem{
		ID:        message.ID,
		ThreadID:  message.ThreadID,
		UserID:    message.UserID,
		Content:   message.Content,
		CreatedAt: message.CreatedAt.Format(time.RFC3339Nano),
		UpdatedAt: message.UpdatedAt.Format(time.RFC3339Nano),
	}

	av, err := attributevalue.MarshalMap(item)
	if err != nil {
		return err
	}

	_, err = r.db.Client.PutItem(ctx, &dynamodb.PutItemInput{
		TableName: aws.String(r.db.TableName("plan_comment_messages")),
		Item:      av,
	})
	return err
}

func (r *PlanCommentMessageRepository) FindByID(ctx context.Context, id string) (*domain.PlanCommentMessage, error) {
	result, err := r.db.Client.GetItem(ctx, &dynamodb.GetItemInput{
		TableName: aws.String(r.db.TableName("plan_comment_messages")),
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

	var item planCommentMessageItem
	if err := attributevalue.UnmarshalMap(result.Item, &item); err != nil {
		return nil, err
	}

	return r.itemToMessage(&item), nil
}

func (r *PlanCommentMessageRepository) FindByThreadID(ctx context.Context, threadID string) ([]*domain.PlanCommentMessage, error) {
	keyCond := expression.Key("thread_id").Equal(expression.Value(threadID))
	expr, err := expression.NewBuilder().WithKeyCondition(keyCond).Build()
	if err != nil {
		return nil, err
	}

	result, err := r.db.Client.Query(ctx, &dynamodb.QueryInput{
		TableName:                 aws.String(r.db.TableName("plan_comment_messages")),
		IndexName:                 aws.String("thread_id-created_at-index"),
		KeyConditionExpression:    expr.KeyCondition(),
		ExpressionAttributeNames:  expr.Names(),
		ExpressionAttributeValues: expr.Values(),
		ScanIndexForward:          aws.Bool(true), // Ascending order by created_at
	})
	if err != nil {
		return nil, err
	}

	var items []planCommentMessageItem
	if err := attributevalue.UnmarshalListOfMaps(result.Items, &items); err != nil {
		return nil, err
	}

	messages := make([]*domain.PlanCommentMessage, 0, len(items))
	for _, item := range items {
		messages = append(messages, r.itemToMessage(&item))
	}

	return messages, nil
}

func (r *PlanCommentMessageRepository) Update(ctx context.Context, message *domain.PlanCommentMessage) error {
	message.UpdatedAt = time.Now()

	update := expression.Set(expression.Name("content"), expression.Value(message.Content)).
		Set(expression.Name("updated_at"), expression.Value(message.UpdatedAt.Format(time.RFC3339Nano)))
	expr, err := expression.NewBuilder().WithUpdate(update).Build()
	if err != nil {
		return err
	}

	_, err = r.db.Client.UpdateItem(ctx, &dynamodb.UpdateItemInput{
		TableName: aws.String(r.db.TableName("plan_comment_messages")),
		Key: map[string]types.AttributeValue{
			"id": &types.AttributeValueMemberS{Value: message.ID},
		},
		UpdateExpression:          expr.Update(),
		ExpressionAttributeNames:  expr.Names(),
		ExpressionAttributeValues: expr.Values(),
	})
	return err
}

func (r *PlanCommentMessageRepository) Delete(ctx context.Context, id string) error {
	_, err := r.db.Client.DeleteItem(ctx, &dynamodb.DeleteItemInput{
		TableName: aws.String(r.db.TableName("plan_comment_messages")),
		Key: map[string]types.AttributeValue{
			"id": &types.AttributeValueMemberS{Value: id},
		},
	})
	return err
}

func (r *PlanCommentMessageRepository) DeleteByThreadID(ctx context.Context, threadID string) error {
	// First, find all messages for this thread
	messages, err := r.FindByThreadID(ctx, threadID)
	if err != nil {
		return err
	}

	// Delete each message
	for _, msg := range messages {
		if err := r.Delete(ctx, msg.ID); err != nil {
			return err
		}
	}

	return nil
}

func (r *PlanCommentMessageRepository) itemToMessage(item *planCommentMessageItem) *domain.PlanCommentMessage {
	createdAt, _ := time.Parse(time.RFC3339Nano, item.CreatedAt)
	updatedAt, _ := time.Parse(time.RFC3339Nano, item.UpdatedAt)

	return &domain.PlanCommentMessage{
		ID:        item.ID,
		ThreadID:  item.ThreadID,
		UserID:    item.UserID,
		Content:   item.Content,
		CreatedAt: createdAt,
		UpdatedAt: updatedAt,
	}
}
