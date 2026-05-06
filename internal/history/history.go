// Package history tracks alert and job event history for reporting.
package history

import (
	"sync"
	"time"
)

// EventType classifies what kind of event occurred.
type EventType string

const (
	EventMissed     EventType = "missed"
	EventLongRunning EventType = "long_running"
	EventRecovered  EventType = "recovered"
)

// Event represents a single recorded job event.
type Event struct {
	JobName   string
	Type      EventType
	OccurredAt time.Time
	Message   string
}

// Store holds an in-memory ring buffer of recent events.
type Store struct {
	mu     sync.RWMutex
	events []Event
	maxLen int
}

// New creates a Store that retains up to maxLen events.
func New(maxLen int) *Store {
	if maxLen <= 0 {
		maxLen = 100
	}
	return &Store{maxLen: maxLen}
}

// Record appends a new event, evicting the oldest if the buffer is full.
func (s *Store) Record(e Event) {
	if e.OccurredAt.IsZero() {
		e.OccurredAt = time.Now().UTC()
	}
	s.mu.Lock()
	defer s.mu.Unlock()
	if len(s.events) >= s.maxLen {
		s.events = s.events[1:]
	}
	s.events = append(s.events, e)
}

// All returns a shallow copy of all stored events, oldest first.
func (s *Store) All() []Event {
	s.mu.RLock()
	defer s.mu.RUnlock()
	out := make([]Event, len(s.events))
	copy(out, s.events)
	return out
}

// ForJob returns events for a specific job name.
func (s *Store) ForJob(name string) []Event {
	s.mu.RLock()
	defer s.mu.RUnlock()
	var out []Event
	for _, e := range s.events {
		if e.JobName == name {
			out = append(out, e)
		}
	}
	return out
}

// Len returns the current number of stored events.
func (s *Store) Len() int {
	s.mu.RLock()
	defer s.mu.RUnlock()
	return len(s.events)
}
