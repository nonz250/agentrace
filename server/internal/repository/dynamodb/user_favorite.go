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

type UserFavoriteRepository struct {
	db *DB
}

func NewUserFavoriteRepository(db *DB) *UserFavoriteRepository {
	return &UserFavoriteRepository{db: db}
}

type userFavoriteItem struct {
	ID                     string `dynamodbav:"id"`
	UserID                 string `dynamodbav:"user_id"`
	TargetType             string `dynamodbav:"target_type"`
	TargetID               string `dynamodbav:"target_id"`
	UserTargetTypeTargetID string `dynamodbav:"user_target_type_target_id"` // Composite key for unique lookup
	CreatedAt              string `dynamodbav:"created_at"`
}

func (r *UserFavoriteRepository) Create(ctx context.Context, fav *domain.UserFavorite) error {
	if fav.ID == "" {
		fav.ID = uuid.New().String()
	}
	if fav.CreatedAt.IsZero() {
		fav.CreatedAt = time.Now()
	}

	item := userFavoriteItem{
		ID:                     fav.ID,
		UserID:                 fav.UserID,
		TargetType:             string(fav.TargetType),
		TargetID:               fav.TargetID,
		UserTargetTypeTargetID: fav.UserID + "#" + string(fav.TargetType) + "#" + fav.TargetID,
		CreatedAt:              fav.CreatedAt.Format(time.RFC3339Nano),
	}

	av, err := attributevalue.MarshalMap(item)
	if err != nil {
		return err
	}

	_, err = r.db.Client.PutItem(ctx, &dynamodb.PutItemInput{
		TableName: aws.String(r.db.TableName("user_favorites")),
		Item:      av,
	})
	return err
}

func (r *UserFavoriteRepository) FindByUserID(ctx context.Context, userID string) ([]*domain.UserFavorite, error) {
	keyCond := expression.Key("user_id").Equal(expression.Value(userID))
	expr, err := expression.NewBuilder().WithKeyCondition(keyCond).Build()
	if err != nil {
		return nil, err
	}

	result, err := r.db.Client.Query(ctx, &dynamodb.QueryInput{
		TableName:                 aws.String(r.db.TableName("user_favorites")),
		IndexName:                 aws.String("user_id-index"),
		KeyConditionExpression:    expr.KeyCondition(),
		ExpressionAttributeNames:  expr.Names(),
		ExpressionAttributeValues: expr.Values(),
	})
	if err != nil {
		return nil, err
	}

	var items []userFavoriteItem
	if err := attributevalue.UnmarshalListOfMaps(result.Items, &items); err != nil {
		return nil, err
	}

	favorites := make([]*domain.UserFavorite, len(items))
	for i, item := range items {
		favorites[i] = r.itemToUserFavorite(&item)
	}

	return favorites, nil
}

func (r *UserFavoriteRepository) FindByUserAndTargetType(ctx context.Context, userID string, targetType domain.UserFavoriteTargetType) ([]*domain.UserFavorite, error) {
	keyCond := expression.Key("user_id").Equal(expression.Value(userID))
	filterExpr := expression.Name("target_type").Equal(expression.Value(string(targetType)))
	expr, err := expression.NewBuilder().
		WithKeyCondition(keyCond).
		WithFilter(filterExpr).
		Build()
	if err != nil {
		return nil, err
	}

	result, err := r.db.Client.Query(ctx, &dynamodb.QueryInput{
		TableName:                 aws.String(r.db.TableName("user_favorites")),
		IndexName:                 aws.String("user_id-index"),
		KeyConditionExpression:    expr.KeyCondition(),
		FilterExpression:          expr.Filter(),
		ExpressionAttributeNames:  expr.Names(),
		ExpressionAttributeValues: expr.Values(),
	})
	if err != nil {
		return nil, err
	}

	var items []userFavoriteItem
	if err := attributevalue.UnmarshalListOfMaps(result.Items, &items); err != nil {
		return nil, err
	}

	favorites := make([]*domain.UserFavorite, len(items))
	for i, item := range items {
		favorites[i] = r.itemToUserFavorite(&item)
	}

	return favorites, nil
}

func (r *UserFavoriteRepository) FindByUserAndTarget(ctx context.Context, userID string, targetType domain.UserFavoriteTargetType, targetID string) (*domain.UserFavorite, error) {
	compositeKey := userID + "#" + string(targetType) + "#" + targetID
	keyCond := expression.Key("user_target_type_target_id").Equal(expression.Value(compositeKey))
	expr, err := expression.NewBuilder().WithKeyCondition(keyCond).Build()
	if err != nil {
		return nil, err
	}

	result, err := r.db.Client.Query(ctx, &dynamodb.QueryInput{
		TableName:                 aws.String(r.db.TableName("user_favorites")),
		IndexName:                 aws.String("user_target_type_target_id-index"),
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

	var item userFavoriteItem
	if err := attributevalue.UnmarshalMap(result.Items[0], &item); err != nil {
		return nil, err
	}

	return r.itemToUserFavorite(&item), nil
}

func (r *UserFavoriteRepository) GetTargetIDs(ctx context.Context, userID string, targetType domain.UserFavoriteTargetType) ([]string, error) {
	favorites, err := r.FindByUserAndTargetType(ctx, userID, targetType)
	if err != nil {
		return nil, err
	}

	targetIDs := make([]string, len(favorites))
	for i, fav := range favorites {
		targetIDs[i] = fav.TargetID
	}

	return targetIDs, nil
}

func (r *UserFavoriteRepository) Delete(ctx context.Context, id string) error {
	_, err := r.db.Client.DeleteItem(ctx, &dynamodb.DeleteItemInput{
		TableName: aws.String(r.db.TableName("user_favorites")),
		Key: map[string]types.AttributeValue{
			"id": &types.AttributeValueMemberS{Value: id},
		},
	})
	return err
}

func (r *UserFavoriteRepository) DeleteByUserAndTarget(ctx context.Context, userID string, targetType domain.UserFavoriteTargetType, targetID string) error {
	fav, err := r.FindByUserAndTarget(ctx, userID, targetType, targetID)
	if err != nil {
		return err
	}
	if fav == nil {
		return nil
	}
	return r.Delete(ctx, fav.ID)
}

func (r *UserFavoriteRepository) itemToUserFavorite(item *userFavoriteItem) *domain.UserFavorite {
	createdAt, _ := time.Parse(time.RFC3339Nano, item.CreatedAt)

	return &domain.UserFavorite{
		ID:         item.ID,
		UserID:     item.UserID,
		TargetType: domain.UserFavoriteTargetType(item.TargetType),
		TargetID:   item.TargetID,
		CreatedAt:  createdAt,
	}
}
