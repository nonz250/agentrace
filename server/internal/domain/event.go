package domain

import (
	"strings"
	"time"
)

type Event struct {
	ID        string
	SessionID string
	UUID      string // Claude Code transcript line UUID (unique per session)
	EventType string
	Payload   map[string]interface{}
	CreatedAt time.Time
}

// HiddenEventTypes defines event types that should be filtered from display and counts.
// These are internal events not useful for display:
// - file-history-snapshot: Claude Code internal file history tracking
// - system: Internal events (init, mcp_server_status, stop_hook_summary, etc.)
// - summary: Compact summary events
// - progress: Progress indicator events
var HiddenEventTypes = []string{
	"file-history-snapshot",
	"system",
	"summary",
	"progress",
}

// IsHiddenEventType checks if the given event type should be hidden
func IsHiddenEventType(eventType string) bool {
	for _, hidden := range HiddenEventTypes {
		if eventType == hidden {
			return true
		}
	}
	return false
}

// HiddenEventTypesSQL returns hidden event types formatted for SQL IN clause.
// e.g., "'file-history-snapshot', 'system', 'summary', 'progress'"
func HiddenEventTypesSQL() string {
	quoted := make([]string, len(HiddenEventTypes))
	for i, t := range HiddenEventTypes {
		quoted[i] = "'" + t + "'"
	}
	return strings.Join(quoted, ", ")
}

// SkippedEventTypes defines event types that should not be stored in the database.
// These events are high-volume and not needed for display or analysis:
// - progress: Real-time progress indicator events (very frequent)
// - file-history-snapshot: Claude Code internal file history tracking (large payload)
var SkippedEventTypes = []string{
	"progress",
	"file-history-snapshot",
}

// IsSkippedEventType checks if the given event type should be skipped from storage
func IsSkippedEventType(eventType string) bool {
	for _, skipped := range SkippedEventTypes {
		if eventType == skipped {
			return true
		}
	}
	return false
}
