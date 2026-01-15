package dynamodb

import (
	"context"
	"encoding/json"
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

type EventRepository struct {
	db *DB
}

func NewEventRepository(db *DB) *EventRepository {
	return &EventRepository{db: db}
}

type eventItem struct {
	SessionID string `dynamodbav:"session_id"`
	SortKey   string `dynamodbav:"sort_key"` // created_at#id for chronological ordering
	ID        string `dynamodbav:"id"`
	EventType string `dynamodbav:"event_type"`
	Payload   string `dynamodbav:"payload"` // JSON string
	UUID      string `dynamodbav:"uuid"`
	CreatedAt string `dynamodbav:"created_at"`
}

func (r *EventRepository) Create(ctx context.Context, event *domain.Event) error {
	if event.ID == "" {
		event.ID = uuid.New().String()
	}
	if event.CreatedAt.IsZero() {
		event.CreatedAt = time.Now()
	}

	// Check for duplicate UUID within session
	if event.UUID != "" {
		existing, err := r.findByUUID(ctx, event.SessionID, event.UUID)
		if err != nil {
			return err
		}
		if existing != nil {
			return repository.ErrDuplicateEvent
		}
	}

	payloadJSON, err := json.Marshal(event.Payload)
	if err != nil {
		return err
	}

	createdAtStr := event.CreatedAt.Format(time.RFC3339Nano)
	sortKey := createdAtStr + "#" + event.ID // Chronological ordering with ID as tiebreaker

	item := eventItem{
		SessionID: event.SessionID,
		SortKey:   sortKey,
		ID:        event.ID,
		EventType: event.EventType,
		Payload:   string(payloadJSON),
		UUID:      event.UUID,
		CreatedAt: createdAtStr,
	}

	av, err := attributevalue.MarshalMap(item)
	if err != nil {
		return err
	}

	_, err = r.db.Client.PutItem(ctx, &dynamodb.PutItemInput{
		TableName: aws.String(r.db.TableName("events")),
		Item:      av,
	})
	return err
}

func (r *EventRepository) FindBySessionID(ctx context.Context, sessionID string) ([]*domain.Event, error) {
	keyCond := expression.Key("session_id").Equal(expression.Value(sessionID))
	expr, err := expression.NewBuilder().WithKeyCondition(keyCond).Build()
	if err != nil {
		return nil, err
	}

	result, err := r.db.Client.Query(ctx, &dynamodb.QueryInput{
		TableName:                 aws.String(r.db.TableName("events")),
		KeyConditionExpression:    expr.KeyCondition(),
		ExpressionAttributeNames:  expr.Names(),
		ExpressionAttributeValues: expr.Values(),
		ScanIndexForward:          aws.Bool(true), // Ascending order (chronological)
	})
	if err != nil {
		return nil, err
	}

	var items []eventItem
	if err := attributevalue.UnmarshalListOfMaps(result.Items, &items); err != nil {
		return nil, err
	}

	events := make([]*domain.Event, len(items))
	for i, item := range items {
		events[i] = r.itemToEvent(&item)
	}

	return events, nil
}

func (r *EventRepository) CountBySessionID(ctx context.Context, sessionID string) (int, error) {
	keyCond := expression.Key("session_id").Equal(expression.Value(sessionID))
	expr, err := expression.NewBuilder().WithKeyCondition(keyCond).Build()
	if err != nil {
		return 0, err
	}

	result, err := r.db.Client.Query(ctx, &dynamodb.QueryInput{
		TableName:                 aws.String(r.db.TableName("events")),
		KeyConditionExpression:    expr.KeyCondition(),
		ExpressionAttributeNames:  expr.Names(),
		ExpressionAttributeValues: expr.Values(),
		Select:                    types.SelectCount,
	})
	if err != nil {
		return 0, err
	}

	return int(result.Count), nil
}

func (r *EventRepository) findByUUID(ctx context.Context, sessionID, eventUUID string) (*domain.Event, error) {
	keyCond := expression.Key("session_id").Equal(expression.Value(sessionID))
	filterExpr := expression.Name("uuid").Equal(expression.Value(eventUUID))
	expr, err := expression.NewBuilder().
		WithKeyCondition(keyCond).
		WithFilter(filterExpr).
		Build()
	if err != nil {
		return nil, err
	}

	result, err := r.db.Client.Query(ctx, &dynamodb.QueryInput{
		TableName:                 aws.String(r.db.TableName("events")),
		KeyConditionExpression:    expr.KeyCondition(),
		FilterExpression:          expr.Filter(),
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

	var item eventItem
	if err := attributevalue.UnmarshalMap(result.Items[0], &item); err != nil {
		return nil, err
	}

	return r.itemToEvent(&item), nil
}

func (r *EventRepository) itemToEvent(item *eventItem) *domain.Event {
	createdAt, _ := time.Parse(time.RFC3339Nano, item.CreatedAt)

	var payload map[string]interface{}
	json.Unmarshal([]byte(item.Payload), &payload)

	return &domain.Event{
		ID:        item.ID,
		SessionID: item.SessionID,
		EventType: item.EventType,
		Payload:   payload,
		UUID:      item.UUID,
		CreatedAt: createdAt,
	}
}
