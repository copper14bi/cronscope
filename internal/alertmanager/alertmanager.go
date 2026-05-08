// Package alertmanager deduplicates and dispatches webhook alerts for
// missed or long-running cron jobs.
package alertmanager

import (
	"errors"
	"sync"
	"time"

	"github.com/user/cronscope/internal/webhook"
)

// Alert represents a single alert payload.
type Alert struct {
	JobName   string    `json:"job_name"`
	Kind      string    `json:"kind"` // "missed" | "long_running"
	Message   string    `json:"message"`
	Timestamp time.Time `json:"timestamp"`
}

type entry struct {
	lastSent time.Time
	last     Alert
}

// Manager deduplicates alerts and forwards them via a webhook client.
type Manager struct {
	mu          sync.Mutex
	client      *webhook.Client
	cooldown    time.Duration
	entries     map[string]*entry
	suppressed  map[string]time.Time
}

// New returns a Manager with the given webhook client and alert cooldown.
func New(client *webhook.Client, cooldown time.Duration) *Manager {
	return &Manager{
		client:     client,
		cooldown:   cooldown,
		entries:    make(map[string]*entry),
		suppressed: make(map[string]time.Time),
	}
}

// Send dispatches an alert unless one was already sent within the cooldown
// window or the job is suppressed.
func (m *Manager) Send(a Alert) error {
	m.mu.Lock()
	defer m.mu.Unlock()

	if until, ok := m.suppressed[a.JobName]; ok && time.Now().Before(until) {
		return nil
	}

	e, ok := m.entries[a.JobName]
	if ok && time.Since(e.lastSent) < m.cooldown {
		return nil
	}

	if err := m.client.Send(a); err != nil {
		return err
	}

	m.entries[a.JobName] = &entry{lastSent: time.Now(), last: a}
	return nil
}

// Retry bypasses the cooldown and immediately re-sends the last alert for
// the given job. Returns an error if no previous alert exists.
func (m *Manager) Retry(jobName string) error {
	m.mu.Lock()
	defer m.mu.Unlock()

	e, ok := m.entries[jobName]
	if !ok {
		return errors.New("no previous alert found for job: " + jobName)
	}

	if err := m.client.Send(e.last); err != nil {
		return err
	}
	e.lastSent = time.Now()
	return nil
}

// Suppress silences alerts for jobName until the given time.
func (m *Manager) Suppress(jobName string, until time.Time) {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.suppressed[jobName] = until
}

// Unsuppress removes a suppression for jobName.
func (m *Manager) Unsuppress(jobName string) {
	m.mu.Lock()
	defer m.mu.Unlock()
	delete(m.suppressed, jobName)
}

// IsSuppressed reports whether jobName is currently suppressed.
func (m *Manager) IsSuppressed(jobName string) bool {
	m.mu.Lock()
	defer m.mu.Unlock()
	until, ok := m.suppressed[jobName]
	return ok && time.Now().Before(until)
}
