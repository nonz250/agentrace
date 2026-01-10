package domain

import "time"

type UserFavoriteTargetType string

const (
	UserFavoriteTargetTypeSession UserFavoriteTargetType = "session"
	UserFavoriteTargetTypePlan    UserFavoriteTargetType = "plan"
)

func (t UserFavoriteTargetType) IsValid() bool {
	switch t {
	case UserFavoriteTargetTypeSession, UserFavoriteTargetTypePlan:
		return true
	}
	return false
}

type UserFavorite struct {
	ID         string
	UserID     string
	TargetType UserFavoriteTargetType
	TargetID   string
	CreatedAt  time.Time
}
