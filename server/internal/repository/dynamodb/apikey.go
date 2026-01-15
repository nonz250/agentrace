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

type APIKeyRepository struct {
	db *DB
}

func NewAPIKeyRepository(db *DB) *APIKeyRepository {
	return &APIKeyRepository{db: db}
}

type apiKeyItem struct {
	ID         string  `dynamodbav:"id"`
	UserID     string  `dynamodbav:"user_id"`
	KeyHash    string  `dynamodbav:"key_hash"`
	Name       string  `dynamodbav:"name"`
	LastUsedAt *string `dynamodbav:"last_used_at,omitempty"`
	CreatedAt  string  `dynamodbav:"created_at"`
}

func (r *APIKeyRepository) Create(ctx context.Context, apiKey *domain.APIKey) error {
	if apiKey.ID == "" {
		apiKey.ID = uuid.New().String()
	}
	if apiKey.CreatedAt.IsZero() {
		apiKey.CreatedAt = time.Now()
	}

	var lastUsedAt *string
	if apiKey.LastUsedAt != nil {
		s := apiKey.LastUsedAt.Format(time.RFC3339Nano)
		lastUsedAt = &s
	}

	item := apiKeyItem{
		ID:         apiKey.ID,
		UserID:     apiKey.UserID,
		KeyHash:    apiKey.KeyHash,
		Name:       apiKey.Name,
		LastUsedAt: lastUsedAt,
		CreatedAt:  apiKey.CreatedAt.Format(time.RFC3339Nano),
	}

	av, err := attributevalue.MarshalMap(item)
	if err != nil {
		return err
	}

	_, err = r.db.Client.PutItem(ctx, &dynamodb.PutItemInput{
		TableName: aws.String(r.db.TableName("api_keys")),
		Item:      av,
	})
	return err
}

func (r *APIKeyRepository) FindByID(ctx context.Context, id string) (*domain.APIKey, error) {
	result, err := r.db.Client.GetItem(ctx, &dynamodb.GetItemInput{
		TableName: aws.String(r.db.TableName("api_keys")),
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

	var item apiKeyItem
	if err := attributevalue.UnmarshalMap(result.Item, &item); err != nil {
		return nil, err
	}

	return r.itemToAPIKey(&item), nil
}

func (r *APIKeyRepository) FindByUserID(ctx context.Context, userID string) ([]*domain.APIKey, error) {
	keyCond := expression.Key("user_id").Equal(expression.Value(userID))
	expr, err := expression.NewBuilder().WithKeyCondition(keyCond).Build()
	if err != nil {
		return nil, err
	}

	result, err := r.db.Client.Query(ctx, &dynamodb.QueryInput{
		TableName:                 aws.String(r.db.TableName("api_keys")),
		IndexName:                 aws.String("user_id-index"),
		KeyConditionExpression:    expr.KeyCondition(),
		ExpressionAttributeNames:  expr.Names(),
		ExpressionAttributeValues: expr.Values(),
	})
	if err != nil {
		return nil, err
	}

	var items []apiKeyItem
	if err := attributevalue.UnmarshalListOfMaps(result.Items, &items); err != nil {
		return nil, err
	}

	apiKeys := make([]*domain.APIKey, len(items))
	for i, item := range items {
		apiKeys[i] = r.itemToAPIKey(&item)
	}

	return apiKeys, nil
}

func (r *APIKeyRepository) FindByKeyHash(ctx context.Context, keyHash string) (*domain.APIKey, error) {
	keyCond := expression.Key("key_hash").Equal(expression.Value(keyHash))
	expr, err := expression.NewBuilder().WithKeyCondition(keyCond).Build()
	if err != nil {
		return nil, err
	}

	result, err := r.db.Client.Query(ctx, &dynamodb.QueryInput{
		TableName:                 aws.String(r.db.TableName("api_keys")),
		IndexName:                 aws.String("key_hash-index"),
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

	var item apiKeyItem
	if err := attributevalue.UnmarshalMap(result.Items[0], &item); err != nil {
		return nil, err
	}

	return r.itemToAPIKey(&item), nil
}

func (r *APIKeyRepository) Delete(ctx context.Context, id string) error {
	_, err := r.db.Client.DeleteItem(ctx, &dynamodb.DeleteItemInput{
		TableName: aws.String(r.db.TableName("api_keys")),
		Key: map[string]types.AttributeValue{
			"id": &types.AttributeValueMemberS{Value: id},
		},
	})
	return err
}

func (r *APIKeyRepository) UpdateLastUsedAt(ctx context.Context, id string) error {
	update := expression.Set(expression.Name("last_used_at"), expression.Value(time.Now().Format(time.RFC3339Nano)))
	expr, err := expression.NewBuilder().WithUpdate(update).Build()
	if err != nil {
		return err
	}

	_, err = r.db.Client.UpdateItem(ctx, &dynamodb.UpdateItemInput{
		TableName: aws.String(r.db.TableName("api_keys")),
		Key: map[string]types.AttributeValue{
			"id": &types.AttributeValueMemberS{Value: id},
		},
		UpdateExpression:          expr.Update(),
		ExpressionAttributeNames:  expr.Names(),
		ExpressionAttributeValues: expr.Values(),
	})
	return err
}

func (r *APIKeyRepository) itemToAPIKey(item *apiKeyItem) *domain.APIKey {
	createdAt, _ := time.Parse(time.RFC3339Nano, item.CreatedAt)

	var lastUsedAt *time.Time
	if item.LastUsedAt != nil {
		t, _ := time.Parse(time.RFC3339Nano, *item.LastUsedAt)
		lastUsedAt = &t
	}

	return &domain.APIKey{
		ID:         item.ID,
		UserID:     item.UserID,
		KeyHash:    item.KeyHash,
		Name:       item.Name,
		LastUsedAt: lastUsedAt,
		CreatedAt:  createdAt,
	}
}
