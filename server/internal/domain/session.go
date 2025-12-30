package domain

import "time"

type Session struct {
	ID              string
	UserID          *string // nullable - set when user is authenticated
	ClaudeSessionID string
	ProjectPath     string
	GitRemoteURL    string // git remote origin URL
	GitBranch       string // git current branch
	StartedAt       time.Time
	EndedAt         *time.Time
	CreatedAt       time.Time
}
