package domain

import "time"

// PasswordCredential represents password authentication for a user
type PasswordCredential struct {
	ID           string
	UserID       string
	PasswordHash string // bcrypt hash
	CreatedAt    time.Time
	UpdatedAt    time.Time
}
