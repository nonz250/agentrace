package postgres

import (
	"context"
	"database/sql"
	"time"

	"github.com/google/uuid"
	"github.com/satetsu888/agentrace/server/internal/domain"
)

type UserFavoriteRepository struct {
	db *DB
}

func NewUserFavoriteRepository(db *DB) *UserFavoriteRepository {
	return &UserFavoriteRepository{db: db}
}

func (r *UserFavoriteRepository) Create(ctx context.Context, favorite *domain.UserFavorite) error {
	if favorite.ID == "" {
		favorite.ID = uuid.New().String()
	}
	if favorite.CreatedAt.IsZero() {
		favorite.CreatedAt = time.Now()
	}

	_, err := r.db.ExecContext(ctx,
		`INSERT INTO user_favorites (id, user_id, target_type, target_id, created_at)
		 VALUES ($1, $2, $3, $4, $5)`,
		favorite.ID, favorite.UserID, string(favorite.TargetType), favorite.TargetID, favorite.CreatedAt,
	)
	return err
}

func (r *UserFavoriteRepository) Delete(ctx context.Context, id string) error {
	_, err := r.db.ExecContext(ctx, `DELETE FROM user_favorites WHERE id = $1`, id)
	return err
}

func (r *UserFavoriteRepository) DeleteByUserAndTarget(ctx context.Context, userID string, targetType domain.UserFavoriteTargetType, targetID string) error {
	_, err := r.db.ExecContext(ctx,
		`DELETE FROM user_favorites WHERE user_id = $1 AND target_type = $2 AND target_id = $3`,
		userID, string(targetType), targetID,
	)
	return err
}

func (r *UserFavoriteRepository) FindByUserID(ctx context.Context, userID string) ([]*domain.UserFavorite, error) {
	rows, err := r.db.QueryContext(ctx,
		`SELECT id, user_id, target_type, target_id, created_at
		 FROM user_favorites WHERE user_id = $1 ORDER BY created_at DESC`,
		userID,
	)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var favorites []*domain.UserFavorite
	for rows.Next() {
		favorite, err := r.scanFavoriteFromRows(rows)
		if err != nil {
			return nil, err
		}
		favorites = append(favorites, favorite)
	}

	return favorites, rows.Err()
}

func (r *UserFavoriteRepository) FindByUserAndTargetType(ctx context.Context, userID string, targetType domain.UserFavoriteTargetType) ([]*domain.UserFavorite, error) {
	rows, err := r.db.QueryContext(ctx,
		`SELECT id, user_id, target_type, target_id, created_at
		 FROM user_favorites WHERE user_id = $1 AND target_type = $2 ORDER BY created_at DESC`,
		userID, string(targetType),
	)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var favorites []*domain.UserFavorite
	for rows.Next() {
		favorite, err := r.scanFavoriteFromRows(rows)
		if err != nil {
			return nil, err
		}
		favorites = append(favorites, favorite)
	}

	return favorites, rows.Err()
}

func (r *UserFavoriteRepository) FindByUserAndTarget(ctx context.Context, userID string, targetType domain.UserFavoriteTargetType, targetID string) (*domain.UserFavorite, error) {
	return r.scanFavorite(r.db.QueryRowContext(ctx,
		`SELECT id, user_id, target_type, target_id, created_at
		 FROM user_favorites WHERE user_id = $1 AND target_type = $2 AND target_id = $3`,
		userID, string(targetType), targetID,
	))
}

func (r *UserFavoriteRepository) GetTargetIDs(ctx context.Context, userID string, targetType domain.UserFavoriteTargetType) ([]string, error) {
	rows, err := r.db.QueryContext(ctx,
		`SELECT target_id FROM user_favorites WHERE user_id = $1 AND target_type = $2`,
		userID, string(targetType),
	)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var targetIDs []string
	for rows.Next() {
		var targetID string
		if err := rows.Scan(&targetID); err != nil {
			return nil, err
		}
		targetIDs = append(targetIDs, targetID)
	}

	return targetIDs, rows.Err()
}

func (r *UserFavoriteRepository) scanFavorite(row *sql.Row) (*domain.UserFavorite, error) {
	var favorite domain.UserFavorite
	var targetType string

	err := row.Scan(&favorite.ID, &favorite.UserID, &targetType, &favorite.TargetID, &favorite.CreatedAt)
	if err == sql.ErrNoRows {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}

	favorite.TargetType = domain.UserFavoriteTargetType(targetType)

	return &favorite, nil
}

func (r *UserFavoriteRepository) scanFavoriteFromRows(rows *sql.Rows) (*domain.UserFavorite, error) {
	var favorite domain.UserFavorite
	var targetType string

	err := rows.Scan(&favorite.ID, &favorite.UserID, &targetType, &favorite.TargetID, &favorite.CreatedAt)
	if err != nil {
		return nil, err
	}

	favorite.TargetType = domain.UserFavoriteTargetType(targetType)

	return &favorite, nil
}
