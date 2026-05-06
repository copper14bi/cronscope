// Package historyapi exposes job event history over HTTP as JSON.
package historyapi

import (
	"encoding/json"
	"net/http"
	"strings"

	"github.com/yourorg/cronscope/internal/history"
)

// Handler returns an http.Handler that serves event history.
// GET /history        — all events
// GET /history/{job}  — events for a specific job
func Handler(store *history.Store) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodGet {
			http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
			return
		}

		// Strip leading slash and optional "/history" prefix.
		path := strings.TrimPrefix(r.URL.Path, "/")
		path = strings.TrimPrefix(path, "history")
		path = strings.TrimPrefix(path, "/")

		var events []history.Event
		if path == "" {
			events = store.All()
		} else {
			events = store.ForJob(path)
		}

		if events == nil {
			events = []history.Event{}
		}

		w.Header().Set("Content-Type", "application/json")
		if err := json.NewEncoder(w).Encode(events); err != nil {
			http.Error(w, "encoding error", http.StatusInternalServerError)
		}
	})
}
