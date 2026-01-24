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

type PlanCommentRepository struct {
	db *DB
}

func NewPlanCommentRepository(db *DB) *PlanCommentRepository {
	return &PlanCommentRepository{db: db}
}

type planCommentItem struct {
	ID               string `dynamodbav:"id"`
	PlanDocumentID   string `dynamodbav:"plan_document_id"`
	UserID           string `dynamodbav:"user_id"`
	TargetText       string `dynamodbav:"target_text"`
	ContextBefore    string `dynamodbav:"context_before"`
	ContextAfter     string `dynamodbav:"context_after"`
	OriginalBodyHash string `dynamodbav:"original_body_hash"`
	Content          string `dynamodbav:"content"`
	Status           string `dynamodbav:"status"`
	CreatedAt        string `dynamodbav:"created_at"`
	UpdatedAt        string `dynamodbav:"updated_at"`
}

func (r *PlanCommentRepository) Create(ctx context.Context, comment *domain.PlanComment) error {
	if comment.ID == "" {
		comment.ID = uuid.New().String()
	}
	now := time.Now()
	if comment.CreatedAt.IsZero() {
		comment.CreatedAt = now
	}
	if comment.UpdatedAt.IsZero() {
		comment.UpdatedAt = now
	}
	if comment.Status == "" {
		comment.Status = domain.PlanCommentStatusActive
	}

	item := planCommentItem{
		ID:               comment.ID,
		PlanDocumentID:   comment.PlanDocumentID,
		UserID:           comment.UserID,
		TargetText:       comment.TargetText,
		ContextBefore:    comment.ContextBefore,
		ContextAfter:     comment.ContextAfter,
		OriginalBodyHash: comment.OriginalBodyHash,
		Content:          comment.Content,
		Status:           string(comment.Status),
		CreatedAt:        comment.CreatedAt.Format(time.RFC3339Nano),
		UpdatedAt:        comment.UpdatedAt.Format(time.RFC3339Nano),
	}

	av, err := attributevalue.MarshalMap(item)
	if err != nil {
		return err
	}

	_, err = r.db.Client.PutItem(ctx, &dynamodb.PutItemInput{
		TableName: aws.String(r.db.TableName("plan_comments")),
		Item:      av,
	})
	return err
}

func (r *PlanCommentRepository) FindByID(ctx context.Context, id string) (*domain.PlanComment, error) {
	result, err := r.db.Client.GetItem(ctx, &dynamodb.GetItemInput{
		TableName: aws.String(r.db.TableName("plan_comments")),
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

	var item planCommentItem
	if err := attributevalue.UnmarshalMap(result.Item, &item); err != nil {
		return nil, err
	}

	return r.itemToPlanComment(&item), nil
}

func (r *PlanCommentRepository) FindByPlanDocumentID(ctx context.Context, planDocumentID string, status *domain.PlanCommentStatus) ([]*domain.PlanComment, error) {
	keyCond := expression.Key("plan_document_id").Equal(expression.Value(planDocumentID))
	builder := expression.NewBuilder().WithKeyCondition(keyCond)

	if status != nil {
		filter := expression.Name("status").Equal(expression.Value(string(*status)))
		builder = builder.WithFilter(filter)
	}

	expr, err := builder.Build()
	if err != nil {
		return nil, err
	}

	input := &dynamodb.QueryInput{
		TableName:                 aws.String(r.db.TableName("plan_comments")),
		IndexName:                 aws.String("plan_document_id-created_at-index"),
		KeyConditionExpression:    expr.KeyCondition(),
		ExpressionAttributeNames:  expr.Names(),
		ExpressionAttributeValues: expr.Values(),
		ScanIndexForward:          aws.Bool(true), // Ascending order by created_at
	}
	if expr.Filter() != nil {
		input.FilterExpression = expr.Filter()
	}

	result, err := r.db.Client.Query(ctx, input)
	if err != nil {
		return nil, err
	}

	var items []planCommentItem
	if err := attributevalue.UnmarshalListOfMaps(result.Items, &items); err != nil {
		return nil, err
	}

	comments := make([]*domain.PlanComment, 0, len(items))
	for _, item := range items {
		comments = append(comments, r.itemToPlanComment(&item))
	}

	return comments, nil
}

func (r *PlanCommentRepository) Update(ctx context.Context, comment *domain.PlanComment) error {
	comment.UpdatedAt = time.Now()

	update := expression.Set(expression.Name("content"), expression.Value(comment.Content)).
		Set(expression.Name("status"), expression.Value(string(comment.Status))).
		Set(expression.Name("updated_at"), expression.Value(comment.UpdatedAt.Format(time.RFC3339Nano)))
	expr, err := expression.NewBuilder().WithUpdate(update).Build()
	if err != nil {
		return err
	}

	_, err = r.db.Client.UpdateItem(ctx, &dynamodb.UpdateItemInput{
		TableName: aws.String(r.db.TableName("plan_comments")),
		Key: map[string]types.AttributeValue{
			"id": &types.AttributeValueMemberS{Value: comment.ID},
		},
		UpdateExpression:          expr.Update(),
		ExpressionAttributeNames:  expr.Names(),
		ExpressionAttributeValues: expr.Values(),
	})
	return err
}

func (r *PlanCommentRepository) Delete(ctx context.Context, id string) error {
	_, err := r.db.Client.DeleteItem(ctx, &dynamodb.DeleteItemInput{
		TableName: aws.String(r.db.TableName("plan_comments")),
		Key: map[string]types.AttributeValue{
			"id": &types.AttributeValueMemberS{Value: id},
		},
	})
	return err
}

func (r *PlanCommentRepository) MarkOutdatedByPlanDocumentID(ctx context.Context, planDocumentID string) error {
	// First, find all active comments for this plan document
	activeStatus := domain.PlanCommentStatusActive
	comments, err := r.FindByPlanDocumentID(ctx, planDocumentID, &activeStatus)
	if err != nil {
		return err
	}

	// Update each comment to outdated
	now := time.Now()
	for _, comment := range comments {
		comment.Status = domain.PlanCommentStatusOutdated
		comment.UpdatedAt = now
		if err := r.Update(ctx, comment); err != nil {
			return err
		}
	}

	return nil
}

func (r *PlanCommentRepository) itemToPlanComment(item *planCommentItem) *domain.PlanComment {
	createdAt, _ := time.Parse(time.RFC3339Nano, item.CreatedAt)
	updatedAt, _ := time.Parse(time.RFC3339Nano, item.UpdatedAt)

	return &domain.PlanComment{
		ID:               item.ID,
		PlanDocumentID:   item.PlanDocumentID,
		UserID:           item.UserID,
		TargetText:       item.TargetText,
		ContextBefore:    item.ContextBefore,
		ContextAfter:     item.ContextAfter,
		OriginalBodyHash: item.OriginalBodyHash,
		Content:          item.Content,
		Status:           domain.PlanCommentStatus(item.Status),
		CreatedAt:        createdAt,
		UpdatedAt:        updatedAt,
	}
}
