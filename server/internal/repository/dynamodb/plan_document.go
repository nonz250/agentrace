package dynamodb

import (
	"context"
	"strings"
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

type PlanDocumentRepository struct {
	db *DB
}

func NewPlanDocumentRepository(db *DB) *PlanDocumentRepository {
	return &PlanDocumentRepository{db: db}
}

type planDocumentItem struct {
	ID          string `dynamodbav:"id"`
	ProjectID   string `dynamodbav:"project_id"`
	Description string `dynamodbav:"description"`
	Body        string `dynamodbav:"body"`
	Status      string `dynamodbav:"status"`
	CreatedAt   string `dynamodbav:"created_at"`
	UpdatedAt   string `dynamodbav:"updated_at"`
	GSIPK       string `dynamodbav:"_gsi_pk"` // Fixed value for global queries
}

const planDocumentGSIPK = "PLANDOC"

func (r *PlanDocumentRepository) Create(ctx context.Context, doc *domain.PlanDocument) error {
	if doc.ID == "" {
		doc.ID = uuid.New().String()
	}
	now := time.Now()
	if doc.CreatedAt.IsZero() {
		doc.CreatedAt = now
	}
	if doc.UpdatedAt.IsZero() {
		doc.UpdatedAt = now
	}
	if doc.Status == "" {
		doc.Status = domain.PlanDocumentStatusPlanning
	}
	if doc.ProjectID == "" {
		doc.ProjectID = domain.DefaultProjectID
	}

	item := planDocumentItem{
		ID:          doc.ID,
		ProjectID:   doc.ProjectID,
		Description: doc.Description,
		Body:        doc.Body,
		Status:      string(doc.Status),
		CreatedAt:   doc.CreatedAt.Format(time.RFC3339Nano),
		UpdatedAt:   doc.UpdatedAt.Format(time.RFC3339Nano),
		GSIPK:       planDocumentGSIPK,
	}

	av, err := attributevalue.MarshalMap(item)
	if err != nil {
		return err
	}

	_, err = r.db.Client.PutItem(ctx, &dynamodb.PutItemInput{
		TableName: aws.String(r.db.TableName("plan_documents")),
		Item:      av,
	})
	return err
}

func (r *PlanDocumentRepository) FindByID(ctx context.Context, id string) (*domain.PlanDocument, error) {
	result, err := r.db.Client.GetItem(ctx, &dynamodb.GetItemInput{
		TableName: aws.String(r.db.TableName("plan_documents")),
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

	var item planDocumentItem
	if err := attributevalue.UnmarshalMap(result.Item, &item); err != nil {
		return nil, err
	}

	return r.itemToPlanDocument(&item), nil
}

func (r *PlanDocumentRepository) Find(ctx context.Context, query domain.PlanDocumentQuery) ([]*domain.PlanDocument, string, error) {
	sortAttr := "updated_at"
	if query.SortBy == "created_at" {
		sortAttr = "created_at"
	}

	// If filtering by project_id, use the project_id GSI
	if query.ProjectID != "" {
		return r.findByProjectID(ctx, query, sortAttr)
	}

	// Otherwise, use the global GSI
	return r.findAll(ctx, query, sortAttr)
}

func (r *PlanDocumentRepository) findByProjectID(ctx context.Context, query domain.PlanDocumentQuery, sortAttr string) ([]*domain.PlanDocument, string, error) {
	indexName := "project_id-updated_at-index"
	if sortAttr == "created_at" {
		indexName = "project_id-created_at-index"
	}

	keyCond := expression.Key("project_id").Equal(expression.Value(query.ProjectID))
	builder := expression.NewBuilder().WithKeyCondition(keyCond)

	// Build filter expressions (excluding cursor - handled by ExclusiveStartKey)
	var filterConditions []expression.ConditionBuilder

	if len(query.Statuses) > 0 {
		statusValues := make([]expression.OperandBuilder, len(query.Statuses))
		for i, s := range query.Statuses {
			statusValues[i] = expression.Value(string(s))
		}
		filterConditions = append(filterConditions, expression.Name("status").In(statusValues[0], statusValues[1:]...))
	}

	if query.DescriptionContains != "" {
		filterConditions = append(filterConditions, expression.Contains(expression.Name("description"), strings.ToLower(query.DescriptionContains)))
	}

	if len(query.PlanDocumentIDs) > 0 {
		idValues := make([]expression.OperandBuilder, len(query.PlanDocumentIDs))
		for i, id := range query.PlanDocumentIDs {
			idValues[i] = expression.Value(id)
		}
		filterConditions = append(filterConditions, expression.Name("id").In(idValues[0], idValues[1:]...))
	}

	if len(filterConditions) > 0 {
		var combinedFilter expression.ConditionBuilder
		combinedFilter = filterConditions[0]
		for i := 1; i < len(filterConditions); i++ {
			combinedFilter = expression.And(combinedFilter, filterConditions[i])
		}
		builder = builder.WithFilter(combinedFilter)
	}

	expr, err := builder.Build()
	if err != nil {
		return nil, "", err
	}

	input := &dynamodb.QueryInput{
		TableName:                 aws.String(r.db.TableName("plan_documents")),
		IndexName:                 aws.String(indexName),
		KeyConditionExpression:    expr.KeyCondition(),
		ExpressionAttributeNames:  expr.Names(),
		ExpressionAttributeValues: expr.Values(),
		ScanIndexForward:          aws.Bool(false),
	}
	if expr.Filter() != nil {
		input.FilterExpression = expr.Filter()
	}
	if query.Limit > 0 {
		input.Limit = aws.Int32(int32(query.Limit + 1))
	}

	// Use ExclusiveStartKey for cursor-based pagination
	if query.Cursor != "" {
		cursorInfo := repository.DecodeCursor(query.Cursor)
		if cursorInfo != nil {
			input.ExclusiveStartKey = map[string]types.AttributeValue{
				"id":         &types.AttributeValueMemberS{Value: cursorInfo.ID},
				"project_id": &types.AttributeValueMemberS{Value: query.ProjectID},
				sortAttr:     &types.AttributeValueMemberS{Value: cursorInfo.SortValue},
			}
		}
	}

	result, err := r.db.Client.Query(ctx, input)
	if err != nil {
		return nil, "", err
	}

	return r.processResults(result.Items, query.Limit, sortAttr)
}

func (r *PlanDocumentRepository) findAll(ctx context.Context, query domain.PlanDocumentQuery, sortAttr string) ([]*domain.PlanDocument, string, error) {
	indexName := "gsi-updated_at-index"
	if sortAttr == "created_at" {
		indexName = "gsi-created_at-index"
	}

	keyCond := expression.Key("_gsi_pk").Equal(expression.Value(planDocumentGSIPK))
	builder := expression.NewBuilder().WithKeyCondition(keyCond)

	// Build filter expressions (excluding cursor - handled by ExclusiveStartKey)
	var filterConditions []expression.ConditionBuilder

	if len(query.Statuses) > 0 {
		statusValues := make([]expression.OperandBuilder, len(query.Statuses))
		for i, s := range query.Statuses {
			statusValues[i] = expression.Value(string(s))
		}
		filterConditions = append(filterConditions, expression.Name("status").In(statusValues[0], statusValues[1:]...))
	}

	if query.DescriptionContains != "" {
		filterConditions = append(filterConditions, expression.Contains(expression.Name("description"), strings.ToLower(query.DescriptionContains)))
	}

	if len(query.PlanDocumentIDs) > 0 {
		idValues := make([]expression.OperandBuilder, len(query.PlanDocumentIDs))
		for i, id := range query.PlanDocumentIDs {
			idValues[i] = expression.Value(id)
		}
		filterConditions = append(filterConditions, expression.Name("id").In(idValues[0], idValues[1:]...))
	}

	if len(filterConditions) > 0 {
		var combinedFilter expression.ConditionBuilder
		combinedFilter = filterConditions[0]
		for i := 1; i < len(filterConditions); i++ {
			combinedFilter = expression.And(combinedFilter, filterConditions[i])
		}
		builder = builder.WithFilter(combinedFilter)
	}

	expr, err := builder.Build()
	if err != nil {
		return nil, "", err
	}

	input := &dynamodb.QueryInput{
		TableName:                 aws.String(r.db.TableName("plan_documents")),
		IndexName:                 aws.String(indexName),
		KeyConditionExpression:    expr.KeyCondition(),
		ExpressionAttributeNames:  expr.Names(),
		ExpressionAttributeValues: expr.Values(),
		ScanIndexForward:          aws.Bool(false),
	}
	if expr.Filter() != nil {
		input.FilterExpression = expr.Filter()
	}
	if query.Limit > 0 {
		input.Limit = aws.Int32(int32(query.Limit + 1))
	}

	// Use ExclusiveStartKey for cursor-based pagination
	if query.Cursor != "" {
		cursorInfo := repository.DecodeCursor(query.Cursor)
		if cursorInfo != nil {
			input.ExclusiveStartKey = map[string]types.AttributeValue{
				"id":      &types.AttributeValueMemberS{Value: cursorInfo.ID},
				"_gsi_pk": &types.AttributeValueMemberS{Value: planDocumentGSIPK},
				sortAttr:  &types.AttributeValueMemberS{Value: cursorInfo.SortValue},
			}
		}
	}

	result, err := r.db.Client.Query(ctx, input)
	if err != nil {
		return nil, "", err
	}

	return r.processResults(result.Items, query.Limit, sortAttr)
}

func (r *PlanDocumentRepository) processResults(items []map[string]types.AttributeValue, limit int, sortAttr string) ([]*domain.PlanDocument, string, error) {
	var planItems []planDocumentItem
	if err := attributevalue.UnmarshalListOfMaps(items, &planItems); err != nil {
		return nil, "", err
	}

	docs := make([]*domain.PlanDocument, 0, len(planItems))
	for _, item := range planItems {
		docs = append(docs, r.itemToPlanDocument(&item))
	}

	var nextCursor string
	if limit > 0 && len(docs) > limit {
		docs = docs[:limit]
		lastItem := docs[limit-1]
		var sortTime time.Time
		if sortAttr == "created_at" {
			sortTime = lastItem.CreatedAt
		} else {
			sortTime = lastItem.UpdatedAt
		}
		nextCursor = repository.EncodeCursor(sortTime, lastItem.ID)
	}

	return docs, nextCursor, nil
}

func (r *PlanDocumentRepository) Update(ctx context.Context, doc *domain.PlanDocument) error {
	doc.UpdatedAt = time.Now()

	update := expression.Set(expression.Name("project_id"), expression.Value(doc.ProjectID)).
		Set(expression.Name("description"), expression.Value(doc.Description)).
		Set(expression.Name("body"), expression.Value(doc.Body)).
		Set(expression.Name("status"), expression.Value(string(doc.Status))).
		Set(expression.Name("updated_at"), expression.Value(doc.UpdatedAt.Format(time.RFC3339Nano)))
	expr, err := expression.NewBuilder().WithUpdate(update).Build()
	if err != nil {
		return err
	}

	_, err = r.db.Client.UpdateItem(ctx, &dynamodb.UpdateItemInput{
		TableName: aws.String(r.db.TableName("plan_documents")),
		Key: map[string]types.AttributeValue{
			"id": &types.AttributeValueMemberS{Value: doc.ID},
		},
		UpdateExpression:          expr.Update(),
		ExpressionAttributeNames:  expr.Names(),
		ExpressionAttributeValues: expr.Values(),
	})
	return err
}

func (r *PlanDocumentRepository) Delete(ctx context.Context, id string) error {
	_, err := r.db.Client.DeleteItem(ctx, &dynamodb.DeleteItemInput{
		TableName: aws.String(r.db.TableName("plan_documents")),
		Key: map[string]types.AttributeValue{
			"id": &types.AttributeValueMemberS{Value: id},
		},
	})
	return err
}

func (r *PlanDocumentRepository) SetStatus(ctx context.Context, id string, status domain.PlanDocumentStatus) error {
	update := expression.Set(expression.Name("status"), expression.Value(string(status))).
		Set(expression.Name("updated_at"), expression.Value(time.Now().Format(time.RFC3339Nano)))
	expr, err := expression.NewBuilder().WithUpdate(update).Build()
	if err != nil {
		return err
	}

	_, err = r.db.Client.UpdateItem(ctx, &dynamodb.UpdateItemInput{
		TableName: aws.String(r.db.TableName("plan_documents")),
		Key: map[string]types.AttributeValue{
			"id": &types.AttributeValueMemberS{Value: id},
		},
		UpdateExpression:          expr.Update(),
		ExpressionAttributeNames:  expr.Names(),
		ExpressionAttributeValues: expr.Values(),
	})
	return err
}

func (r *PlanDocumentRepository) itemToPlanDocument(item *planDocumentItem) *domain.PlanDocument {
	createdAt, _ := time.Parse(time.RFC3339Nano, item.CreatedAt)
	updatedAt, _ := time.Parse(time.RFC3339Nano, item.UpdatedAt)

	return &domain.PlanDocument{
		ID:          item.ID,
		ProjectID:   item.ProjectID,
		Description: item.Description,
		Body:        item.Body,
		Status:      domain.PlanDocumentStatus(item.Status),
		CreatedAt:   createdAt,
		UpdatedAt:   updatedAt,
	}
}
