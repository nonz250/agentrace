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

// SessionQuery はセッション検索のクエリ条件を表す
type SessionQuery struct {
	ProjectID string   // Filter by project ID (empty = all projects)
	UserIDs   []string // Filter by user IDs (empty = all users)
	Limit     int      // Max results (0 = use default)
	Cursor    string   // Cursor for pagination (empty = first page)
	SortBy    string   // Sort field: "updated_at" (default) or "created_at"
}
