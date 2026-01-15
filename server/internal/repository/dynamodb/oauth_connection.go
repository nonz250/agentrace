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

type OAuthConnectionRepository struct {
	db *DB
}

func NewOAuthConnectionRepository(db *DB) *OAuthConnectionRepository {
	return &OAuthConnectionRepository{db: db}
}

type oauthConnectionItem struct {
	ID                 string `dynamodbav:"id"`
	UserID             string `dynamodbav:"user_id"`
	Provider           string `dynamodbav:"provider"`
	ProviderID         string `dynamodbav:"provider_id"`
	ProviderProviderID string `dynamodbav:"provider_provider_id"` // Composite key for GSI
	CreatedAt          string `dynamodbav:"created_at"`
}

func (r *OAuthConnectionRepository) Create(ctx context.Context, conn *domain.OAuthConnection) error {
	if conn.ID == "" {
		conn.ID = uuid.New().String()
	}
	if conn.CreatedAt.IsZero() {
		conn.CreatedAt = time.Now()
	}

	item := oauthConnectionItem{
		ID:                 conn.ID,
		UserID:             conn.UserID,
		Provider:           conn.Provider,
		ProviderID:         conn.ProviderID,
		ProviderProviderID: conn.Provider + "#" + conn.ProviderID,
		CreatedAt:          conn.CreatedAt.Format(time.RFC3339Nano),
	}

	av, err := attributevalue.MarshalMap(item)
	if err != nil {
		return err
	}

	_, err = r.db.Client.PutItem(ctx, &dynamodb.PutItemInput{
		TableName: aws.String(r.db.TableName("oauth_connections")),
		Item:      av,
	})
	return err
}

func (r *OAuthConnectionRepository) FindByUserID(ctx context.Context, userID string) ([]*domain.OAuthConnection, error) {
	keyCond := expression.Key("user_id").Equal(expression.Value(userID))
	expr, err := expression.NewBuilder().WithKeyCondition(keyCond).Build()
	if err != nil {
		return nil, err
	}

	result, err := r.db.Client.Query(ctx, &dynamodb.QueryInput{
		TableName:                 aws.String(r.db.TableName("oauth_connections")),
		IndexName:                 aws.String("user_id-index"),
		KeyConditionExpression:    expr.KeyCondition(),
		ExpressionAttributeNames:  expr.Names(),
		ExpressionAttributeValues: expr.Values(),
	})
	if err != nil {
		return nil, err
	}

	var items []oauthConnectionItem
	if err := attributevalue.UnmarshalListOfMaps(result.Items, &items); err != nil {
		return nil, err
	}

	connections := make([]*domain.OAuthConnection, len(items))
	for i, item := range items {
		connections[i] = r.itemToOAuthConnection(&item)
	}

	return connections, nil
}

func (r *OAuthConnectionRepository) FindByProviderAndProviderID(ctx context.Context, provider, providerID string) (*domain.OAuthConnection, error) {
	compositeKey := provider + "#" + providerID
	keyCond := expression.Key("provider_provider_id").Equal(expression.Value(compositeKey))
	expr, err := expression.NewBuilder().WithKeyCondition(keyCond).Build()
	if err != nil {
		return nil, err
	}

	result, err := r.db.Client.Query(ctx, &dynamodb.QueryInput{
		TableName:                 aws.String(r.db.TableName("oauth_connections")),
		IndexName:                 aws.String("provider_provider_id-index"),
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

	var item oauthConnectionItem
	if err := attributevalue.UnmarshalMap(result.Items[0], &item); err != nil {
		return nil, err
	}

	return r.itemToOAuthConnection(&item), nil
}

func (r *OAuthConnectionRepository) Delete(ctx context.Context, id string) error {
	_, err := r.db.Client.DeleteItem(ctx, &dynamodb.DeleteItemInput{
		TableName: aws.String(r.db.TableName("oauth_connections")),
		Key: map[string]types.AttributeValue{
			"id": &types.AttributeValueMemberS{Value: id},
		},
	})
	return err
}

func (r *OAuthConnectionRepository) itemToOAuthConnection(item *oauthConnectionItem) *domain.OAuthConnection {
	createdAt, _ := time.Parse(time.RFC3339Nano, item.CreatedAt)

	return &domain.OAuthConnection{
		ID:         item.ID,
		UserID:     item.UserID,
		Provider:   item.Provider,
		ProviderID: item.ProviderID,
		CreatedAt:  createdAt,
	}
}
