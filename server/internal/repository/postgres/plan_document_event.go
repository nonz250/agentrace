package postgres

import (
	"context"
	"database/sql"
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

	_, err := r.db.ExecContext(ctx,
		`INSERT INTO plan_document_events (id, plan_document_id, session_id, user_id, patch, created_at)
		 VALUES ($1, $2, $3, $4, $5, $6)`,
		event.ID, event.PlanDocumentID, event.SessionID, event.UserID, event.Patch, event.CreatedAt,
	)
	return err
}

func (r *PlanDocumentEventRepository) FindByPlanDocumentID(ctx context.Context, planDocumentID string) ([]*domain.PlanDocumentEvent, error) {
	rows, err := r.db.QueryContext(ctx,
		`SELECT id, plan_document_id, session_id, user_id, patch, created_at
		 FROM plan_document_events WHERE plan_document_id = $1
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

func (r *PlanDocumentEventRepository) FindBySessionID(ctx context.Context, sessionID string) ([]*domain.PlanDocumentEvent, error) {
	rows, err := r.db.QueryContext(ctx,
		`SELECT id, plan_document_id, session_id, user_id, patch, created_at
		 FROM plan_document_events WHERE session_id = $1
		 ORDER BY created_at ASC`,
		sessionID,
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
		 WHERE plan_document_id = $1 AND user_id IS NOT NULL`,
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

func (r *PlanDocumentEventRepository) scanEvent(rows *sql.Rows) (*domain.PlanDocumentEvent, error) {
	var event domain.PlanDocumentEvent
	var sessionID, userID sql.NullString
	var createdAt sql.NullTime

	err := rows.Scan(&event.ID, &event.PlanDocumentID, &sessionID, &userID, &event.Patch, &createdAt)
	if err != nil {
		return nil, err
	}

	if sessionID.Valid {
		event.SessionID = &sessionID.String
	}
	if userID.Valid {
		event.UserID = &userID.String
	}
	if createdAt.Valid {
		event.CreatedAt = createdAt.Time
	}

	return &event, nil
}
