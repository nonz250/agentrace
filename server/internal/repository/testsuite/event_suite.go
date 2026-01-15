package testsuite

import (
	"context"
	"time"

	"github.com/satetsu888/agentrace/server/internal/domain"
	"github.com/satetsu888/agentrace/server/internal/repository"
	"github.com/stretchr/testify/suite"
)

// EventRepositorySuite tests EventRepository implementations
type EventRepositorySuite struct {
	suite.Suite
	Repo        repository.EventRepository
	SessionRepo repository.SessionRepository // Optional: for FK constraint support
	Cleanup     func()
}

// createTestSession creates a session for FK constraint tests
func (s *EventRepositorySuite) createTestSession(id string) {
	if s.SessionRepo == nil {
		return
	}
	ctx := context.Background()
	session := &domain.Session{
		ID:              id,
		ClaudeSessionID: "claude-" + id,
	}
	_ = s.SessionRepo.Create(ctx, session)
}

func (s *EventRepositorySuite) TearDownTest() {
	if s.Cleanup != nil {
		s.Cleanup()
	}
}

func (s *EventRepositorySuite) TestCreate() {
	ctx := context.Background()

	s.createTestSession("session-1")

	event := &domain.Event{
		SessionID: "session-1",
		UUID:      "event-uuid-1",
		EventType: "message",
		Payload: map[string]interface{}{
			"content": "Hello, world!",
		},
	}

	err := s.Repo.Create(ctx, event)
	s.Require().NoError(err)

	// ID should be auto-generated
	s.NotEmpty(event.ID)

	// CreatedAt should be set
	s.False(event.CreatedAt.IsZero())
}

func (s *EventRepositorySuite) TestCreate_DuplicateUUID() {
	ctx := context.Background()

	s.createTestSession("session-dup")

	// Create first event
	event1 := &domain.Event{
		SessionID: "session-dup",
		UUID:      "duplicate-uuid",
		EventType: "message",
		Payload:   map[string]interface{}{},
	}
	err := s.Repo.Create(ctx, event1)
	s.Require().NoError(err)

	// Try to create second event with same UUID in same session
	event2 := &domain.Event{
		SessionID: "session-dup",
		UUID:      "duplicate-uuid",
		EventType: "message",
		Payload:   map[string]interface{}{},
	}
	err = s.Repo.Create(ctx, event2)
	s.Require().Error(err)
	s.ErrorIs(err, repository.ErrDuplicateEvent)
}

func (s *EventRepositorySuite) TestCreate_SameUUIDDifferentSession() {
	ctx := context.Background()

	s.createTestSession("session-a")
	s.createTestSession("session-b")

	// Create event in first session
	event1 := &domain.Event{
		SessionID: "session-a",
		UUID:      "same-uuid",
		EventType: "message",
		Payload:   map[string]interface{}{},
	}
	err := s.Repo.Create(ctx, event1)
	s.Require().NoError(err)

	// Create event with same UUID in different session - should succeed
	event2 := &domain.Event{
		SessionID: "session-b",
		UUID:      "same-uuid",
		EventType: "message",
		Payload:   map[string]interface{}{},
	}
	err = s.Repo.Create(ctx, event2)
	s.Require().NoError(err)
}

func (s *EventRepositorySuite) TestFindBySessionID() {
	ctx := context.Background()

	sessionID := "session-find"
	s.createTestSession(sessionID)
	s.createTestSession("other-session")

	// Create multiple events
	for i := 0; i < 5; i++ {
		event := &domain.Event{
			SessionID: sessionID,
			UUID:      "event-" + string(rune('a'+i)),
			EventType: "message",
			Payload:   map[string]interface{}{"index": i},
			CreatedAt: time.Now().Add(time.Duration(i) * time.Millisecond),
		}
		err := s.Repo.Create(ctx, event)
		s.Require().NoError(err)
	}

	// Create event for different session
	otherEvent := &domain.Event{
		SessionID: "other-session",
		UUID:      "other-event",
		EventType: "message",
		Payload:   map[string]interface{}{},
	}
	err := s.Repo.Create(ctx, otherEvent)
	s.Require().NoError(err)

	// Find events for session
	events, err := s.Repo.FindBySessionID(ctx, sessionID)
	s.Require().NoError(err)
	s.Len(events, 5)

	// All events should belong to the session
	for _, e := range events {
		s.Equal(sessionID, e.SessionID)
	}
}

func (s *EventRepositorySuite) TestFindBySessionID_ChronologicalOrder() {
	ctx := context.Background()

	sessionID := "session-chrono"
	s.createTestSession(sessionID)

	baseTime := time.Now()

	// Create events in non-sequential order but with specific timestamps
	// We'll create them out of order to ensure sorting is by CreatedAt, not insertion order
	timestamps := []time.Duration{
		300 * time.Millisecond, // third
		100 * time.Millisecond, // first
		500 * time.Millisecond, // fifth
		200 * time.Millisecond, // second
		400 * time.Millisecond, // fourth
	}

	for i, offset := range timestamps {
		event := &domain.Event{
			SessionID: sessionID,
			UUID:      "chrono-event-" + string(rune('a'+i)),
			EventType: "message",
			Payload:   map[string]interface{}{"order": i},
			CreatedAt: baseTime.Add(offset),
		}
		err := s.Repo.Create(ctx, event)
		s.Require().NoError(err)
	}

	// Find events
	events, err := s.Repo.FindBySessionID(ctx, sessionID)
	s.Require().NoError(err)
	s.Require().Len(events, 5)

	// Verify events are in chronological order (ascending by CreatedAt)
	for i := 1; i < len(events); i++ {
		s.True(
			events[i-1].CreatedAt.Before(events[i].CreatedAt) || events[i-1].CreatedAt.Equal(events[i].CreatedAt),
			"Events should be in chronological order: event[%d].CreatedAt=%v should be <= event[%d].CreatedAt=%v",
			i-1, events[i-1].CreatedAt, i, events[i].CreatedAt,
		)
	}

	// Also verify the first and last events have the expected timestamps
	s.Equal(baseTime.Add(100*time.Millisecond).UnixNano(), events[0].CreatedAt.UnixNano(), "First event should have earliest timestamp")
	s.Equal(baseTime.Add(500*time.Millisecond).UnixNano(), events[4].CreatedAt.UnixNano(), "Last event should have latest timestamp")
}

func (s *EventRepositorySuite) TestFindBySessionID_Empty() {
	ctx := context.Background()

	events, err := s.Repo.FindBySessionID(ctx, "non-existing-session")
	s.Require().NoError(err)
	s.Empty(events)
}

func (s *EventRepositorySuite) TestCountBySessionID() {
	ctx := context.Background()

	sessionID := "session-count"
	s.createTestSession(sessionID)

	// Create multiple events
	for i := 0; i < 7; i++ {
		event := &domain.Event{
			SessionID: sessionID,
			UUID:      "count-event-" + string(rune('a'+i)),
			EventType: "message",
			Payload:   map[string]interface{}{},
		}
		err := s.Repo.Create(ctx, event)
		s.Require().NoError(err)
	}

	count, err := s.Repo.CountBySessionID(ctx, sessionID)
	s.Require().NoError(err)
	s.Equal(7, count)
}

func (s *EventRepositorySuite) TestCountBySessionID_Empty() {
	ctx := context.Background()

	count, err := s.Repo.CountBySessionID(ctx, "non-existing-session")
	s.Require().NoError(err)
	s.Equal(0, count)
}
