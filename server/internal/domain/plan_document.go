package domain

import "time"

type PlanDocumentStatus string

const (
	PlanDocumentStatusDraft          PlanDocumentStatus = "draft"
	PlanDocumentStatusPlanning       PlanDocumentStatus = "planning"
	PlanDocumentStatusPending        PlanDocumentStatus = "pending"
	PlanDocumentStatusImplementation PlanDocumentStatus = "implementation"
	PlanDocumentStatusComplete       PlanDocumentStatus = "complete"
)

func (s PlanDocumentStatus) IsValid() bool {
	switch s {
	case PlanDocumentStatusDraft, PlanDocumentStatusPlanning, PlanDocumentStatusPending, PlanDocumentStatusImplementation, PlanDocumentStatusComplete:
		return true
	}
	return false
}

type PlanDocument struct {
	ID           string
	Description  string
	Body         string
	GitRemoteURL string
	Status       PlanDocumentStatus
	CreatedAt    time.Time
	UpdatedAt    time.Time
}
