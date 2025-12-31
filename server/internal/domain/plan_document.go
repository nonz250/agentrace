package domain

import "time"

type PlanDocument struct {
	ID           string
	Description  string
	Body         string
	GitRemoteURL string
	CreatedAt    time.Time
	UpdatedAt    time.Time
}
