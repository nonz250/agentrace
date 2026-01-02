package api

import (
	"encoding/json"
	"net/http"
	"strconv"

	"github.com/gorilla/mux"
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

type ProjectResponse struct {
	ID                     string `json:"id"`
	CanonicalGitRepository string `json:"canonical_git_repository"`
	CreatedAt              string `json:"created_at"`
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

// Get returns a single project by ID
func (h *ProjectHandler) Get(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	vars := mux.Vars(r)
	id := vars["id"]

	project, err := h.repos.Project.FindByID(ctx, id)
	if err != nil {
		http.Error(w, `{"error": "failed to fetch project"}`, http.StatusInternalServerError)
		return
	}
	if project == nil {
		http.Error(w, `{"error": "project not found"}`, http.StatusNotFound)
		return
	}

	response := ProjectResponse{
		ID:                     project.ID,
		CanonicalGitRepository: project.CanonicalGitRepository,
		CreatedAt:              project.CreatedAt.Format("2006-01-02T15:04:05Z07:00"),
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(response)
}
