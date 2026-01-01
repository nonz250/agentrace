package domain

import "time"

type Session struct {
	ID              string
	UserID          *string // nullable - set when user is authenticated
	ProjectID       string  // reference to Project
	ClaudeSessionID string
	ProjectPath     string
	GitBranch       string // git current branch
	StartedAt       time.Time
	EndedAt         *time.Time
	CreatedAt       time.Time
}
