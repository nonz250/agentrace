package api

import (
	"encoding/json"
	"net/http"
	"strconv"

	"github.com/gorilla/mux"
	"github.com/satetsu888/agentrace/server/internal/domain"
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
	Projects   []*ProjectListItemResponse `json:"projects"`
	NextCursor string                     `json:"next_cursor,omitempty"`
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
	cursor := r.URL.Query().Get("cursor")
	if limitStr := r.URL.Query().Get("limit"); limitStr != "" {
		if l, err := strconv.Atoi(limitStr); err == nil && l > 0 && l <= 100 {
			limit = l
		}
	}

	projects, nextCursor, err := h.repos.Project.FindAll(ctx, limit, cursor)
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

	response := ProjectListResponse{
		Projects:   responses,
		NextCursor: nextCursor,
	}

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

// Delete deletes a project if it has no related sessions or plan documents
func (h *ProjectHandler) Delete(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	vars := mux.Vars(r)
	id := vars["id"]

	// Prevent deletion of default project
	if id == domain.DefaultProjectID {
		http.Error(w, `{"error": "cannot delete default project"}`, http.StatusBadRequest)
		return
	}

	// Check if project exists
	project, err := h.repos.Project.FindByID(ctx, id)
	if err != nil {
		http.Error(w, `{"error": "failed to fetch project"}`, http.StatusInternalServerError)
		return
	}
	if project == nil {
		http.Error(w, `{"error": "project not found"}`, http.StatusNotFound)
		return
	}

	// Check if project has related data (sessions or plan documents)
	hasData, err := h.repos.Project.HasRelatedData(ctx, id)
	if err != nil {
		http.Error(w, `{"error": "failed to check related data"}`, http.StatusInternalServerError)
		return
	}
	if hasData {
		http.Error(w, `{"error": "cannot delete project with sessions or plans"}`, http.StatusConflict)
		return
	}

	// Delete the project
	if err := h.repos.Project.Delete(ctx, id); err != nil {
		http.Error(w, `{"error": "failed to delete project"}`, http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusNoContent)
}
