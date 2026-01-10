package api

import (
	"encoding/json"
	"net/http"

	"github.com/satetsu888/agentrace/server/internal/domain"
	"github.com/satetsu888/agentrace/server/internal/repository"
)

type UserFavoriteHandler struct {
	repos *repository.Repositories
}

func NewUserFavoriteHandler(repos *repository.Repositories) *UserFavoriteHandler {
	return &UserFavoriteHandler{repos: repos}
}

type createUserFavoriteRequest struct {
	TargetType string `json:"target_type"`
	TargetID   string `json:"target_id"`
}

type userFavoriteResponse struct {
	ID         string `json:"id"`
	UserID     string `json:"user_id"`
	TargetType string `json:"target_type"`
	TargetID   string `json:"target_id"`
	CreatedAt  string `json:"created_at"`
}

func (h *UserFavoriteHandler) List(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	userID := GetUserIDFromContext(ctx)
	if userID == "" {
		http.Error(w, `{"error": "unauthorized"}`, http.StatusUnauthorized)
		return
	}

	targetType := r.URL.Query().Get("target_type")

	var favorites []*domain.UserFavorite
	var err error

	if targetType != "" {
		tt := domain.UserFavoriteTargetType(targetType)
		if !tt.IsValid() {
			http.Error(w, `{"error": "invalid target_type"}`, http.StatusBadRequest)
			return
		}
		favorites, err = h.repos.UserFavorite.FindByUserAndTargetType(ctx, userID, tt)
	} else {
		favorites, err = h.repos.UserFavorite.FindByUserID(ctx, userID)
	}

	if err != nil {
		http.Error(w, `{"error": "failed to get favorites"}`, http.StatusInternalServerError)
		return
	}

	response := struct {
		Favorites []userFavoriteResponse `json:"favorites"`
	}{
		Favorites: make([]userFavoriteResponse, 0, len(favorites)),
	}

	for _, f := range favorites {
		response.Favorites = append(response.Favorites, userFavoriteResponse{
			ID:         f.ID,
			UserID:     f.UserID,
			TargetType: string(f.TargetType),
			TargetID:   f.TargetID,
			CreatedAt:  f.CreatedAt.Format("2006-01-02T15:04:05Z07:00"),
		})
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(response)
}

func (h *UserFavoriteHandler) Create(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	userID := GetUserIDFromContext(ctx)
	if userID == "" {
		http.Error(w, `{"error": "unauthorized"}`, http.StatusUnauthorized)
		return
	}

	var req createUserFavoriteRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, `{"error": "invalid request body"}`, http.StatusBadRequest)
		return
	}

	targetType := domain.UserFavoriteTargetType(req.TargetType)
	if !targetType.IsValid() {
		http.Error(w, `{"error": "invalid target_type"}`, http.StatusBadRequest)
		return
	}

	if req.TargetID == "" {
		http.Error(w, `{"error": "target_id is required"}`, http.StatusBadRequest)
		return
	}

	// Check if already favorited
	existing, err := h.repos.UserFavorite.FindByUserAndTarget(ctx, userID, targetType, req.TargetID)
	if err != nil {
		http.Error(w, `{"error": "failed to check existing favorite"}`, http.StatusInternalServerError)
		return
	}
	if existing != nil {
		// Already favorited, return the existing one
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(userFavoriteResponse{
			ID:         existing.ID,
			UserID:     existing.UserID,
			TargetType: string(existing.TargetType),
			TargetID:   existing.TargetID,
			CreatedAt:  existing.CreatedAt.Format("2006-01-02T15:04:05Z07:00"),
		})
		return
	}

	favorite := &domain.UserFavorite{
		UserID:     userID,
		TargetType: targetType,
		TargetID:   req.TargetID,
	}

	if err := h.repos.UserFavorite.Create(ctx, favorite); err != nil {
		http.Error(w, `{"error": "failed to create favorite"}`, http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(userFavoriteResponse{
		ID:         favorite.ID,
		UserID:     favorite.UserID,
		TargetType: string(favorite.TargetType),
		TargetID:   favorite.TargetID,
		CreatedAt:  favorite.CreatedAt.Format("2006-01-02T15:04:05Z07:00"),
	})
}

func (h *UserFavoriteHandler) Delete(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	userID := GetUserIDFromContext(ctx)
	if userID == "" {
		http.Error(w, `{"error": "unauthorized"}`, http.StatusUnauthorized)
		return
	}

	targetType := r.URL.Query().Get("target_type")
	targetID := r.URL.Query().Get("target_id")

	if targetType == "" || targetID == "" {
		http.Error(w, `{"error": "target_type and target_id are required"}`, http.StatusBadRequest)
		return
	}

	tt := domain.UserFavoriteTargetType(targetType)
	if !tt.IsValid() {
		http.Error(w, `{"error": "invalid target_type"}`, http.StatusBadRequest)
		return
	}

	if err := h.repos.UserFavorite.DeleteByUserAndTarget(ctx, userID, tt, targetID); err != nil {
		http.Error(w, `{"error": "failed to delete favorite"}`, http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusNoContent)
}
