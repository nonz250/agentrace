package domain

import "time"

type Session struct {
	ID              string
	ClaudeSessionID string
	ProjectPath     string
	StartedAt       time.Time
	EndedAt         *time.Time
	CreatedAt       time.Time
}
