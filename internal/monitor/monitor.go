// Package monitor provides job state tracking and missed/long-running detection.
package monitor

import (
	"sync"
	"time"

	"github.com/cronscope/cronscope/internal/config"
	"github.com/cronscope/cronscope/internal/schedule"
	"github.com/cronscope/cronscope/internal/webhook"
)

// JobState holds runtime state for a single monitored job.
type JobState struct {
	Name      string
	StartedAt *time.Time
	LastSeen  *time.Time
}

// Monitor tracks job states and fires webhook alerts.
type Monitor struct {
	mu     sync.Mutex
	jobs   map[string]*JobState
	cfg    *config.Config
	client *webhook.Client
}

// New creates a Monitor from the provided config.
func New(cfg *config.Config, client *webhook.Client) *Monitor {
	jobs := make(map[string]*JobState, len(cfg.Jobs))
	for _, j := range cfg.Jobs {
		jobs[j.Name] = &JobState{Name: j.Name}
	}
	return &Monitor{jobs: jobs, cfg: cfg, client: client}
}

// RecordStart marks a job as started at the given time.
func (m *Monitor) RecordStart(name string, at time.Time) {
	m.mu.Lock()
	defer m.mu.Unlock()
	if s, ok := m.jobs[name]; ok {
		s.StartedAt = &at
		s.LastSeen = &at
	}
}

// RecordFinish clears the running state for a job.
func (m *Monitor) RecordFinish(name string, at time.Time) {
	m.mu.Lock()
	defer m.mu.Unlock()
	if s, ok := m.jobs[name]; ok {
		s.StartedAt = nil
		s.LastSeen = &at
	}
}

// Check evaluates all jobs at the given time and sends alerts as needed.
func (m *Monitor) Check(now time.Time) {
	m.mu.Lock()
	defer m.mu.Unlock()

	for _, j := range m.cfg.Jobs {
		state := m.jobs[j.Name]

		if state.StartedAt != nil {
			if schedule.IsLongRunning(*state.StartedAt, now, j.MaxDuration) {
				_ = m.client.Send(webhook.Payload{
					Event:   "long_running",
					Job:     j.Name,
					Message: "job has exceeded its maximum allowed duration",
				})
			}
			continue
		}

		if schedule.IsMissed(j.Schedule, state.LastSeen, now) {
			_ = m.client.Send(webhook.Payload{
				Event:   "missed",
				Job:     j.Name,
				Message: "job did not run within the expected window",
			})
		}
	}
}
