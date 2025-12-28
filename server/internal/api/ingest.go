package api

import (
	"encoding/json"
	"net/http"

	"github.com/satetsu888/agentrace/server/internal/domain"
	"github.com/satetsu888/agentrace/server/internal/repository"
)

type IngestHandler struct {
	repos *repository.Repositories
}

func NewIngestHandler(repos *repository.Repositories) *IngestHandler {
	return &IngestHandler{repos: repos}
}

type IngestRequest struct {
	SessionID       string                   `json:"session_id"`
	TranscriptLines []map[string]interface{} `json:"transcript_lines"`
	Cwd             string                   `json:"cwd"`
}

type IngestResponse struct {
	OK            bool `json:"ok"`
	EventsCreated int  `json:"events_created"`
}

func (h *IngestHandler) Handle(w http.ResponseWriter, r *http.Request) {
	var req IngestRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, `{"error": "invalid json"}`, http.StatusBadRequest)
		return
	}

	ctx := r.Context()

	// Find or create session
	session, err := h.repos.Session.FindOrCreateByClaudeSessionID(ctx, req.SessionID)
	if err != nil {
		http.Error(w, `{"error": "failed to create session"}`, http.StatusInternalServerError)
		return
	}

	// Update project path if provided
	if req.Cwd != "" && session.ProjectPath == "" {
		session.ProjectPath = req.Cwd
	}

	// Create events from transcript lines
	eventsCreated := 0
	for _, line := range req.TranscriptLines {
		event := &domain.Event{
			SessionID: session.ID,
			Payload:   line,
		}

		// Extract type if present
		if eventType, ok := line["type"].(string); ok {
			event.EventType = eventType
		}

		if err := h.repos.Event.Create(ctx, event); err != nil {
			http.Error(w, `{"error": "failed to create event"}`, http.StatusInternalServerError)
			return
		}
		eventsCreated++
	}

	resp := IngestResponse{
		OK:            true,
		EventsCreated: eventsCreated,
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(resp)
}
