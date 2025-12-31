package domain

import "time"

type PlanDocumentEvent struct {
	ID             string
	PlanDocumentID string
	SessionID      *string // nullable - Claude Code session link
	UserID         *string // nullable - user who made the change
	Patch          string  // diff-match-patch format
	CreatedAt      time.Time
}
