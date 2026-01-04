package api

import (
	"context"
	"encoding/json"
	"net/http"
	"strconv"
	"strings"

	"github.com/gorilla/mux"
	"github.com/satetsu888/agentrace/server/internal/domain"
	"github.com/satetsu888/agentrace/server/internal/repository"
)

type PlanDocumentHandler struct {
	repos *repository.Repositories
}

func NewPlanDocumentHandler(repos *repository.Repositories) *PlanDocumentHandler {
	return &PlanDocumentHandler{repos: repos}
}

// Response types

type CollaboratorResponse struct {
	ID          string `json:"id"`
	DisplayName string `json:"display_name"`
}

type PlanDocumentProjectResponse struct {
	ID                     string `json:"id"`
	CanonicalGitRepository string `json:"canonical_git_repository"`
}

type PlanDocumentResponse struct {
	ID            string                       `json:"id"`
	Project       *PlanDocumentProjectResponse `json:"project"`
	Description   string                       `json:"description"`
	Body          string                       `json:"body"`
	Status        string                       `json:"status"`
	Collaborators []*CollaboratorResponse      `json:"collaborators"`
	CreatedAt     string                       `json:"created_at"`
	UpdatedAt     string                       `json:"updated_at"`
}

type PlanDocumentListResponse struct {
	Plans []*PlanDocumentResponse `json:"plans"`
}

type PlanDocumentEventResponse struct {
	ID              string  `json:"id"`
	PlanDocumentID  string  `json:"plan_document_id"`
	ClaudeSessionID *string `json:"claude_session_id"`
	ToolUseID       *string `json:"tool_use_id"`
	UserID          *string `json:"user_id"`
	UserName        *string `json:"user_name"`
	EventType       string  `json:"event_type"`
	Patch           string  `json:"patch"`
	CreatedAt       string  `json:"created_at"`
}

type PlanDocumentEventsResponse struct {
	Events []*PlanDocumentEventResponse `json:"events"`
}

// Request types

type CreatePlanDocumentRequest struct {
	Description     string  `json:"description"`
	Body            string  `json:"body"`
	ProjectID       *string `json:"project_id"`
	Status          *string `json:"status"`
	ClaudeSessionID *string `json:"claude_session_id"`
	ToolUseID       *string `json:"tool_use_id"`
}

type UpdatePlanDocumentRequest struct {
	Description     *string `json:"description"`
	Body            *string `json:"body"`
	Patch           *string `json:"patch"`
	ClaudeSessionID *string `json:"claude_session_id"`
	ToolUseID       *string `json:"tool_use_id"`
	ProjectID       *string `json:"project_id"`
}

type SetPlanDocumentStatusRequest struct {
	Status string `json:"status"`
}

// Helper functions

func (h *PlanDocumentHandler) planDocumentToResponse(ctx context.Context, doc *domain.PlanDocument) (*PlanDocumentResponse, error) {
	// Get collaborator user IDs from events
	userIDs, err := h.repos.PlanDocumentEvent.GetCollaboratorUserIDs(ctx, doc.ID)
	if err != nil {
		return nil, err
	}

	// Fetch user details
	collaborators := make([]*CollaboratorResponse, 0, len(userIDs))
	for _, userID := range userIDs {
		user, err := h.repos.User.FindByID(ctx, userID)
		if err == nil && user != nil {
			collaborators = append(collaborators, &CollaboratorResponse{
				ID:          user.ID,
				DisplayName: user.GetDisplayName(),
			})
		}
	}

	// Get project info
	var projectResp *PlanDocumentProjectResponse
	if doc.ProjectID != "" {
		project, err := h.repos.Project.FindByID(ctx, doc.ProjectID)
		if err == nil && project != nil {
			projectResp = &PlanDocumentProjectResponse{
				ID:                     project.ID,
				CanonicalGitRepository: project.CanonicalGitRepository,
			}
		}
	}

	return &PlanDocumentResponse{
		ID:            doc.ID,
		Project:       projectResp,
		Description:   doc.Description,
		Body:          doc.Body,
		Status:        string(doc.Status),
		Collaborators: collaborators,
		CreatedAt:     doc.CreatedAt.Format("2006-01-02T15:04:05Z07:00"),
		UpdatedAt:     doc.UpdatedAt.Format("2006-01-02T15:04:05Z07:00"),
	}, nil
}

func (h *PlanDocumentHandler) eventToResponse(ctx context.Context, event *domain.PlanDocumentEvent) *PlanDocumentEventResponse {
	var userName *string
	if event.UserID != nil {
		user, err := h.repos.User.FindByID(ctx, *event.UserID)
		if err == nil && user != nil {
			displayName := user.GetDisplayName()
			userName = &displayName
		}
	}

	eventType := string(event.EventType)
	if eventType == "" {
		eventType = string(domain.PlanDocumentEventTypeBodyChange)
	}

	return &PlanDocumentEventResponse{
		ID:              event.ID,
		PlanDocumentID:  event.PlanDocumentID,
		ClaudeSessionID: event.ClaudeSessionID,
		ToolUseID:       event.ToolUseID,
		UserID:          event.UserID,
		UserName:        userName,
		EventType:       eventType,
		Patch:           event.Patch,
		CreatedAt:       event.CreatedAt.Format("2006-01-02T15:04:05Z07:00"),
	}
}

// Handlers

// List returns all plan documents, optionally filtered by project_id, git_remote_url, or status
func (h *PlanDocumentHandler) List(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	// Parse query parameters
	limit := 100
	offset := 0
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

	projectID := r.URL.Query().Get("project_id")
	gitRemoteURL := r.URL.Query().Get("git_remote_url")
	statusParam := r.URL.Query().Get("status")
	descriptionParam := r.URL.Query().Get("description")

	// Parse status parameter (comma-separated)
	var statuses []domain.PlanDocumentStatus
	if statusParam != "" {
		statusStrs := strings.Split(statusParam, ",")
		for _, s := range statusStrs {
			s = strings.TrimSpace(s)
			status := domain.PlanDocumentStatus(s)
			if status.IsValid() {
				statuses = append(statuses, status)
			}
		}
	}

	// Resolve project ID from git_remote_url if needed
	if projectID == "" && gitRemoteURL != "" {
		canonicalURL := domain.NormalizeGitURL(gitRemoteURL)
		project, projErr := h.repos.Project.FindByCanonicalGitRepository(ctx, canonicalURL)
		if projErr != nil {
			http.Error(w, `{"error": "failed to find project"}`, http.StatusInternalServerError)
			return
		}
		if project != nil {
			projectID = project.ID
		}
	}

	// Use unified Find method with query object
	query := domain.PlanDocumentQuery{
		ProjectID:           projectID,
		Statuses:            statuses,
		DescriptionContains: descriptionParam,
		Limit:               limit,
		Offset:              offset,
	}
	docs, err := h.repos.PlanDocument.Find(ctx, query)

	if err != nil {
		http.Error(w, `{"error": "failed to fetch plan documents"}`, http.StatusInternalServerError)
		return
	}

	plans := make([]*PlanDocumentResponse, 0, len(docs))
	for _, doc := range docs {
		resp, err := h.planDocumentToResponse(ctx, doc)
		if err != nil {
			http.Error(w, `{"error": "failed to build response"}`, http.StatusInternalServerError)
			return
		}
		plans = append(plans, resp)
	}

	response := PlanDocumentListResponse{Plans: plans}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(response)
}

// Get returns a single plan document by ID
func (h *PlanDocumentHandler) Get(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	vars := mux.Vars(r)
	id := vars["id"]

	doc, err := h.repos.PlanDocument.FindByID(ctx, id)
	if err != nil {
		http.Error(w, `{"error": "failed to fetch plan document"}`, http.StatusInternalServerError)
		return
	}
	if doc == nil {
		http.Error(w, `{"error": "plan document not found"}`, http.StatusNotFound)
		return
	}

	resp, err := h.planDocumentToResponse(ctx, doc)
	if err != nil {
		http.Error(w, `{"error": "failed to build response"}`, http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(resp)
}

// GetEvents returns the change history for a plan document
func (h *PlanDocumentHandler) GetEvents(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	vars := mux.Vars(r)
	id := vars["id"]

	// First check if the plan document exists
	doc, err := h.repos.PlanDocument.FindByID(ctx, id)
	if err != nil {
		http.Error(w, `{"error": "failed to fetch plan document"}`, http.StatusInternalServerError)
		return
	}
	if doc == nil {
		http.Error(w, `{"error": "plan document not found"}`, http.StatusNotFound)
		return
	}

	events, err := h.repos.PlanDocumentEvent.FindByPlanDocumentID(ctx, id)
	if err != nil {
		http.Error(w, `{"error": "failed to fetch events"}`, http.StatusInternalServerError)
		return
	}

	eventResponses := make([]*PlanDocumentEventResponse, len(events))
	for i, event := range events {
		eventResponses[i] = h.eventToResponse(ctx, event)
	}

	response := PlanDocumentEventsResponse{Events: eventResponses}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(response)
}

// Create creates a new plan document
func (h *PlanDocumentHandler) Create(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	var req CreatePlanDocumentRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, `{"error": "invalid request body"}`, http.StatusBadRequest)
		return
	}

	if req.Description == "" {
		http.Error(w, `{"error": "description is required"}`, http.StatusBadRequest)
		return
	}

	// Get user ID from context (set by auth middleware)
	userID := GetUserIDFromContext(ctx)

	// Determine project ID: explicit project_id > session-based > default
	projectID := domain.DefaultProjectID
	if req.ProjectID != nil && *req.ProjectID != "" {
		// Use explicitly provided project ID
		projectID = *req.ProjectID
	} else if req.ClaudeSessionID != nil && *req.ClaudeSessionID != "" {
		// Fall back to session-based project ID
		session, err := h.repos.Session.FindByClaudeSessionID(ctx, *req.ClaudeSessionID)
		if err != nil {
			http.Error(w, `{"error": "failed to find session"}`, http.StatusInternalServerError)
			return
		}
		if session != nil && session.ProjectID != "" {
			projectID = session.ProjectID
		}
	}

	// Determine initial status (default: planning)
	status := domain.PlanDocumentStatusPlanning
	if req.Status != nil && *req.Status != "" {
		requestedStatus := domain.PlanDocumentStatus(*req.Status)
		if requestedStatus.IsValid() {
			status = requestedStatus
		}
	}

	doc := &domain.PlanDocument{
		ProjectID:   projectID,
		Description: req.Description,
		Body:        req.Body,
		Status:      status,
	}

	if err := h.repos.PlanDocument.Create(ctx, doc); err != nil {
		http.Error(w, `{"error": "failed to create plan document"}`, http.StatusInternalServerError)
		return
	}

	// Create initial event (empty patch for creation)
	event := &domain.PlanDocumentEvent{
		PlanDocumentID:  doc.ID,
		ClaudeSessionID: req.ClaudeSessionID,
		ToolUseID:       req.ToolUseID,
		Patch:           "", // Empty patch for initial creation
	}
	if userID != "" {
		event.UserID = &userID
	}

	if err := h.repos.PlanDocumentEvent.Create(ctx, event); err != nil {
		// Log error but don't fail the request
		// The document was created successfully
	}

	resp, err := h.planDocumentToResponse(ctx, doc)
	if err != nil {
		http.Error(w, `{"error": "failed to build response"}`, http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(resp)
}

// Update updates an existing plan document
func (h *PlanDocumentHandler) Update(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	vars := mux.Vars(r)
	id := vars["id"]

	doc, err := h.repos.PlanDocument.FindByID(ctx, id)
	if err != nil {
		http.Error(w, `{"error": "failed to fetch plan document"}`, http.StatusInternalServerError)
		return
	}
	if doc == nil {
		http.Error(w, `{"error": "plan document not found"}`, http.StatusNotFound)
		return
	}

	var req UpdatePlanDocumentRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, `{"error": "invalid request body"}`, http.StatusBadRequest)
		return
	}

	// Get user ID from context (set by auth middleware)
	userID := GetUserIDFromContext(ctx)

	// Update fields if provided
	if req.Description != nil {
		doc.Description = *req.Description
	}
	if req.Body != nil {
		doc.Body = *req.Body
	}
	if req.ProjectID != nil {
		doc.ProjectID = *req.ProjectID
	}

	if err := h.repos.PlanDocument.Update(ctx, doc); err != nil {
		http.Error(w, `{"error": "failed to update plan document"}`, http.StatusInternalServerError)
		return
	}

	// Create event with patch if provided
	if req.Patch != nil {
		event := &domain.PlanDocumentEvent{
			PlanDocumentID:  doc.ID,
			ClaudeSessionID: req.ClaudeSessionID,
			ToolUseID:       req.ToolUseID,
			Patch:           *req.Patch,
		}
		if userID != "" {
			event.UserID = &userID
		}

		if err := h.repos.PlanDocumentEvent.Create(ctx, event); err != nil {
			// Log error but don't fail the request
			// The document was updated successfully
		}
	}

	resp, err := h.planDocumentToResponse(ctx, doc)
	if err != nil {
		http.Error(w, `{"error": "failed to build response"}`, http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(resp)
}

// Delete deletes a plan document
func (h *PlanDocumentHandler) Delete(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	vars := mux.Vars(r)
	id := vars["id"]

	doc, err := h.repos.PlanDocument.FindByID(ctx, id)
	if err != nil {
		http.Error(w, `{"error": "failed to fetch plan document"}`, http.StatusInternalServerError)
		return
	}
	if doc == nil {
		http.Error(w, `{"error": "plan document not found"}`, http.StatusNotFound)
		return
	}

	if err := h.repos.PlanDocument.Delete(ctx, id); err != nil {
		http.Error(w, `{"error": "failed to delete plan document"}`, http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusNoContent)
}

// SetStatus sets the status of a plan document
func (h *PlanDocumentHandler) SetStatus(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	vars := mux.Vars(r)
	id := vars["id"]

	doc, err := h.repos.PlanDocument.FindByID(ctx, id)
	if err != nil {
		http.Error(w, `{"error": "failed to fetch plan document"}`, http.StatusInternalServerError)
		return
	}
	if doc == nil {
		http.Error(w, `{"error": "plan document not found"}`, http.StatusNotFound)
		return
	}

	var req SetPlanDocumentStatusRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, `{"error": "invalid request body"}`, http.StatusBadRequest)
		return
	}

	status := domain.PlanDocumentStatus(req.Status)
	if !status.IsValid() {
		http.Error(w, `{"error": "invalid status. must be one of: scratch, draft, planning, pending, ready, implementation, complete"}`, http.StatusBadRequest)
		return
	}

	// Store old status for event
	oldStatus := doc.Status

	if err := h.repos.PlanDocument.SetStatus(ctx, id, status); err != nil {
		http.Error(w, `{"error": "failed to update status"}`, http.StatusInternalServerError)
		return
	}

	// Get user ID from context (set by auth middleware)
	userID := GetUserIDFromContext(ctx)

	// Create status change event
	event := &domain.PlanDocumentEvent{
		PlanDocumentID: doc.ID,
		EventType:      domain.PlanDocumentEventTypeStatusChange,
		Patch:          string(oldStatus) + " -> " + string(status),
	}
	if userID != "" {
		event.UserID = &userID
	}
	// Ignore error - status was updated successfully
	h.repos.PlanDocumentEvent.Create(ctx, event)

	// Fetch updated document
	doc, err = h.repos.PlanDocument.FindByID(ctx, id)
	if err != nil {
		http.Error(w, `{"error": "failed to fetch updated plan document"}`, http.StatusInternalServerError)
		return
	}

	resp, err := h.planDocumentToResponse(ctx, doc)
	if err != nil {
		http.Error(w, `{"error": "failed to build response"}`, http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(resp)
}
