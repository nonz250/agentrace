package domain

import "time"

type Event struct {
	ID        string
	SessionID string
	EventType string
	ToolName  string
	Payload   map[string]interface{}
	CreatedAt time.Time
}
