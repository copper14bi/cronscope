// Package healthcheck provides an HTTP handler that exposes a simple liveness
// endpoint for cronscope. External monitoring tools can poll this endpoint to
// confirm the daemon is running and its job registry is populated.
package healthcheck

import (
	"encoding/json"
	"net/http"
	"time"
)

// StateReader is the subset of state.State behaviour required by the handler.
type StateReader interface {
	Names() []string
}

// Response is the JSON body returned by the health endpoint.
type Response struct {
	Status    string    `json:"status"`
	JobCount  int       `json:"job_count"`
	CheckedAt time.Time `json:"checked_at"`
}

// Handler returns an http.HandlerFunc that writes a JSON health response.
// The status field is always "ok" while the daemon is alive; consumers can
// treat a non-200 or missing response as a sign the process has died.
func Handler(sr StateReader) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		resp := Response{
			Status:    "ok",
			JobCount:  len(sr.Names()),
			CheckedAt: time.Now().UTC(),
		}

		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)

		if err := json.NewEncoder(w).Encode(resp); err != nil {
			http.Error(w, "failed to encode response", http.StatusInternalServerError)
		}
	}
}
