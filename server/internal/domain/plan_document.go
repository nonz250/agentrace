package domain

import "time"

type PlanDocumentStatus string

const (
	PlanDocumentStatusScratch        PlanDocumentStatus = "scratch"
	PlanDocumentStatusDraft          PlanDocumentStatus = "draft"
	PlanDocumentStatusPlanning       PlanDocumentStatus = "planning"
	PlanDocumentStatusPending        PlanDocumentStatus = "pending"
	PlanDocumentStatusImplementation PlanDocumentStatus = "implementation"
	PlanDocumentStatusComplete       PlanDocumentStatus = "complete"
)

func (s PlanDocumentStatus) IsValid() bool {
	switch s {
	case PlanDocumentStatusScratch, PlanDocumentStatusDraft, PlanDocumentStatusPlanning, PlanDocumentStatusPending, PlanDocumentStatusImplementation, PlanDocumentStatusComplete:
		return true
	}
	return false
}

type PlanDocument struct {
	ID          string
	ProjectID   string // reference to Project
	Description string
	Body        string
	Status      PlanDocumentStatus
	CreatedAt   time.Time
	UpdatedAt   time.Time
}

// PlanDocumentQuery represents search criteria for plan documents
type PlanDocumentQuery struct {
	ProjectID string                 // Filter by project ID (empty = all projects)
	Statuses  []PlanDocumentStatus   // Filter by statuses (empty = all statuses)
	Limit     int                    // Max results (0 = use default)
	Offset    int                    // Skip first N results
}
