package domain

import "time"

// DefaultProjectID is the UUID for the "no project" default project
const DefaultProjectID = "00000000-0000-0000-0000-000000000000"

type Project struct {
	ID                     string
	CanonicalGitRepository string // Normalized HTTP-style git URL (empty = no project)
	CreatedAt              time.Time
}

// IsDefaultProject returns true if this is the "no project" default project
func (p *Project) IsDefaultProject() bool {
	return p.CanonicalGitRepository == ""
}
