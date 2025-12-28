package domain

import "time"

type APIKey struct {
	ID         string
	UserID     string
	Name       string     // Key name (e.g., "MacBook Pro", "Work PC")
	KeyHash    string     // bcrypt hash
	KeyPrefix  string     // "agtr_xxxx..." (for display)
	LastUsedAt *time.Time
	CreatedAt  time.Time
}
