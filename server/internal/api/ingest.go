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
	SessionID     string                 `json:"session_id"`
	HookEventName string                 `json:"hook_event_name"`
	ToolName      string                 `json:"tool_name"`
	ToolInput     map[string]interface{} `json:"tool_input"`
	ToolResponse  map[string]interface{} `json:"tool_response"`
	Cwd           string                 `json:"cwd"`
}

type IngestResponse struct {
	OK      bool   `json:"ok"`
	EventID string `json:"event_id"`
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

	// Create event
	payload := map[string]interface{}{
		"tool_input":    req.ToolInput,
		"tool_response": req.ToolResponse,
		"cwd":           req.Cwd,
	}

	event := &domain.Event{
		SessionID: session.ID,
		EventType: req.HookEventName,
		ToolName:  req.ToolName,
		Payload:   payload,
	}

	if err := h.repos.Event.Create(ctx, event); err != nil {
		http.Error(w, `{"error": "failed to create event"}`, http.StatusInternalServerError)
		return
	}

	resp := IngestResponse{
		OK:      true,
		EventID: event.ID,
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(resp)
}
