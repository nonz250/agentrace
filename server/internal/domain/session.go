package domain

import "time"

type Session struct {
	ID              string
	UserID          *string // nullable - set when user is authenticated
	ClaudeSessionID string
	ProjectPath     string
	StartedAt       time.Time
	EndedAt         *time.Time
	CreatedAt       time.Time
}
