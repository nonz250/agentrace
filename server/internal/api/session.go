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
	ClaudeSessionID string  `json:"claude_session_id"`
	ProjectPath     string  `json:"project_path"`
	StartedAt       string  `json:"started_at"`
	EndedAt         *string `json:"ended_at"`
	CreatedAt       string  `json:"created_at"`
}

type SessionDetailResponse struct {
	SessionResponse
	Events []*EventResponse `json:"events"`
}

type EventResponse struct {
	ID        string                 `json:"id"`
	EventType string                 `json:"event_type"`
	ToolName  string                 `json:"tool_name"`
	Payload   map[string]interface{} `json:"payload"`
	CreatedAt string                 `json:"created_at"`
}

func sessionToResponse(s *domain.Session) *SessionResponse {
	var endedAt *string
	if s.EndedAt != nil {
		t := s.EndedAt.Format("2006-01-02T15:04:05Z07:00")
		endedAt = &t
	}
	return &SessionResponse{
		ID:              s.ID,
		ClaudeSessionID: s.ClaudeSessionID,
		ProjectPath:     s.ProjectPath,
		StartedAt:       s.StartedAt.Format("2006-01-02T15:04:05Z07:00"),
		EndedAt:         endedAt,
		CreatedAt:       s.CreatedAt.Format("2006-01-02T15:04:05Z07:00"),
	}
}

func eventToResponse(e *domain.Event) *EventResponse {
	return &EventResponse{
		ID:        e.ID,
		EventType: e.EventType,
		ToolName:  e.ToolName,
		Payload:   e.Payload,
		CreatedAt: e.CreatedAt.Format("2006-01-02T15:04:05Z07:00"),
	}
}

func (h *SessionHandler) List(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	sessions, err := h.repos.Session.FindAll(ctx, 100, 0)
	if err != nil {
		http.Error(w, `{"error": "failed to fetch sessions"}`, http.StatusInternalServerError)
		return
	}

	response := make([]*SessionResponse, len(sessions))
	for i, s := range sessions {
		response[i] = sessionToResponse(s)
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

	events, err := h.repos.Event.FindBySessionID(ctx, session.ID)
	if err != nil {
		http.Error(w, `{"error": "failed to fetch events"}`, http.StatusInternalServerError)
		return
	}

	eventResponses := make([]*EventResponse, len(events))
	for i, e := range events {
		eventResponses[i] = eventToResponse(e)
	}

	response := SessionDetailResponse{
		SessionResponse: *sessionToResponse(session),
		Events:          eventResponses,
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(response)
}
