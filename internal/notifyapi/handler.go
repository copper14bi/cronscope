// Package notifyapi provides an HTTP handler for listing and managing
// active alert notifications sent by cronscope.
package notifyapi

import (
	"encoding/json"
	"net/http"
	"time"
)

// NotifyRecord represents a single dispatched alert notification.
type NotifyRecord struct {
	JobName   string    `json:"job_name"`
	Reason    string    `json:"reason"`
	SentAt    time.Time `json:"sent_at"`
	Cooldown  bool      `json:"within_cooldown"`
}

// Store is the interface for retrieving notification history.
type Store interface {
	Recent(limit int) []NotifyRecord
}

// Handler returns an http.Handler that serves recent alert notifications.
// GET /notify        — returns the last N notifications (default 50).
func Handler(store Store) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodGet {
			http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
			return
		}

		records := store.Recent(50)
		if records == nil {
			records = []NotifyRecord{}
		}

		w.Header().Set("Content-Type", "application/json")
		if err := json.NewEncoder(w).Encode(records); err != nil {
			http.Error(w, "encoding error", http.StatusInternalServerError)
		}
	})
}
