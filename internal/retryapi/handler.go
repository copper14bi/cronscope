// Package retryapi exposes an HTTP endpoint for manually retrying
// alert delivery for a named cron job.
package retryapi

import (
	"encoding/json"
	"net/http"
	"strings"
	"time"
)

// Retryer is the interface satisfied by alertmanager.Manager.
type Retryer interface {
	Retry(jobName string) error
	IsSuppressed(jobName string) bool
}

type response struct {
	Job       string    `json:"job"`
	Retried   bool      `json:"retried"`
	Timestamp time.Time `json:"timestamp"`
	Message   string    `json:"message,omitempty"`
}

// Handler returns an http.Handler that accepts POST /retry?job=<name> and
// triggers an immediate alert retry via the provided Retryer.
func Handler(r Retryer) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, req *http.Request) {
		w.Header().Set("Content-Type", "application/json")

		if req.Method != http.MethodPost {
			w.WriteHeader(http.StatusMethodNotAllowed)
			_ = json.NewEncoder(w).Encode(response{Message: "method not allowed"})
			return
		}

		job := strings.TrimSpace(req.URL.Query().Get("job"))
		if job == "" {
			w.WriteHeader(http.StatusBadRequest)
			_ = json.NewEncoder(w).Encode(response{Message: "missing job query parameter"})
			return
		}

		if r.IsSuppressed(job) {
			w.WriteHeader(http.StatusConflict)
			_ = json.NewEncoder(w).Encode(response{
				Job:     job,
				Retried: false,
				Message: "job is suppressed; remove suppression before retrying",
			})
			return
		}

		if err := r.Retry(job); err != nil {
			w.WriteHeader(http.StatusInternalServerError)
			_ = json.NewEncoder(w).Encode(response{
				Job:     job,
				Retried: false,
				Message: err.Error(),
			})
			return
		}

		w.WriteHeader(http.StatusOK)
		_ = json.NewEncoder(w).Encode(response{
			Job:       job,
			Retried:   true,
			Timestamp: time.Now().UTC(),
		})
	})
}
