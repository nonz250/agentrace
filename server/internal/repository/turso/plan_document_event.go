package turso

import (
	"context"
	"database/sql"
	"strings"
	"time"

	"github.com/google/uuid"
	"github.com/satetsu888/agentrace/server/internal/domain"
)

type PlanDocumentEventRepository struct {
	db *DB
}

func NewPlanDocumentEventRepository(db *DB) *PlanDocumentEventRepository {
	return &PlanDocumentEventRepository{db: db}
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

	_, err := r.db.ExecContext(ctx,
		`INSERT INTO plan_document_events (id, plan_document_id, claude_session_id, tool_use_id, user_id, event_type, patch, message, created_at)
		 VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?)`,
		event.ID, event.PlanDocumentID, event.ClaudeSessionID, event.ToolUseID, event.UserID, string(event.EventType), event.Patch, event.Message,
		event.CreatedAt.Format(time.RFC3339),
	)
	return err
}

func (r *PlanDocumentEventRepository) FindByPlanDocumentID(ctx context.Context, planDocumentID string) ([]*domain.PlanDocumentEvent, error) {
	rows, err := r.db.QueryContext(ctx,
		`SELECT id, plan_document_id, claude_session_id, tool_use_id, user_id, event_type, patch, message, created_at
		 FROM plan_document_events WHERE plan_document_id = ?
		 ORDER BY created_at ASC`,
		planDocumentID,
	)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var events []*domain.PlanDocumentEvent
	for rows.Next() {
		event, err := r.scanEvent(rows)
		if err != nil {
			return nil, err
		}
		events = append(events, event)
	}

	return events, rows.Err()
}

func (r *PlanDocumentEventRepository) FindByClaudeSessionID(ctx context.Context, claudeSessionID string) ([]*domain.PlanDocumentEvent, error) {
	rows, err := r.db.QueryContext(ctx,
		`SELECT id, plan_document_id, claude_session_id, tool_use_id, user_id, event_type, patch, message, created_at
		 FROM plan_document_events WHERE claude_session_id = ?
		 ORDER BY created_at ASC`,
		claudeSessionID,
	)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var events []*domain.PlanDocumentEvent
	for rows.Next() {
		event, err := r.scanEvent(rows)
		if err != nil {
			return nil, err
		}
		events = append(events, event)
	}

	return events, rows.Err()
}

func (r *PlanDocumentEventRepository) GetCollaboratorUserIDs(ctx context.Context, planDocumentID string) ([]string, error) {
	rows, err := r.db.QueryContext(ctx,
		`SELECT DISTINCT user_id FROM plan_document_events
		 WHERE plan_document_id = ? AND user_id IS NOT NULL`,
		planDocumentID,
	)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var userIDs []string
	for rows.Next() {
		var userID string
		if err := rows.Scan(&userID); err != nil {
			return nil, err
		}
		userIDs = append(userIDs, userID)
	}

	return userIDs, rows.Err()
}

func (r *PlanDocumentEventRepository) GetPlanDocumentIDsByUserIDs(ctx context.Context, userIDs []string) ([]string, error) {
	if len(userIDs) == 0 {
		return []string{}, nil
	}

	// Build placeholders for IN clause
	placeholders := make([]string, len(userIDs))
	args := make([]any, len(userIDs))
	for i, id := range userIDs {
		placeholders[i] = "?"
		args[i] = id
	}

	query := `SELECT DISTINCT plan_document_id FROM plan_document_events
		 WHERE user_id IN (` + strings.Join(placeholders, ", ") + `)`

	rows, err := r.db.QueryContext(ctx, query, args...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var planDocIDs []string
	for rows.Next() {
		var planDocID string
		if err := rows.Scan(&planDocID); err != nil {
			return nil, err
		}
		planDocIDs = append(planDocIDs, planDocID)
	}

	return planDocIDs, rows.Err()
}

func (r *PlanDocumentEventRepository) scanEvent(rows *sql.Rows) (*domain.PlanDocumentEvent, error) {
	var event domain.PlanDocumentEvent
	var claudeSessionID, toolUseID, userID, message sql.NullString
	var eventType string
	var createdAt string

	err := rows.Scan(&event.ID, &event.PlanDocumentID, &claudeSessionID, &toolUseID, &userID, &eventType, &event.Patch, &message, &createdAt)
	if err != nil {
		return nil, err
	}

	if claudeSessionID.Valid {
		event.ClaudeSessionID = &claudeSessionID.String
	}
	if toolUseID.Valid {
		event.ToolUseID = &toolUseID.String
	}
	if userID.Valid {
		event.UserID = &userID.String
	}
	if message.Valid {
		event.Message = message.String
	}
	event.EventType = domain.PlanDocumentEventType(eventType)
	if event.EventType == "" {
		event.EventType = domain.PlanDocumentEventTypeBodyChange
	}
	event.CreatedAt, _ = time.Parse(time.RFC3339, createdAt)

	return &event, nil
}
