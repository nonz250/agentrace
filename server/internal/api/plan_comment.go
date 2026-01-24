package api

import (
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"net/http"
	"strings"

	"github.com/gorilla/mux"
	"github.com/satetsu888/agentrace/server/internal/domain"
	"github.com/satetsu888/agentrace/server/internal/repository"
)

type PlanCommentHandler struct {
	repos *repository.Repositories
}

func NewPlanCommentHandler(repos *repository.Repositories) *PlanCommentHandler {
	return &PlanCommentHandler{repos: repos}
}

// Response types

type CommentPositionResponse struct {
	StartOffset int  `json:"start_offset"`
	EndOffset   int  `json:"end_offset"`
	StartLine   int  `json:"start_line"`
	StartColumn int  `json:"start_column"`
	EndLine     int  `json:"end_line"`
	EndColumn   int  `json:"end_column"`
	Found       bool `json:"found"`
}

type PlanCommentResponse struct {
	ID             string                   `json:"id"`
	PlanDocumentID string                   `json:"plan_document_id"`
	UserID         string                   `json:"user_id"`
	UserName       string                   `json:"user_name"`
	TargetText     string                   `json:"target_text"`
	Content        string                   `json:"content"`
	Status         string                   `json:"status"`
	Position       *CommentPositionResponse `json:"position"`
	CreatedAt      string                   `json:"created_at"`
	UpdatedAt      string                   `json:"updated_at"`
}

type PlanCommentListResponse struct {
	Comments []*PlanCommentResponse `json:"comments"`
}

// Request types

type CreatePlanCommentRequest struct {
	TargetText    string `json:"target_text"`
	ContextBefore string `json:"context_before"`
	ContextAfter  string `json:"context_after"`
	Content       string `json:"content"`
}

type UpdatePlanCommentRequest struct {
	Content *string `json:"content"`
}

// Helper functions

func (h *PlanCommentHandler) commentToResponse(comment *domain.PlanComment, userName string, position *domain.CommentPosition) *PlanCommentResponse {
	resp := &PlanCommentResponse{
		ID:             comment.ID,
		PlanDocumentID: comment.PlanDocumentID,
		UserID:         comment.UserID,
		UserName:       userName,
		TargetText:     comment.TargetText,
		Content:        comment.Content,
		Status:         string(comment.Status),
		CreatedAt:      comment.CreatedAt.Format("2006-01-02T15:04:05Z07:00"),
		UpdatedAt:      comment.UpdatedAt.Format("2006-01-02T15:04:05Z07:00"),
	}

	if position != nil {
		resp.Position = &CommentPositionResponse{
			StartOffset: position.StartOffset,
			EndOffset:   position.EndOffset,
			StartLine:   position.StartLine,
			StartColumn: position.StartColumn,
			EndLine:     position.EndLine,
			EndColumn:   position.EndColumn,
			Found:       position.Found,
		}
	}

	return resp
}

// offsetToLineColumn converts a character offset to line and column numbers (1-indexed)
func offsetToLineColumn(body string, offset int) (line, column int) {
	line = 1
	column = 1
	for i, ch := range body {
		if i >= offset {
			break
		}
		if ch == '\n' {
			line++
			column = 1
		} else {
			column++
		}
	}
	return line, column
}

// createPosition creates a CommentPosition with offset and line/column information
func createPosition(body string, startOffset, endOffset int) *domain.CommentPosition {
	startLine, startColumn := offsetToLineColumn(body, startOffset)
	endLine, endColumn := offsetToLineColumn(body, endOffset)
	return &domain.CommentPosition{
		StartOffset: startOffset,
		EndOffset:   endOffset,
		StartLine:   startLine,
		StartColumn: startColumn,
		EndLine:     endLine,
		EndColumn:   endColumn,
		Found:       true,
	}
}

// FindCommentPosition searches for the target text in the document body
// and returns the position where it was found
func FindCommentPosition(body string, comment *domain.PlanComment) *domain.CommentPosition {
	targetText := comment.TargetText
	contextBefore := comment.ContextBefore
	contextAfter := comment.ContextAfter

	// Try to find with full context first
	searchPattern := contextBefore + targetText + contextAfter
	idx := strings.Index(body, searchPattern)
	if idx >= 0 {
		startOffset := idx + len(contextBefore)
		return createPosition(body, startOffset, startOffset+len(targetText))
	}

	// Try to find just the target text with context before
	if contextBefore != "" {
		searchPattern = contextBefore + targetText
		idx = strings.Index(body, searchPattern)
		if idx >= 0 {
			startOffset := idx + len(contextBefore)
			return createPosition(body, startOffset, startOffset+len(targetText))
		}
	}

	// Try to find just the target text with context after
	if contextAfter != "" {
		searchPattern = targetText + contextAfter
		idx = strings.Index(body, searchPattern)
		if idx >= 0 {
			return createPosition(body, idx, idx+len(targetText))
		}
	}

	// Last resort: find just the target text (first occurrence)
	idx = strings.Index(body, targetText)
	if idx >= 0 {
		return createPosition(body, idx, idx+len(targetText))
	}

	// Not found
	return &domain.CommentPosition{
		Found: false,
	}
}

// computeBodyHash returns a SHA256 hash of the document body
func computeBodyHash(body string) string {
	hash := sha256.Sum256([]byte(body))
	return hex.EncodeToString(hash[:])
}

// Handlers

// List returns all comments for a plan document
func (h *PlanCommentHandler) List(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	vars := mux.Vars(r)
	planID := vars["id"]

	// First check if the plan document exists
	doc, err := h.repos.PlanDocument.FindByID(ctx, planID)
	if err != nil {
		http.Error(w, `{"error": "failed to fetch plan document"}`, http.StatusInternalServerError)
		return
	}
	if doc == nil {
		http.Error(w, `{"error": "plan document not found"}`, http.StatusNotFound)
		return
	}

	// Parse status filter
	var statusFilter *domain.PlanCommentStatus
	statusParam := r.URL.Query().Get("status")
	if statusParam != "" && statusParam != "all" {
		status := domain.PlanCommentStatus(statusParam)
		if status.IsValid() {
			statusFilter = &status
		}
	}

	comments, err := h.repos.PlanComment.FindByPlanDocumentID(ctx, planID, statusFilter)
	if err != nil {
		http.Error(w, `{"error": "failed to fetch comments"}`, http.StatusInternalServerError)
		return
	}

	// Build responses with user names and positions
	responses := make([]*PlanCommentResponse, 0, len(comments))
	for _, comment := range comments {
		// Get user name
		var userName string
		user, err := h.repos.User.FindByID(ctx, comment.UserID)
		if err == nil && user != nil {
			userName = user.GetDisplayName()
		}

		// Calculate position in current document
		position := FindCommentPosition(doc.Body, comment)

		responses = append(responses, h.commentToResponse(comment, userName, position))
	}

	response := PlanCommentListResponse{Comments: responses}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(response)
}

// Create creates a new comment on a plan document
func (h *PlanCommentHandler) Create(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	vars := mux.Vars(r)
	planID := vars["id"]

	// Get user ID from context
	userID := GetUserIDFromContext(ctx)
	if userID == "" {
		http.Error(w, `{"error": "authentication required"}`, http.StatusUnauthorized)
		return
	}

	// Check if the plan document exists
	doc, err := h.repos.PlanDocument.FindByID(ctx, planID)
	if err != nil {
		http.Error(w, `{"error": "failed to fetch plan document"}`, http.StatusInternalServerError)
		return
	}
	if doc == nil {
		http.Error(w, `{"error": "plan document not found"}`, http.StatusNotFound)
		return
	}

	var req CreatePlanCommentRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, `{"error": "invalid request body"}`, http.StatusBadRequest)
		return
	}

	if req.TargetText == "" {
		http.Error(w, `{"error": "target_text is required"}`, http.StatusBadRequest)
		return
	}
	if req.Content == "" {
		http.Error(w, `{"error": "content is required"}`, http.StatusBadRequest)
		return
	}

	// Create the comment
	comment := &domain.PlanComment{
		PlanDocumentID:   planID,
		UserID:           userID,
		TargetText:       req.TargetText,
		ContextBefore:    req.ContextBefore,
		ContextAfter:     req.ContextAfter,
		OriginalBodyHash: computeBodyHash(doc.Body),
		Content:          req.Content,
		Status:           domain.PlanCommentStatusActive,
	}

	if err := h.repos.PlanComment.Create(ctx, comment); err != nil {
		http.Error(w, `{"error": "failed to create comment"}`, http.StatusInternalServerError)
		return
	}

	// Get user name for response
	var userName string
	user, err := h.repos.User.FindByID(ctx, userID)
	if err == nil && user != nil {
		userName = user.GetDisplayName()
	}

	// Calculate position
	position := FindCommentPosition(doc.Body, comment)

	resp := h.commentToResponse(comment, userName, position)

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(resp)
}

// Update updates an existing comment (content only, owner only)
func (h *PlanCommentHandler) Update(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	vars := mux.Vars(r)
	planID := vars["id"]
	commentID := vars["commentId"]

	// Get user ID from context
	userID := GetUserIDFromContext(ctx)
	if userID == "" {
		http.Error(w, `{"error": "authentication required"}`, http.StatusUnauthorized)
		return
	}

	// Check if the plan document exists
	doc, err := h.repos.PlanDocument.FindByID(ctx, planID)
	if err != nil {
		http.Error(w, `{"error": "failed to fetch plan document"}`, http.StatusInternalServerError)
		return
	}
	if doc == nil {
		http.Error(w, `{"error": "plan document not found"}`, http.StatusNotFound)
		return
	}

	// Find the comment
	comment, err := h.repos.PlanComment.FindByID(ctx, commentID)
	if err != nil {
		http.Error(w, `{"error": "failed to fetch comment"}`, http.StatusInternalServerError)
		return
	}
	if comment == nil {
		http.Error(w, `{"error": "comment not found"}`, http.StatusNotFound)
		return
	}

	// Check ownership
	if comment.UserID != userID {
		http.Error(w, `{"error": "you can only edit your own comments"}`, http.StatusForbidden)
		return
	}

	// Check plan document ID matches
	if comment.PlanDocumentID != planID {
		http.Error(w, `{"error": "comment does not belong to this plan"}`, http.StatusBadRequest)
		return
	}

	var req UpdatePlanCommentRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, `{"error": "invalid request body"}`, http.StatusBadRequest)
		return
	}

	// Update content if provided
	if req.Content != nil {
		comment.Content = *req.Content
	}

	if err := h.repos.PlanComment.Update(ctx, comment); err != nil {
		http.Error(w, `{"error": "failed to update comment"}`, http.StatusInternalServerError)
		return
	}

	// Get user name for response
	var userName string
	user, err := h.repos.User.FindByID(ctx, userID)
	if err == nil && user != nil {
		userName = user.GetDisplayName()
	}

	// Calculate position
	position := FindCommentPosition(doc.Body, comment)

	resp := h.commentToResponse(comment, userName, position)

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(resp)
}

// Delete deletes a comment (owner only)
func (h *PlanCommentHandler) Delete(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	vars := mux.Vars(r)
	planID := vars["id"]
	commentID := vars["commentId"]

	// Get user ID from context
	userID := GetUserIDFromContext(ctx)
	if userID == "" {
		http.Error(w, `{"error": "authentication required"}`, http.StatusUnauthorized)
		return
	}

	// Find the comment
	comment, err := h.repos.PlanComment.FindByID(ctx, commentID)
	if err != nil {
		http.Error(w, `{"error": "failed to fetch comment"}`, http.StatusInternalServerError)
		return
	}
	if comment == nil {
		http.Error(w, `{"error": "comment not found"}`, http.StatusNotFound)
		return
	}

	// Check ownership
	if comment.UserID != userID {
		http.Error(w, `{"error": "you can only delete your own comments"}`, http.StatusForbidden)
		return
	}

	// Check plan document ID matches
	if comment.PlanDocumentID != planID {
		http.Error(w, `{"error": "comment does not belong to this plan"}`, http.StatusBadRequest)
		return
	}

	if err := h.repos.PlanComment.Delete(ctx, commentID); err != nil {
		http.Error(w, `{"error": "failed to delete comment"}`, http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusNoContent)
}

// Resolve marks a comment as resolved
func (h *PlanCommentHandler) Resolve(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	vars := mux.Vars(r)
	planID := vars["id"]
	commentID := vars["commentId"]

	// Get user ID from context
	userID := GetUserIDFromContext(ctx)
	if userID == "" {
		http.Error(w, `{"error": "authentication required"}`, http.StatusUnauthorized)
		return
	}

	// Check if the plan document exists
	doc, err := h.repos.PlanDocument.FindByID(ctx, planID)
	if err != nil {
		http.Error(w, `{"error": "failed to fetch plan document"}`, http.StatusInternalServerError)
		return
	}
	if doc == nil {
		http.Error(w, `{"error": "plan document not found"}`, http.StatusNotFound)
		return
	}

	// Find the comment
	comment, err := h.repos.PlanComment.FindByID(ctx, commentID)
	if err != nil {
		http.Error(w, `{"error": "failed to fetch comment"}`, http.StatusInternalServerError)
		return
	}
	if comment == nil {
		http.Error(w, `{"error": "comment not found"}`, http.StatusNotFound)
		return
	}

	// Check plan document ID matches
	if comment.PlanDocumentID != planID {
		http.Error(w, `{"error": "comment does not belong to this plan"}`, http.StatusBadRequest)
		return
	}

	// Update status to resolved
	comment.Status = domain.PlanCommentStatusResolved

	if err := h.repos.PlanComment.Update(ctx, comment); err != nil {
		http.Error(w, `{"error": "failed to resolve comment"}`, http.StatusInternalServerError)
		return
	}

	// Get user name for response
	var userName string
	user, err := h.repos.User.FindByID(ctx, comment.UserID)
	if err == nil && user != nil {
		userName = user.GetDisplayName()
	}

	// Calculate position
	position := FindCommentPosition(doc.Body, comment)

	resp := h.commentToResponse(comment, userName, position)

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(resp)
}
