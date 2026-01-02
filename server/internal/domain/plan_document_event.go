package domain

import "time"

type PlanDocumentEventType string

const (
	PlanDocumentEventTypeBodyChange   PlanDocumentEventType = "body_change"
	PlanDocumentEventTypeStatusChange PlanDocumentEventType = "status_change"
)

type PlanDocumentEvent struct {
	ID              string
	PlanDocumentID  string
	ClaudeSessionID *string               // nullable - Claude Code session ID
	UserID          *string               // nullable - user who made the change
	EventType       PlanDocumentEventType // body_change or status_change
	Patch           string                // diff-match-patch format (for body_change) or "old_status -> new_status" (for status_change)
	CreatedAt       time.Time
}
