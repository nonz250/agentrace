package memory

import (
	"context"
	"sync"
	"time"

	"github.com/google/uuid"
	"github.com/satetsu888/agentrace/server/internal/domain"
)

type UserFavoriteRepository struct {
	mu        sync.RWMutex
	favorites map[string]*domain.UserFavorite
}

func NewUserFavoriteRepository() *UserFavoriteRepository {
	return &UserFavoriteRepository{
		favorites: make(map[string]*domain.UserFavorite),
	}
}

func (r *UserFavoriteRepository) Create(ctx context.Context, favorite *domain.UserFavorite) error {
	r.mu.Lock()
	defer r.mu.Unlock()

	if favorite.ID == "" {
		favorite.ID = uuid.New().String()
	}
	if favorite.CreatedAt.IsZero() {
		favorite.CreatedAt = time.Now()
	}

	r.favorites[favorite.ID] = favorite
	return nil
}

func (r *UserFavoriteRepository) Delete(ctx context.Context, id string) error {
	r.mu.Lock()
	defer r.mu.Unlock()

	delete(r.favorites, id)
	return nil
}

func (r *UserFavoriteRepository) DeleteByUserAndTarget(ctx context.Context, userID string, targetType domain.UserFavoriteTargetType, targetID string) error {
	r.mu.Lock()
	defer r.mu.Unlock()

	for id, f := range r.favorites {
		if f.UserID == userID && f.TargetType == targetType && f.TargetID == targetID {
			delete(r.favorites, id)
			return nil
		}
	}
	return nil
}

func (r *UserFavoriteRepository) FindByUserID(ctx context.Context, userID string) ([]*domain.UserFavorite, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()

	favorites := make([]*domain.UserFavorite, 0)
	for _, f := range r.favorites {
		if f.UserID == userID {
			favorites = append(favorites, f)
		}
	}
	return favorites, nil
}

func (r *UserFavoriteRepository) FindByUserAndTargetType(ctx context.Context, userID string, targetType domain.UserFavoriteTargetType) ([]*domain.UserFavorite, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()

	favorites := make([]*domain.UserFavorite, 0)
	for _, f := range r.favorites {
		if f.UserID == userID && f.TargetType == targetType {
			favorites = append(favorites, f)
		}
	}
	return favorites, nil
}

func (r *UserFavoriteRepository) FindByUserAndTarget(ctx context.Context, userID string, targetType domain.UserFavoriteTargetType, targetID string) (*domain.UserFavorite, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()

	for _, f := range r.favorites {
		if f.UserID == userID && f.TargetType == targetType && f.TargetID == targetID {
			return f, nil
		}
	}
	return nil, nil
}

func (r *UserFavoriteRepository) GetTargetIDs(ctx context.Context, userID string, targetType domain.UserFavoriteTargetType) ([]string, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()

	targetIDs := make([]string, 0)
	for _, f := range r.favorites {
		if f.UserID == userID && f.TargetType == targetType {
			targetIDs = append(targetIDs, f.TargetID)
		}
	}
	return targetIDs, nil
}
