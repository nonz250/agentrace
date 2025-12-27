package memory

import "github.com/satetsu888/agentrace/server/internal/repository"

func NewRepositories() *repository.Repositories {
	return &repository.Repositories{
		Session: NewSessionRepository(),
		Event:   NewEventRepository(),
	}
}
