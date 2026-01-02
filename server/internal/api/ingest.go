package api

import (
	"encoding/json"
	"net/http"
	"time"

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
	GitRemoteURL    string                   `json:"git_remote_url"`
	GitBranch       string                   `json:"git_branch"`
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

	// Get user ID from context (set by auth middleware)
	var userID *string
	if uid := GetUserIDFromContext(ctx); uid != "" {
		userID = &uid
	}

	// Find or create session
	session, err := h.repos.Session.FindOrCreateByClaudeSessionID(ctx, req.SessionID, userID)
	if err != nil {
		http.Error(w, `{"error": "failed to create session"}`, http.StatusInternalServerError)
		return
	}

	// Update project path if provided and not already set
	if req.Cwd != "" && session.ProjectPath == "" {
		if err := h.repos.Session.UpdateProjectPath(ctx, session.ID, req.Cwd); err != nil {
			http.Error(w, `{"error": "failed to update project path"}`, http.StatusInternalServerError)
			return
		}
		session.ProjectPath = req.Cwd
	}

	// Update project and git branch if provided and project not already set
	if req.GitRemoteURL != "" && session.ProjectID == domain.DefaultProjectID {
		// Normalize the git URL and find or create the project
		canonicalURL := domain.NormalizeGitURL(req.GitRemoteURL)
		project, err := h.repos.Project.FindOrCreateByCanonicalGitRepository(ctx, canonicalURL)
		if err != nil {
			http.Error(w, `{"error": "failed to create project"}`, http.StatusInternalServerError)
			return
		}

		// Update session's project ID
		if err := h.repos.Session.UpdateProjectID(ctx, session.ID, project.ID); err != nil {
			http.Error(w, `{"error": "failed to update project"}`, http.StatusInternalServerError)
			return
		}
		session.ProjectID = project.ID
	}

	// Update git branch if provided and not already set
	if req.GitBranch != "" && session.GitBranch == "" {
		if err := h.repos.Session.UpdateGitBranch(ctx, session.ID, req.GitBranch); err != nil {
			http.Error(w, `{"error": "failed to update git branch"}`, http.StatusInternalServerError)
			return
		}
		session.GitBranch = req.GitBranch
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

	// Update session's updated_at timestamp if events were created
	if eventsCreated > 0 {
		_ = h.repos.Session.UpdateUpdatedAt(ctx, session.ID, time.Now())
	}

	resp := IngestResponse{
		OK:            true,
		EventsCreated: eventsCreated,
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(resp)
}
