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
	// Subagent (Task tool) related fields
	ParentSessionID *string // nullable - parent session ID for subagents
	AgentID         *string // nullable - subagent ID (e.g., "a5d7f46")
	IsSidechain     bool    // true if this session is a subagent
}
