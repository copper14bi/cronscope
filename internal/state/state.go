// Package state tracks the runtime state of monitored cron jobs,
// recording when each job was last seen and whether it is currently
// considered running.
package state

import (
	"sync"
	"time"
)

// JobState holds the observed state for a single cron job.
type JobState struct {
	LastSeen  time.Time
	Running   bool
	StartedAt time.Time
}

// Store is a thread-safe in-memory store for job states.
type Store struct {
	mu    sync.RWMutex
	jobs  map[string]*JobState
}

// New creates an empty Store.
func New() *Store {
	return &Store{
		jobs: make(map[string]*JobState),
	}
}

// MarkSeen records that the named job was observed at the given time.
// If the job was previously marked as running it is cleared.
func (s *Store) MarkSeen(name string, at time.Time) {
	s.mu.Lock()
	defer s.mu.Unlock()

	js, ok := s.jobs[name]
	if !ok {
		js = &JobState{}
		s.jobs[name] = js
	}
	js.LastSeen = at
	js.Running = false
	js.StartedAt = time.Time{}
}

// MarkRunning records that the named job started at the given time.
func (s *Store) MarkRunning(name string, at time.Time) {
	s.mu.Lock()
	defer s.mu.Unlock()

	js, ok := s.jobs[name]
	if !ok {
		js = &JobState{}
		s.jobs[name] = js
	}
	js.Running = true
	js.StartedAt = at
}

// Get returns a copy of the JobState for the named job and whether it exists.
func (s *Store) Get(name string) (JobState, bool) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	js, ok := s.jobs[name]
	if !ok {
		return JobState{}, false
	}
	return *js, true
}

// Names returns all job names currently tracked by the store.
func (s *Store) Names() []string {
	s.mu.RLock()
	defer s.mu.RUnlock()

	names := make([]string, 0, len(s.jobs))
	for n := range s.jobs {
		names = append(names, n)
	}
	return names
}
