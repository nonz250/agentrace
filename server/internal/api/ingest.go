package api

import (
	"encoding/json"
	"errors"
	"net/http"
	"strings"
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

		// Extract uuid from transcript line (Claude Code's unique identifier)
		if uuid, ok := line["uuid"].(string); ok {
			event.UUID = uuid
		}

		// Extract type if present
		eventType := ""
		if et, ok := line["type"].(string); ok {
			eventType = et
			event.EventType = et
		}

		// Auto-generate title from first user message if not set
		// Skip meta messages, command messages, and tool results
		if eventType == "user" && session.Title == nil && !isMetaMessage(line) {
			if text := extractUserMessageText(line); text != "" && isValidUserInput(text) {
				title := truncateString(text, 45)
				if err := h.repos.Session.UpdateTitle(ctx, session.ID, title); err == nil {
					session.Title = &title
				}
			}
		}

		if err := h.repos.Event.Create(ctx, event); err != nil {
			// Skip duplicate events (same uuid within session)
			if errors.Is(err, repository.ErrDuplicateEvent) {
				continue
			}
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

// isMetaMessage checks if the payload is a meta message (e.g., Caveat messages)
func isMetaMessage(payload map[string]interface{}) bool {
	if isMeta, ok := payload["isMeta"].(bool); ok && isMeta {
		return true
	}
	return false
}

// isValidUserInput checks if the text is a valid user input (not a command or system message)
func isValidUserInput(text string) bool {
	trimmed := strings.TrimSpace(text)
	if trimmed == "" {
		return false
	}

	// Skip command messages (e.g., <command-name>/clear</command-name>)
	if strings.HasPrefix(trimmed, "<command-name>") {
		return false
	}

	// Skip local command stdout (e.g., <local-command-stdout>...</local-command-stdout>)
	if strings.HasPrefix(trimmed, "<local-command-stdout>") {
		return false
	}

	// Skip system reminder messages
	if strings.HasPrefix(trimmed, "<system-reminder>") {
		return false
	}

	// Skip messages that start with slash commands
	if strings.HasPrefix(trimmed, "/") {
		return false
	}

	// Skip caveat messages
	if strings.HasPrefix(trimmed, "Caveat:") {
		return false
	}

	return true
}

// extractUserMessageText extracts text content from a user message payload
// Supports both string content and array content formats:
// - String: { "message": { "content": "text here" } }
// - Array: { "message": { "content": [{ "type": "text", "text": "..." }] } }
func extractUserMessageText(payload map[string]interface{}) string {
	message, ok := payload["message"].(map[string]interface{})
	if !ok {
		return ""
	}

	content := message["content"]

	// Handle string content (Claude Code's typical format)
	if contentStr, ok := content.(string); ok {
		return contentStr
	}

	// Handle array content (API format with content blocks)
	if contentArr, ok := content.([]interface{}); ok {
		for _, item := range contentArr {
			block, ok := item.(map[string]interface{})
			if !ok {
				continue
			}
			if blockType, ok := block["type"].(string); ok && blockType == "text" {
				if text, ok := block["text"].(string); ok {
					return text
				}
			}
		}
	}

	return ""
}

// truncateString truncates a string to maxLen characters (rune-aware)
func truncateString(s string, maxLen int) string {
	runes := []rune(s)
	if len(runes) <= maxLen {
		return s
	}
	return string(runes[:maxLen])
}
