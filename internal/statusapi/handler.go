// Package statusapi provides an HTTP handler that exposes the current
// runtime status of all monitored cron jobs, including last-seen time,
// running state, and recent alert history.
package statusapi

import (
	"encoding/json"
	"net/http"
	"time"

	"github.com/cronscope/cronscope/internal/state"
)

// JobStatus is the JSON representation of a single job's current state.
type JobStatus struct {
	Name      string     `json:"name"`
	LastSeen  *time.Time `json:"last_seen,omitempty"`
	Running   bool       `json:"running"`
	RunSince  *time.Time `json:"run_since,omitempty"`
}

// StatusResponse is the top-level JSON envelope returned by the handler.
type StatusResponse struct {
	Jobs      []JobStatus `json:"jobs"`
	ReportedAt time.Time  `json:"reported_at"`
}

// Store is the subset of state.Store behaviour required by this handler.
type Store interface {
	Names() []string
	Get(name string) (state.Entry, bool)
}

// Handler returns an http.HandlerFunc that serialises the current job
// states from store as JSON. Only GET requests are accepted.
func Handler(store Store) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodGet {
			http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
			return
		}

		names := store.Names()
		jobs := make([]JobStatus, 0, len(names))

		for _, name := range names {
			entry, ok := store.Get(name)
			if !ok {
				continue
			}

			js := JobStatus{
				Name:    name,
				Running: entry.Running,
			}
			if !entry.LastSeen.IsZero() {
				t := entry.LastSeen
				js.LastSeen = &t
			}
			if entry.Running && !entry.RunSince.IsZero() {
				t := entry.RunSince
				js.RunSince = &t
			}
			jobs = append(jobs, js)
		}

		resp := StatusResponse{
			Jobs:       jobs,
			ReportedAt: time.Now().UTC(),
		}

		w.Header().Set("Content-Type", "application/json")
		_ = json.NewEncoder(w).Encode(resp)
	}
}
