package domain

import "time"

// PlanCommentMessage represents a message in a comment thread
type PlanCommentMessage struct {
	ID       string
	ThreadID string // reference to PlanCommentThread
	UserID   string // reference to User (required)

	Content string // メッセージ本文

	CreatedAt time.Time
	UpdatedAt time.Time
}
