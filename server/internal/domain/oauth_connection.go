package domain

import "time"

// OAuthConnection represents a connection between a user and an OAuth provider
type OAuthConnection struct {
	ID         string
	UserID     string
	Provider   string // "github", "google", etc.
	ProviderID string // Provider's user ID
	CreatedAt  time.Time
}
