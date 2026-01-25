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

type PlanCommentThreadHandler struct {
	repos *repository.Repositories
}

func NewPlanCommentThreadHandler(repos *repository.Repositories) *PlanCommentThreadHandler {
	return &PlanCommentThreadHandler{repos: repos}
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

type MessageResponse struct {
	ID        string `json:"id"`
	ThreadID  string `json:"thread_id"`
	UserID    string `json:"user_id"`
	UserName  string `json:"user_name"`
	Content   string `json:"content"`
	CreatedAt string `json:"created_at"`
	UpdatedAt string `json:"updated_at"`
}

type ThreadResponse struct {
	ID             string                   `json:"id"`
	PlanDocumentID string                   `json:"plan_document_id"`
	TargetText     string                   `json:"target_text"`
	Status         string                   `json:"status"`
	Position       *CommentPositionResponse `json:"position"`
	Messages       []*MessageResponse       `json:"messages"`
	CreatedAt      string                   `json:"created_at"`
	UpdatedAt      string                   `json:"updated_at"`
}

type ThreadListResponse struct {
	Threads []*ThreadResponse `json:"threads"`
}

// Request types

type CreateThreadRequest struct {
	TargetText    string `json:"target_text"`
	ContextBefore string `json:"context_before"`
	ContextAfter  string `json:"context_after"`
	Content       string `json:"content"` // First message content
}

type CreateMessageRequest struct {
	Content string `json:"content"`
}

type UpdateMessageRequest struct {
	Content *string `json:"content"`
}

// Helper functions

func (h *PlanCommentThreadHandler) messageToResponse(msg *domain.PlanCommentMessage, userName string) *MessageResponse {
	return &MessageResponse{
		ID:        msg.ID,
		ThreadID:  msg.ThreadID,
		UserID:    msg.UserID,
		UserName:  userName,
		Content:   msg.Content,
		CreatedAt: msg.CreatedAt.Format("2006-01-02T15:04:05Z07:00"),
		UpdatedAt: msg.UpdatedAt.Format("2006-01-02T15:04:05Z07:00"),
	}
}

func (h *PlanCommentThreadHandler) threadToResponse(thread *domain.PlanCommentThread, messages []*MessageResponse, position *domain.CommentPosition) *ThreadResponse {
	resp := &ThreadResponse{
		ID:             thread.ID,
		PlanDocumentID: thread.PlanDocumentID,
		TargetText:     thread.TargetText,
		Status:         string(thread.Status),
		Messages:       messages,
		CreatedAt:      thread.CreatedAt.Format("2006-01-02T15:04:05Z07:00"),
		UpdatedAt:      thread.UpdatedAt.Format("2006-01-02T15:04:05Z07:00"),
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

// stripInlineMarkdown removes inline markdown formatting from text
// and returns the stripped text along with a mapping from stripped positions to original positions
func stripInlineMarkdown(text string) (stripped string, posMap []int) {
	// posMap[i] = original position of stripped[i]
	posMap = make([]int, 0, len(text))
	result := strings.Builder{}

	runes := []rune(text)
	i := 0
	for i < len(runes) {
		ch := runes[i]

		// Handle inline code: `code`
		if ch == '`' {
			// Find closing backtick
			j := i + 1
			for j < len(runes) && runes[j] != '`' {
				j++
			}
			if j < len(runes) {
				// Found closing backtick, extract content
				for k := i + 1; k < j; k++ {
					result.WriteRune(runes[k])
					posMap = append(posMap, k)
				}
				i = j + 1
				continue
			}
		}

		// Handle links: [text](url) or [text](url "title")
		if ch == '[' {
			// Find closing bracket
			j := i + 1
			bracketDepth := 1
			for j < len(runes) && bracketDepth > 0 {
				if runes[j] == '[' {
					bracketDepth++
				} else if runes[j] == ']' {
					bracketDepth--
				}
				j++
			}
			// j now points to character after ']'
			if j < len(runes) && runes[j] == '(' {
				// Find closing parenthesis
				k := j + 1
				parenDepth := 1
				for k < len(runes) && parenDepth > 0 {
					if runes[k] == '(' {
						parenDepth++
					} else if runes[k] == ')' {
						parenDepth--
					}
					k++
				}
				// k now points to character after ')'
				if parenDepth == 0 {
					// Valid link found, extract link text (between [ and ])
					for l := i + 1; l < j-1; l++ {
						result.WriteRune(runes[l])
						posMap = append(posMap, l)
					}
					i = k
					continue
				}
			}
		}

		// Handle bold/italic: ** or * or __ or _
		// For simplicity, just skip the markers
		if ch == '*' || ch == '_' {
			// Check for double marker
			if i+1 < len(runes) && runes[i+1] == ch {
				i += 2
				continue
			}
			i++
			continue
		}

		// Handle strikethrough: ~~
		if ch == '~' && i+1 < len(runes) && runes[i+1] == '~' {
			i += 2
			continue
		}

		// Regular character
		result.WriteRune(ch)
		posMap = append(posMap, i)
		i++
	}

	return result.String(), posMap
}

// runeIndex finds the first occurrence of needle in haystack (as rune slices)
// and returns the rune index, or -1 if not found
func runeIndex(haystack, needle []rune) int {
	if len(needle) == 0 {
		return 0
	}
	if len(needle) > len(haystack) {
		return -1
	}
	for i := 0; i <= len(haystack)-len(needle); i++ {
		match := true
		for j := 0; j < len(needle); j++ {
			if haystack[i+j] != needle[j] {
				match = false
				break
			}
		}
		if match {
			return i
		}
	}
	return -1
}

// findInStrippedText searches for targetText in the stripped version of body
// and returns the original byte offset if found
func findInStrippedText(body, targetText, contextBefore, contextAfter string) (found bool, startOffset int, endOffset int) {
	strippedBody, posMap := stripInlineMarkdown(body)
	strippedTarget := stripInlineMarkdownSimple(targetText)
	strippedBefore := stripInlineMarkdownSimple(contextBefore)
	strippedAfter := stripInlineMarkdownSimple(contextAfter)

	// Convert to rune slices for proper Unicode handling
	strippedBodyRunes := []rune(strippedBody)
	strippedTargetRunes := []rune(strippedTarget)
	strippedBeforeRunes := []rune(strippedBefore)
	strippedAfterRunes := []rune(strippedAfter)

	// Try to find with full context first
	searchPatternRunes := append(append(strippedBeforeRunes, strippedTargetRunes...), strippedAfterRunes...)
	idx := runeIndex(strippedBodyRunes, searchPatternRunes)
	if idx >= 0 {
		strippedStart := idx + len(strippedBeforeRunes)
		strippedEnd := strippedStart + len(strippedTargetRunes)

		if strippedStart < len(posMap) && strippedEnd <= len(posMap) {
			// Convert stripped positions to original rune positions
			origStartRune := posMap[strippedStart]
			origEndRune := posMap[strippedEnd-1] + 1

			// Convert rune positions to byte offsets
			runes := []rune(body)
			byteStart := len(string(runes[:origStartRune]))
			byteEnd := len(string(runes[:origEndRune]))

			return true, byteStart, byteEnd
		}
	}

	// Try with context before only
	if len(strippedBeforeRunes) > 0 {
		searchPatternRunes = append(strippedBeforeRunes, strippedTargetRunes...)
		idx = runeIndex(strippedBodyRunes, searchPatternRunes)
		if idx >= 0 {
			strippedStart := idx + len(strippedBeforeRunes)
			strippedEnd := strippedStart + len(strippedTargetRunes)

			if strippedStart < len(posMap) && strippedEnd <= len(posMap) {
				origStartRune := posMap[strippedStart]
				origEndRune := posMap[strippedEnd-1] + 1

				runes := []rune(body)
				byteStart := len(string(runes[:origStartRune]))
				byteEnd := len(string(runes[:origEndRune]))

				return true, byteStart, byteEnd
			}
		}
	}

	// Try with context after only
	if len(strippedAfterRunes) > 0 {
		searchPatternRunes = append(strippedTargetRunes, strippedAfterRunes...)
		idx = runeIndex(strippedBodyRunes, searchPatternRunes)
		if idx >= 0 {
			strippedStart := idx
			strippedEnd := strippedStart + len(strippedTargetRunes)

			if strippedStart < len(posMap) && strippedEnd <= len(posMap) {
				origStartRune := posMap[strippedStart]
				origEndRune := posMap[strippedEnd-1] + 1

				runes := []rune(body)
				byteStart := len(string(runes[:origStartRune]))
				byteEnd := len(string(runes[:origEndRune]))

				return true, byteStart, byteEnd
			}
		}
	}

	// Try target only (no context)
	if len(strippedBeforeRunes) == 0 && len(strippedAfterRunes) == 0 {
		idx = runeIndex(strippedBodyRunes, strippedTargetRunes)
		if idx >= 0 {
			strippedStart := idx
			strippedEnd := strippedStart + len(strippedTargetRunes)

			if strippedStart < len(posMap) && strippedEnd <= len(posMap) {
				origStartRune := posMap[strippedStart]
				origEndRune := posMap[strippedEnd-1] + 1

				runes := []rune(body)
				byteStart := len(string(runes[:origStartRune]))
				byteEnd := len(string(runes[:origEndRune]))

				return true, byteStart, byteEnd
			}
		}
	}

	return false, 0, 0
}

// stripInlineMarkdownSimple is a simplified version that just returns the stripped text
func stripInlineMarkdownSimple(text string) string {
	stripped, _ := stripInlineMarkdown(text)
	return stripped
}

// FindThreadPosition searches for the target text in the document body
// and returns the position where it was found
func FindThreadPosition(body string, thread *domain.PlanCommentThread) *domain.CommentPosition {
	targetText := thread.TargetText
	contextBefore := thread.ContextBefore
	contextAfter := thread.ContextAfter

	// Try to find with full context first (exact match)
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

	// Try matching with inline markdown stripped (for cross-inline-element selections)
	found, startOffset, endOffset := findInStrippedText(body, targetText, contextBefore, contextAfter)
	if found {
		return createPosition(body, startOffset, endOffset)
	}

	// If we had context but couldn't match with it, don't fall back to first occurrence
	// This prevents comments from jumping to unrelated identical text
	if contextBefore != "" || contextAfter != "" {
		return &domain.CommentPosition{
			Found: false,
		}
	}

	// Only fall back to first occurrence if no context was provided at all
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

// List returns all threads with their messages for a plan document
func (h *PlanCommentThreadHandler) List(w http.ResponseWriter, r *http.Request) {
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
	var statusFilter *domain.PlanCommentThreadStatus
	statusParam := r.URL.Query().Get("status")
	if statusParam != "" && statusParam != "all" {
		status := domain.PlanCommentThreadStatus(statusParam)
		if status.IsValid() {
			statusFilter = &status
		}
	}

	threads, err := h.repos.PlanCommentThread.FindByPlanDocumentID(ctx, planID, statusFilter)
	if err != nil {
		http.Error(w, `{"error": "failed to fetch threads"}`, http.StatusInternalServerError)
		return
	}

	// Build responses with messages and positions
	responses := make([]*ThreadResponse, 0, len(threads))
	for _, thread := range threads {
		// Get messages for this thread
		messages, err := h.repos.PlanCommentMessage.FindByThreadID(ctx, thread.ID)
		if err != nil {
			http.Error(w, `{"error": "failed to fetch messages"}`, http.StatusInternalServerError)
			return
		}

		// Convert messages to responses with user names
		msgResponses := make([]*MessageResponse, 0, len(messages))
		for _, msg := range messages {
			var userName string
			user, err := h.repos.User.FindByID(ctx, msg.UserID)
			if err == nil && user != nil {
				userName = user.GetDisplayName()
			}
			msgResponses = append(msgResponses, h.messageToResponse(msg, userName))
		}

		// Calculate position in current document
		position := FindThreadPosition(doc.Body, thread)

		responses = append(responses, h.threadToResponse(thread, msgResponses, position))
	}

	response := ThreadListResponse{Threads: responses}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(response)
}

// Create creates a new thread with its first message
func (h *PlanCommentThreadHandler) Create(w http.ResponseWriter, r *http.Request) {
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

	var req CreateThreadRequest
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

	// Create the thread
	thread := &domain.PlanCommentThread{
		PlanDocumentID:   planID,
		TargetText:       req.TargetText,
		ContextBefore:    req.ContextBefore,
		ContextAfter:     req.ContextAfter,
		OriginalBodyHash: computeBodyHash(doc.Body),
		Status:           domain.PlanCommentThreadStatusActive,
	}

	if err := h.repos.PlanCommentThread.Create(ctx, thread); err != nil {
		http.Error(w, `{"error": "failed to create thread"}`, http.StatusInternalServerError)
		return
	}

	// Create the first message
	message := &domain.PlanCommentMessage{
		ThreadID: thread.ID,
		UserID:   userID,
		Content:  req.Content,
	}

	if err := h.repos.PlanCommentMessage.Create(ctx, message); err != nil {
		// Rollback thread creation
		h.repos.PlanCommentThread.Delete(ctx, thread.ID)
		http.Error(w, `{"error": "failed to create message"}`, http.StatusInternalServerError)
		return
	}

	// Get user name for response
	var userName string
	user, err := h.repos.User.FindByID(ctx, userID)
	if err == nil && user != nil {
		userName = user.GetDisplayName()
	}

	// Calculate position
	position := FindThreadPosition(doc.Body, thread)

	msgResponses := []*MessageResponse{h.messageToResponse(message, userName)}
	resp := h.threadToResponse(thread, msgResponses, position)

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(resp)
}

// Resolve marks a thread as resolved
func (h *PlanCommentThreadHandler) Resolve(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	vars := mux.Vars(r)
	planID := vars["id"]
	threadID := vars["threadId"]

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

	// Find the thread
	thread, err := h.repos.PlanCommentThread.FindByID(ctx, threadID)
	if err != nil {
		http.Error(w, `{"error": "failed to fetch thread"}`, http.StatusInternalServerError)
		return
	}
	if thread == nil {
		http.Error(w, `{"error": "thread not found"}`, http.StatusNotFound)
		return
	}

	// Check plan document ID matches
	if thread.PlanDocumentID != planID {
		http.Error(w, `{"error": "thread does not belong to this plan"}`, http.StatusBadRequest)
		return
	}

	// Update status to resolved
	thread.Status = domain.PlanCommentThreadStatusResolved

	if err := h.repos.PlanCommentThread.Update(ctx, thread); err != nil {
		http.Error(w, `{"error": "failed to resolve thread"}`, http.StatusInternalServerError)
		return
	}

	// Get messages for response
	messages, err := h.repos.PlanCommentMessage.FindByThreadID(ctx, thread.ID)
	if err != nil {
		http.Error(w, `{"error": "failed to fetch messages"}`, http.StatusInternalServerError)
		return
	}

	msgResponses := make([]*MessageResponse, 0, len(messages))
	for _, msg := range messages {
		var userName string
		user, err := h.repos.User.FindByID(ctx, msg.UserID)
		if err == nil && user != nil {
			userName = user.GetDisplayName()
		}
		msgResponses = append(msgResponses, h.messageToResponse(msg, userName))
	}

	// Calculate position
	position := FindThreadPosition(doc.Body, thread)

	resp := h.threadToResponse(thread, msgResponses, position)

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(resp)
}

// Delete deletes a thread and all its messages
func (h *PlanCommentThreadHandler) Delete(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	vars := mux.Vars(r)
	planID := vars["id"]
	threadID := vars["threadId"]

	// Get user ID from context
	userID := GetUserIDFromContext(ctx)
	if userID == "" {
		http.Error(w, `{"error": "authentication required"}`, http.StatusUnauthorized)
		return
	}

	// Find the thread
	thread, err := h.repos.PlanCommentThread.FindByID(ctx, threadID)
	if err != nil {
		http.Error(w, `{"error": "failed to fetch thread"}`, http.StatusInternalServerError)
		return
	}
	if thread == nil {
		http.Error(w, `{"error": "thread not found"}`, http.StatusNotFound)
		return
	}

	// Check plan document ID matches
	if thread.PlanDocumentID != planID {
		http.Error(w, `{"error": "thread does not belong to this plan"}`, http.StatusBadRequest)
		return
	}

	// Get the first message to check ownership
	messages, err := h.repos.PlanCommentMessage.FindByThreadID(ctx, threadID)
	if err != nil {
		http.Error(w, `{"error": "failed to fetch messages"}`, http.StatusInternalServerError)
		return
	}

	// Check ownership (only the creator of the first message can delete the thread)
	if len(messages) > 0 && messages[0].UserID != userID {
		http.Error(w, `{"error": "you can only delete threads you created"}`, http.StatusForbidden)
		return
	}

	// Delete all messages first (due to FK constraint), then the thread
	if err := h.repos.PlanCommentMessage.DeleteByThreadID(ctx, threadID); err != nil {
		http.Error(w, `{"error": "failed to delete messages"}`, http.StatusInternalServerError)
		return
	}

	if err := h.repos.PlanCommentThread.Delete(ctx, threadID); err != nil {
		http.Error(w, `{"error": "failed to delete thread"}`, http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusNoContent)
}

// AddMessage adds a reply to a thread
func (h *PlanCommentThreadHandler) AddMessage(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	vars := mux.Vars(r)
	planID := vars["id"]
	threadID := vars["threadId"]

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

	// Find the thread
	thread, err := h.repos.PlanCommentThread.FindByID(ctx, threadID)
	if err != nil {
		http.Error(w, `{"error": "failed to fetch thread"}`, http.StatusInternalServerError)
		return
	}
	if thread == nil {
		http.Error(w, `{"error": "thread not found"}`, http.StatusNotFound)
		return
	}

	// Check plan document ID matches
	if thread.PlanDocumentID != planID {
		http.Error(w, `{"error": "thread does not belong to this plan"}`, http.StatusBadRequest)
		return
	}

	var req CreateMessageRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, `{"error": "invalid request body"}`, http.StatusBadRequest)
		return
	}

	if req.Content == "" {
		http.Error(w, `{"error": "content is required"}`, http.StatusBadRequest)
		return
	}

	// Create the message
	message := &domain.PlanCommentMessage{
		ThreadID: threadID,
		UserID:   userID,
		Content:  req.Content,
	}

	if err := h.repos.PlanCommentMessage.Create(ctx, message); err != nil {
		http.Error(w, `{"error": "failed to create message"}`, http.StatusInternalServerError)
		return
	}

	// Get user name for response
	var userName string
	user, err := h.repos.User.FindByID(ctx, userID)
	if err == nil && user != nil {
		userName = user.GetDisplayName()
	}

	resp := h.messageToResponse(message, userName)

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(resp)
}

// UpdateMessage updates a message (owner only)
func (h *PlanCommentThreadHandler) UpdateMessage(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	vars := mux.Vars(r)
	planID := vars["id"]
	threadID := vars["threadId"]
	messageID := vars["messageId"]

	// Get user ID from context
	userID := GetUserIDFromContext(ctx)
	if userID == "" {
		http.Error(w, `{"error": "authentication required"}`, http.StatusUnauthorized)
		return
	}

	// Find the thread
	thread, err := h.repos.PlanCommentThread.FindByID(ctx, threadID)
	if err != nil {
		http.Error(w, `{"error": "failed to fetch thread"}`, http.StatusInternalServerError)
		return
	}
	if thread == nil {
		http.Error(w, `{"error": "thread not found"}`, http.StatusNotFound)
		return
	}

	// Check plan document ID matches
	if thread.PlanDocumentID != planID {
		http.Error(w, `{"error": "thread does not belong to this plan"}`, http.StatusBadRequest)
		return
	}

	// Find the message
	message, err := h.repos.PlanCommentMessage.FindByID(ctx, messageID)
	if err != nil {
		http.Error(w, `{"error": "failed to fetch message"}`, http.StatusInternalServerError)
		return
	}
	if message == nil {
		http.Error(w, `{"error": "message not found"}`, http.StatusNotFound)
		return
	}

	// Check thread ID matches
	if message.ThreadID != threadID {
		http.Error(w, `{"error": "message does not belong to this thread"}`, http.StatusBadRequest)
		return
	}

	// Check ownership
	if message.UserID != userID {
		http.Error(w, `{"error": "you can only edit your own messages"}`, http.StatusForbidden)
		return
	}

	var req UpdateMessageRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, `{"error": "invalid request body"}`, http.StatusBadRequest)
		return
	}

	// Update content if provided
	if req.Content != nil {
		message.Content = *req.Content
	}

	if err := h.repos.PlanCommentMessage.Update(ctx, message); err != nil {
		http.Error(w, `{"error": "failed to update message"}`, http.StatusInternalServerError)
		return
	}

	// Get user name for response
	var userName string
	user, err := h.repos.User.FindByID(ctx, userID)
	if err == nil && user != nil {
		userName = user.GetDisplayName()
	}

	resp := h.messageToResponse(message, userName)

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(resp)
}

// DeleteMessage deletes a message (owner only)
func (h *PlanCommentThreadHandler) DeleteMessage(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	vars := mux.Vars(r)
	planID := vars["id"]
	threadID := vars["threadId"]
	messageID := vars["messageId"]

	// Get user ID from context
	userID := GetUserIDFromContext(ctx)
	if userID == "" {
		http.Error(w, `{"error": "authentication required"}`, http.StatusUnauthorized)
		return
	}

	// Find the thread
	thread, err := h.repos.PlanCommentThread.FindByID(ctx, threadID)
	if err != nil {
		http.Error(w, `{"error": "failed to fetch thread"}`, http.StatusInternalServerError)
		return
	}
	if thread == nil {
		http.Error(w, `{"error": "thread not found"}`, http.StatusNotFound)
		return
	}

	// Check plan document ID matches
	if thread.PlanDocumentID != planID {
		http.Error(w, `{"error": "thread does not belong to this plan"}`, http.StatusBadRequest)
		return
	}

	// Find the message
	message, err := h.repos.PlanCommentMessage.FindByID(ctx, messageID)
	if err != nil {
		http.Error(w, `{"error": "failed to fetch message"}`, http.StatusInternalServerError)
		return
	}
	if message == nil {
		http.Error(w, `{"error": "message not found"}`, http.StatusNotFound)
		return
	}

	// Check thread ID matches
	if message.ThreadID != threadID {
		http.Error(w, `{"error": "message does not belong to this thread"}`, http.StatusBadRequest)
		return
	}

	// Check ownership
	if message.UserID != userID {
		http.Error(w, `{"error": "you can only delete your own messages"}`, http.StatusForbidden)
		return
	}

	// Get all messages in the thread to check if this is the last one
	messages, err := h.repos.PlanCommentMessage.FindByThreadID(ctx, threadID)
	if err != nil {
		http.Error(w, `{"error": "failed to fetch messages"}`, http.StatusInternalServerError)
		return
	}

	// If this is the only message, delete the entire thread
	if len(messages) == 1 {
		if err := h.repos.PlanCommentMessage.Delete(ctx, messageID); err != nil {
			http.Error(w, `{"error": "failed to delete message"}`, http.StatusInternalServerError)
			return
		}
		if err := h.repos.PlanCommentThread.Delete(ctx, threadID); err != nil {
			http.Error(w, `{"error": "failed to delete thread"}`, http.StatusInternalServerError)
			return
		}
	} else {
		// Just delete the message
		if err := h.repos.PlanCommentMessage.Delete(ctx, messageID); err != nil {
			http.Error(w, `{"error": "failed to delete message"}`, http.StatusInternalServerError)
			return
		}
	}

	w.WriteHeader(http.StatusNoContent)
}
