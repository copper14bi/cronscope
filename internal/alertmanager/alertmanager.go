// Package alertmanager coordinates alert deduplication and dispatch via webhook.
package alertmanager

import (
	"fmt"
	"sync"
	"time"

	"github.com/yourorg/cronscope/internal/webhook"
)

// AlertType classifies the kind of alert being raised.
type AlertType string

const (
	AlertMissed     AlertType = "missed"
	AlertLongRunning AlertType = "long_running"
)

// Alert represents a single alert event for a cron job.
type Alert struct {
	JobName   string    `json:"job_name"`
	Type      AlertType `json:"type"`
	Message   string    `json:"message"`
	Timestamp time.Time `json:"timestamp"`
}

// Manager deduplicates and dispatches alerts.
type Manager struct {
	client   *webhook.Client
	mu       sync.Mutex
	sent     map[string]time.Time
	cooldown time.Duration
}

// New creates a Manager with the given webhook client and alert cooldown duration.
func New(client *webhook.Client, cooldown time.Duration) *Manager {
	return &Manager{
		client:   client,
		sent:     make(map[string]time.Time),
		cooldown: cooldown,
	}
}

// Send dispatches an alert if it has not been sent within the cooldown window.
// Returns true if the alert was dispatched, false if suppressed.
func (m *Manager) Send(a Alert) (bool, error) {
	key := fmt.Sprintf("%s:%s", a.JobName, a.Type)

	m.mu.Lock()
	last, exists := m.sent[key]
	if exists && time.Since(last) < m.cooldown {
		m.mu.Unlock()
		return false, nil
	}
	m.sent[key] = time.Now()
	m.mu.Unlock()

	if err := m.client.Send(a); err != nil {
		return false, fmt.Errorf("alertmanager: send %q: %w", key, err)
	}
	return true, nil
}

// Reset clears the deduplication state for a specific job, allowing the next
// alert for that job to be dispatched immediately.
func (m *Manager) Reset(jobName string) {
	m.mu.Lock()
	defer m.mu.Unlock()
	for _, t := range []AlertType{AlertMissed, AlertLongRunning} {
		delete(m.sent, fmt.Sprintf("%s:%s", jobName, t))
	}
}
