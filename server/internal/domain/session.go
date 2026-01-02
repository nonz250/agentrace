package domain

import "time"

type Session struct {
	ID              string
	UserID          *string // nullable - set when user is authenticated
	ProjectID       string  // reference to Project
	ClaudeSessionID string
	ProjectPath     string
	GitBranch       string  // git current branch
	Title           *string // nullable - auto-generated from first user message or manually set
	StartedAt       time.Time
	EndedAt         *time.Time
	UpdatedAt       time.Time // last activity time (updated when events are added)
	CreatedAt       time.Time
}
