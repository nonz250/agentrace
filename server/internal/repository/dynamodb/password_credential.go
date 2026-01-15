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

type PasswordCredentialRepository struct {
	db *DB
}

func NewPasswordCredentialRepository(db *DB) *PasswordCredentialRepository {
	return &PasswordCredentialRepository{db: db}
}

type passwordCredentialItem struct {
	ID           string `dynamodbav:"id"`
	UserID       string `dynamodbav:"user_id"`
	PasswordHash string `dynamodbav:"password_hash"`
	CreatedAt    string `dynamodbav:"created_at"`
	UpdatedAt    string `dynamodbav:"updated_at"`
}

func (r *PasswordCredentialRepository) Create(ctx context.Context, cred *domain.PasswordCredential) error {
	if cred.ID == "" {
		cred.ID = uuid.New().String()
	}
	if cred.CreatedAt.IsZero() {
		cred.CreatedAt = time.Now()
	}
	if cred.UpdatedAt.IsZero() {
		cred.UpdatedAt = cred.CreatedAt
	}

	item := passwordCredentialItem{
		ID:           cred.ID,
		UserID:       cred.UserID,
		PasswordHash: cred.PasswordHash,
		CreatedAt:    cred.CreatedAt.Format(time.RFC3339Nano),
		UpdatedAt:    cred.UpdatedAt.Format(time.RFC3339Nano),
	}

	av, err := attributevalue.MarshalMap(item)
	if err != nil {
		return err
	}

	_, err = r.db.Client.PutItem(ctx, &dynamodb.PutItemInput{
		TableName: aws.String(r.db.TableName("password_credentials")),
		Item:      av,
	})
	return err
}

func (r *PasswordCredentialRepository) FindByUserID(ctx context.Context, userID string) (*domain.PasswordCredential, error) {
	keyCond := expression.Key("user_id").Equal(expression.Value(userID))
	expr, err := expression.NewBuilder().WithKeyCondition(keyCond).Build()
	if err != nil {
		return nil, err
	}

	result, err := r.db.Client.Query(ctx, &dynamodb.QueryInput{
		TableName:                 aws.String(r.db.TableName("password_credentials")),
		IndexName:                 aws.String("user_id-index"),
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

	var item passwordCredentialItem
	if err := attributevalue.UnmarshalMap(result.Items[0], &item); err != nil {
		return nil, err
	}

	return r.itemToPasswordCredential(&item), nil
}

func (r *PasswordCredentialRepository) Update(ctx context.Context, cred *domain.PasswordCredential) error {
	cred.UpdatedAt = time.Now()

	update := expression.Set(
		expression.Name("password_hash"), expression.Value(cred.PasswordHash),
	).Set(
		expression.Name("updated_at"), expression.Value(cred.UpdatedAt.Format(time.RFC3339Nano)),
	)
	expr, err := expression.NewBuilder().WithUpdate(update).Build()
	if err != nil {
		return err
	}

	_, err = r.db.Client.UpdateItem(ctx, &dynamodb.UpdateItemInput{
		TableName: aws.String(r.db.TableName("password_credentials")),
		Key: map[string]types.AttributeValue{
			"id": &types.AttributeValueMemberS{Value: cred.ID},
		},
		UpdateExpression:          expr.Update(),
		ExpressionAttributeNames:  expr.Names(),
		ExpressionAttributeValues: expr.Values(),
	})
	return err
}

func (r *PasswordCredentialRepository) Delete(ctx context.Context, id string) error {
	_, err := r.db.Client.DeleteItem(ctx, &dynamodb.DeleteItemInput{
		TableName: aws.String(r.db.TableName("password_credentials")),
		Key: map[string]types.AttributeValue{
			"id": &types.AttributeValueMemberS{Value: id},
		},
	})
	return err
}

func (r *PasswordCredentialRepository) itemToPasswordCredential(item *passwordCredentialItem) *domain.PasswordCredential {
	createdAt, _ := time.Parse(time.RFC3339Nano, item.CreatedAt)
	updatedAt, _ := time.Parse(time.RFC3339Nano, item.UpdatedAt)

	return &domain.PasswordCredential{
		ID:           item.ID,
		UserID:       item.UserID,
		PasswordHash: item.PasswordHash,
		CreatedAt:    createdAt,
		UpdatedAt:    updatedAt,
	}
}
