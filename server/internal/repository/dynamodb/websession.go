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

type WebSessionRepository struct {
	db *DB
}

func NewWebSessionRepository(db *DB) *WebSessionRepository {
	return &WebSessionRepository{db: db}
}

type webSessionItem struct {
	ID        string `dynamodbav:"id"`
	Token     string `dynamodbav:"token"`
	UserID    string `dynamodbav:"user_id"`
	ExpiresAt string `dynamodbav:"expires_at"`
	CreatedAt string `dynamodbav:"created_at"`
}

func (r *WebSessionRepository) Create(ctx context.Context, session *domain.WebSession) error {
	if session.ID == "" {
		session.ID = uuid.New().String()
	}
	if session.CreatedAt.IsZero() {
		session.CreatedAt = time.Now()
	}

	item := webSessionItem{
		ID:        session.ID,
		Token:     session.Token,
		UserID:    session.UserID,
		ExpiresAt: session.ExpiresAt.Format(time.RFC3339Nano),
		CreatedAt: session.CreatedAt.Format(time.RFC3339Nano),
	}

	av, err := attributevalue.MarshalMap(item)
	if err != nil {
		return err
	}

	_, err = r.db.Client.PutItem(ctx, &dynamodb.PutItemInput{
		TableName: aws.String(r.db.TableName("web_sessions")),
		Item:      av,
	})
	return err
}

func (r *WebSessionRepository) FindByToken(ctx context.Context, token string) (*domain.WebSession, error) {
	keyCond := expression.Key("token").Equal(expression.Value(token))
	expr, err := expression.NewBuilder().WithKeyCondition(keyCond).Build()
	if err != nil {
		return nil, err
	}

	result, err := r.db.Client.Query(ctx, &dynamodb.QueryInput{
		TableName:                 aws.String(r.db.TableName("web_sessions")),
		IndexName:                 aws.String("token-index"),
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

	var item webSessionItem
	if err := attributevalue.UnmarshalMap(result.Items[0], &item); err != nil {
		return nil, err
	}

	return r.itemToWebSession(&item), nil
}

func (r *WebSessionRepository) Delete(ctx context.Context, id string) error {
	_, err := r.db.Client.DeleteItem(ctx, &dynamodb.DeleteItemInput{
		TableName: aws.String(r.db.TableName("web_sessions")),
		Key: map[string]types.AttributeValue{
			"id": &types.AttributeValueMemberS{Value: id},
		},
	})
	return err
}

func (r *WebSessionRepository) DeleteExpired(ctx context.Context) error {
	now := time.Now().Format(time.RFC3339Nano)

	// Scan for expired sessions
	filterExpr := expression.Name("expires_at").LessThan(expression.Value(now))
	expr, err := expression.NewBuilder().WithFilter(filterExpr).Build()
	if err != nil {
		return err
	}

	result, err := r.db.Client.Scan(ctx, &dynamodb.ScanInput{
		TableName:                 aws.String(r.db.TableName("web_sessions")),
		FilterExpression:          expr.Filter(),
		ExpressionAttributeNames:  expr.Names(),
		ExpressionAttributeValues: expr.Values(),
	})
	if err != nil {
		return err
	}

	// Delete each expired session
	for _, item := range result.Items {
		id := item["id"].(*types.AttributeValueMemberS).Value
		if err := r.Delete(ctx, id); err != nil {
			return err
		}
	}

	return nil
}

func (r *WebSessionRepository) itemToWebSession(item *webSessionItem) *domain.WebSession {
	expiresAt, _ := time.Parse(time.RFC3339Nano, item.ExpiresAt)
	createdAt, _ := time.Parse(time.RFC3339Nano, item.CreatedAt)

	return &domain.WebSession{
		ID:        item.ID,
		Token:     item.Token,
		UserID:    item.UserID,
		ExpiresAt: expiresAt,
		CreatedAt: createdAt,
	}
}
