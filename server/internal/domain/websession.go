package domain

import "time"

type WebSession struct {
	ID        string
	UserID    string
	Token     string
	ExpiresAt time.Time
	CreatedAt time.Time
}

func (ws *WebSession) IsExpired() bool {
	return time.Now().After(ws.ExpiresAt)
}
