package api

import (
	"encoding/json"
	"net/http"

	"github.com/gorilla/mux"
	"github.com/satetsu888/agentrace/server/internal/domain"
	"github.com/satetsu888/agentrace/server/internal/repository"
)

type SessionHandler struct {
	repos *repository.Repositories
}

func NewSessionHandler(repos *repository.Repositories) *SessionHandler {
	return &SessionHandler{repos: repos}
}

type SessionResponse struct {
	ID              string  `json:"id"`
	UserID          *string `json:"user_id"`
	UserName        *string `json:"user_name"`
	ClaudeSessionID string  `json:"claude_session_id"`
	ProjectPath     string  `json:"project_path"`
	StartedAt       string  `json:"started_at"`
	EndedAt         *string `json:"ended_at"`
	EventCount      int     `json:"event_count"`
	CreatedAt       string  `json:"created_at"`
}

type SessionDetailResponse struct {
	SessionResponse
	Events []*EventResponse `json:"events"`
}

type EventResponse struct {
	ID        string                 `json:"id"`
	EventType string                 `json:"event_type"`
	Payload   map[string]interface{} `json:"payload"`
	CreatedAt string                 `json:"created_at"`
}

func sessionToResponse(s *domain.Session, userName *string, eventCount int) *SessionResponse {
	var endedAt *string
	if s.EndedAt != nil {
		t := s.EndedAt.Format("2006-01-02T15:04:05Z07:00")
		endedAt = &t
	}
	return &SessionResponse{
		ID:              s.ID,
		UserID:          s.UserID,
		UserName:        userName,
		ClaudeSessionID: s.ClaudeSessionID,
		ProjectPath:     s.ProjectPath,
		StartedAt:       s.StartedAt.Format("2006-01-02T15:04:05Z07:00"),
		EndedAt:         endedAt,
		EventCount:      eventCount,
		CreatedAt:       s.CreatedAt.Format("2006-01-02T15:04:05Z07:00"),
	}
}

func eventToResponse(e *domain.Event) *EventResponse {
	return &EventResponse{
		ID:        e.ID,
		EventType: e.EventType,
		Payload:   e.Payload,
		CreatedAt: e.CreatedAt.Format("2006-01-02T15:04:05Z07:00"),
	}
}

// shouldFilterEvent returns true if the event should be hidden from the response
func shouldFilterEvent(e *domain.Event) bool {
	payloadType, _ := e.Payload["type"].(string)

	// Filter out file-history-snapshot events
	if payloadType == "file-history-snapshot" {
		return true
	}

	// Filter out system events (internal events not useful for display)
	if payloadType == "system" {
		// All system subtypes are filtered for now:
		// - stop_hook_summary
		// - init
		// - mcp_server_status
		// - etc.
		return true
	}

	return false
}

// filterEvents returns events that should be displayed
func filterEvents(events []*domain.Event) []*domain.Event {
	filtered := make([]*domain.Event, 0, len(events))
	for _, e := range events {
		if !shouldFilterEvent(e) {
			filtered = append(filtered, e)
		}
	}
	return filtered
}

type SessionListResponse struct {
	Sessions []*SessionResponse `json:"sessions"`
}

func (h *SessionHandler) List(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	sessions, err := h.repos.Session.FindAll(ctx, 100, 0)
	if err != nil {
		http.Error(w, `{"error": "failed to fetch sessions"}`, http.StatusInternalServerError)
		return
	}

	sessionResponses := make([]*SessionResponse, len(sessions))
	for i, s := range sessions {
		// Get user name
		var userName *string
		if s.UserID != nil {
			user, err := h.repos.User.FindByID(ctx, *s.UserID)
			if err == nil && user != nil {
				userName = &user.Name
			}
		}

		// Get event count (filtered)
		events, err := h.repos.Event.FindBySessionID(ctx, s.ID)
		eventCount := 0
		if err == nil {
			eventCount = len(filterEvents(events))
		}

		sessionResponses[i] = sessionToResponse(s, userName, eventCount)
	}

	response := SessionListResponse{
		Sessions: sessionResponses,
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(response)
}

func (h *SessionHandler) Get(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	vars := mux.Vars(r)
	id := vars["id"]

	session, err := h.repos.Session.FindByID(ctx, id)
	if err != nil {
		http.Error(w, `{"error": "failed to fetch session"}`, http.StatusInternalServerError)
		return
	}
	if session == nil {
		http.Error(w, `{"error": "session not found"}`, http.StatusNotFound)
		return
	}

	// Get user name
	var userName *string
	if session.UserID != nil {
		user, err := h.repos.User.FindByID(ctx, *session.UserID)
		if err == nil && user != nil {
			userName = &user.Name
		}
	}

	events, err := h.repos.Event.FindBySessionID(ctx, session.ID)
	if err != nil {
		http.Error(w, `{"error": "failed to fetch events"}`, http.StatusInternalServerError)
		return
	}

	// Filter out internal events that shouldn't be displayed
	filteredEvents := filterEvents(events)

	eventResponses := make([]*EventResponse, len(filteredEvents))
	for i, e := range filteredEvents {
		eventResponses[i] = eventToResponse(e)
	}

	response := SessionDetailResponse{
		SessionResponse: *sessionToResponse(session, userName, len(filteredEvents)),
		Events:          eventResponses,
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(response)
}
