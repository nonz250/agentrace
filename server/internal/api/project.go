package api

import (
	"encoding/json"
	"net/http"
	"strconv"

	"github.com/satetsu888/agentrace/server/internal/repository"
)

type ProjectHandler struct {
	repos *repository.Repositories
}

func NewProjectHandler(repos *repository.Repositories) *ProjectHandler {
	return &ProjectHandler{repos: repos}
}

// Response types

type ProjectListItemResponse struct {
	ID                     string `json:"id"`
	CanonicalGitRepository string `json:"canonical_git_repository"`
	CreatedAt              string `json:"created_at"`
}

type ProjectListResponse struct {
	Projects []*ProjectListItemResponse `json:"projects"`
}

// List returns all projects
func (h *ProjectHandler) List(w http.ResponseWriter, r *http.Request) {
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

	projects, err := h.repos.Project.FindAll(ctx, limit, offset)
	if err != nil {
		http.Error(w, `{"error": "failed to fetch projects"}`, http.StatusInternalServerError)
		return
	}

	responses := make([]*ProjectListItemResponse, len(projects))
	for i, p := range projects {
		responses[i] = &ProjectListItemResponse{
			ID:                     p.ID,
			CanonicalGitRepository: p.CanonicalGitRepository,
			CreatedAt:              p.CreatedAt.Format("2006-01-02T15:04:05Z07:00"),
		}
	}

	response := ProjectListResponse{Projects: responses}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(response)
}
