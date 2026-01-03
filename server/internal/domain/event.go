package domain

import "time"

type Event struct {
	ID        string
	SessionID string
	UUID      string // Claude Code transcript line UUID (unique per session)
	EventType string
	Payload   map[string]interface{}
	CreatedAt time.Time
}
