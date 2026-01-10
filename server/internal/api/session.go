package api

import (
	"context"
	"encoding/json"
	"net/http"
	"strconv"

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
	ID              string           `json:"id"`
	UserID          *string          `json:"user_id"`
	UserName        *string          `json:"user_name"`
	Project         *ProjectResponse `json:"project"`
	ClaudeSessionID string           `json:"claude_session_id"`
	ProjectPath     string           `json:"project_path"`
	GitBranch       string           `json:"git_branch"`
	Title           *string          `json:"title"`
	StartedAt       string           `json:"started_at"`
	EndedAt         *string          `json:"ended_at"`
	UpdatedAt       string           `json:"updated_at"`
	EventCount      int              `json:"event_count"`
	CreatedAt       string           `json:"created_at"`
	IsFavorited     bool             `json:"is_favorited"`
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

func (h *SessionHandler) sessionToResponse(ctx context.Context, s *domain.Session, userName *string, eventCount int, isFavorited bool) *SessionResponse {
	var endedAt *string
	if s.EndedAt != nil {
		t := s.EndedAt.Format("2006-01-02T15:04:05Z07:00")
		endedAt = &t
	}

	// Get project info
	var projectResp *ProjectResponse
	if s.ProjectID != "" {
		project, err := h.repos.Project.FindByID(ctx, s.ProjectID)
		if err == nil && project != nil {
			projectResp = &ProjectResponse{
				ID:                     project.ID,
				CanonicalGitRepository: project.CanonicalGitRepository,
			}
		}
	}

	return &SessionResponse{
		ID:              s.ID,
		UserID:          s.UserID,
		UserName:        userName,
		Project:         projectResp,
		ClaudeSessionID: s.ClaudeSessionID,
		ProjectPath:     s.ProjectPath,
		GitBranch:       s.GitBranch,
		Title:           s.Title,
		StartedAt:       s.StartedAt.Format("2006-01-02T15:04:05Z07:00"),
		EndedAt:         endedAt,
		UpdatedAt:       s.UpdatedAt.Format("2006-01-02T15:04:05Z07:00"),
		EventCount:      eventCount,
		CreatedAt:       s.CreatedAt.Format("2006-01-02T15:04:05Z07:00"),
		IsFavorited:     isFavorited,
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
	userID := GetUserIDFromContext(ctx)

	// Parse query parameters
	limit := 100
	offset := 0
	projectID := r.URL.Query().Get("project_id")
	sortBy := r.URL.Query().Get("sort")
	// Validate sortBy - default to updated_at
	if sortBy != "created_at" {
		sortBy = "updated_at"
	}
	if limitStr := r.URL.Query().Get("limit"); limitStr != "" {
		if l, err := strconv.Atoi(limitStr); err == nil && l > 0 && l <= 100 {
			limit = l
		}
	}
	if offsetStr := r.URL.Query().Get("offset"); offsetStr != "" {
		if o, err := strconv.Atoi(offsetStr); err == nil && o >= 0 {
			offset = o
		}
	}

	var sessions []*domain.Session
	var err error
	if projectID != "" {
		sessions, err = h.repos.Session.FindByProjectID(ctx, projectID, limit, offset, sortBy)
	} else {
		sessions, err = h.repos.Session.FindAll(ctx, limit, offset, sortBy)
	}
	if err != nil {
		http.Error(w, `{"error": "failed to fetch sessions"}`, http.StatusInternalServerError)
		return
	}

	// Get favorited session IDs for the current user
	favoritedIDs := make(map[string]bool)
	if userID != "" {
		targetIDs, err := h.repos.UserFavorite.GetTargetIDs(ctx, userID, domain.UserFavoriteTargetTypeSession)
		if err == nil {
			for _, id := range targetIDs {
				favoritedIDs[id] = true
			}
		}
	}

	sessionResponses := make([]*SessionResponse, len(sessions))
	for i, s := range sessions {
		// Get user name
		var userName *string
		if s.UserID != nil {
			user, err := h.repos.User.FindByID(ctx, *s.UserID)
			if err == nil && user != nil {
				displayName := user.GetDisplayName()
				userName = &displayName
			}
		}

		// Get event count using COUNT query (much faster than fetching all events)
		eventCount, err := h.repos.Event.CountBySessionID(ctx, s.ID)
		if err != nil {
			eventCount = 0
		}

		isFavorited := favoritedIDs[s.ID]
		sessionResponses[i] = h.sessionToResponse(ctx, s, userName, eventCount, isFavorited)
	}

	// Sort: favorited sessions first, then by updated_at desc (already sorted by repo)
	sortSessionsByFavorite(sessionResponses)

	response := SessionListResponse{
		Sessions: sessionResponses,
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(response)
}

// sortSessionsByFavorite sorts sessions with favorited ones first, maintaining original order within each group
func sortSessionsByFavorite(sessions []*SessionResponse) {
	favorited := make([]*SessionResponse, 0)
	notFavorited := make([]*SessionResponse, 0)
	for _, s := range sessions {
		if s.IsFavorited {
			favorited = append(favorited, s)
		} else {
			notFavorited = append(notFavorited, s)
		}
	}
	copy(sessions, append(favorited, notFavorited...))
}

func (h *SessionHandler) Get(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	userID := GetUserIDFromContext(ctx)
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
			displayName := user.GetDisplayName()
			userName = &displayName
		}
	}

	// Check if favorited
	var isFavorited bool
	if userID != "" {
		fav, err := h.repos.UserFavorite.FindByUserAndTarget(ctx, userID, domain.UserFavoriteTargetTypeSession, id)
		if err == nil && fav != nil {
			isFavorited = true
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
		SessionResponse: *h.sessionToResponse(ctx, session, userName, len(filteredEvents), isFavorited),
		Events:          eventResponses,
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(response)
}

type UpdateSessionRequest struct {
	Title     *string `json:"title"`
	ProjectID *string `json:"project_id"`
}

func (h *SessionHandler) Update(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	userID := GetUserIDFromContext(ctx)
	vars := mux.Vars(r)
	id := vars["id"]

	var req UpdateSessionRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, `{"error": "invalid json"}`, http.StatusBadRequest)
		return
	}

	session, err := h.repos.Session.FindByID(ctx, id)
	if err != nil {
		http.Error(w, `{"error": "failed to fetch session"}`, http.StatusInternalServerError)
		return
	}
	if session == nil {
		http.Error(w, `{"error": "session not found"}`, http.StatusNotFound)
		return
	}

	// Update title if provided
	if req.Title != nil {
		if err := h.repos.Session.UpdateTitle(ctx, id, *req.Title); err != nil {
			http.Error(w, `{"error": "failed to update title"}`, http.StatusInternalServerError)
			return
		}
		session.Title = req.Title
	}

	// Update project_id if provided
	if req.ProjectID != nil {
		if err := h.repos.Session.UpdateProjectID(ctx, id, *req.ProjectID); err != nil {
			http.Error(w, `{"error": "failed to update project_id"}`, http.StatusInternalServerError)
			return
		}
		session.ProjectID = *req.ProjectID
	}

	// Get user name
	var userName *string
	if session.UserID != nil {
		user, err := h.repos.User.FindByID(ctx, *session.UserID)
		if err == nil && user != nil {
			displayName := user.GetDisplayName()
			userName = &displayName
		}
	}

	// Check if favorited
	var isFavorited bool
	if userID != "" {
		fav, err := h.repos.UserFavorite.FindByUserAndTarget(ctx, userID, domain.UserFavoriteTargetTypeSession, id)
		if err == nil && fav != nil {
			isFavorited = true
		}
	}

	// Get event count
	eventCount, err := h.repos.Event.CountBySessionID(ctx, session.ID)
	if err != nil {
		eventCount = 0
	}

	response := h.sessionToResponse(ctx, session, userName, eventCount, isFavorited)

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(response)
}
