package domain

import "time"

type User struct {
	ID          string
	Email       string
	DisplayName string // 空の場合はEmailを表示
	CreatedAt   time.Time
}

// GetDisplayName returns DisplayName if set, otherwise Email
func (u *User) GetDisplayName() string {
	if u.DisplayName != "" {
		return u.DisplayName
	}
	return u.Email
}
