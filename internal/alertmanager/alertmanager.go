// Package alertmanager deduplicates and dispatches webhook alerts.
package alertmanager

import (
	"sync"
	"time"

	"github.com/cronscope/cronscope/internal/webhook"
)

// Alert represents a single alert payload.
type Alert struct {
	Job     string    `json:"job"`
	Kind    string    `json:"kind"` // "missed" | "long_running"
	Message string    `json:"message"`
	SentAt  time.Time `json:"sent_at"`
}

// SuppressionRule records until when a job's alerts are suppressed.
type SuppressionRule struct {
	Job    string    `json:"job"`
	Until  time.Time `json:"until"`
}

// Manager deduplicates alerts within a cooldown window and dispatches them.
type Manager struct {
	mu          sync.Mutex
	client      *webhook.Client
	cooldown    time.Duration
	lastSent    map[string]time.Time
	suppressions map[string]time.Time
}

// New creates a Manager with the given webhook client and cooldown duration.
func New(client *webhook.Client, cooldown time.Duration) *Manager {
	return &Manager{
		client:       client,
		cooldown:     cooldown,
		lastSent:     make(map[string]time.Time),
		suppressions: make(map[string]time.Time),
	}
}

// Send dispatches an alert unless it is within cooldown or suppressed.
func (m *Manager) Send(a Alert) error {
	m.mu.Lock()
	defer m.mu.Unlock()

	now := time.Now()
	key := a.Job + "|" + a.Kind

	if until, ok := m.suppressions[a.Job]; ok && now.Before(until) {
		return nil
	}

	if last, ok := m.lastSent[key]; ok && now.Sub(last) < m.cooldown {
		return nil
	}

	if a.SentAt.IsZero() {
		a.SentAt = now
	}
	m.lastSent[key] = now
	return m.client.Send(a)
}

// Suppress silences alerts for the named job for the given duration.
func (m *Manager) Suppress(job string, d time.Duration) {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.suppressions[job] = time.Now().Add(d)
}

// Unsuppress removes any active suppression for the named job.
func (m *Manager) Unsuppress(job string) {
	m.mu.Lock()
	defer m.mu.Unlock()
	delete(m.suppressions, job)
}

// Suppressions returns a snapshot of all active suppression rules.
func (m *Manager) Suppressions() []SuppressionRule {
	m.mu.Lock()
	defer m.mu.Unlock()
	now := time.Now()
	var rules []SuppressionRule
	for job, until := range m.suppressions {
		if now.Before(until) {
			rules = append(rules, SuppressionRule{Job: job, Until: until})
		}
	}
	return rules
}
