// Package snoozeapi provides an HTTP handler for temporarily snoozing
// alerts for a specific job, suppressing notifications for a given duration.
package snoozeapi

import (
	"encoding/json"
	"net/http"
	"strings"
	"time"
)

// Snoozer is the interface for managing per-job snooze state.
type Snoozer interface {
	Snooze(job string, until time.Time)
	Unsnooze(job string)
	SnoozedUntil(job string) (time.Time, bool)
}

type snoozeRequest struct {
	Duration string `json:"duration"`
}

type snoozeRecord struct {
	Job   string    `json:"job"`
	Until time.Time `json:"until"`
}

// Handler returns an http.Handler for the snooze API.
// POST /snooze/{job}?duration=15m  — snooze a job
// GET  /snooze/{job}               — check snooze status
// DELETE /snooze/{job}             — cancel snooze
func Handler(s Snoozer) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		job := strings.TrimPrefix(r.URL.Path, "/snooze/")
		if job == "" {
			http.Error(w, "job name required", http.StatusBadRequest)
			return
		}

		w.Header().Set("Content-Type", "application/json")

		switch r.Method {
		case http.MethodPost:
			var req snoozeRequest
			if err := json.NewDecoder(r.Body).Decode(&req); err != nil || req.Duration == "" {
				http.Error(w, "invalid request body", http.StatusBadRequest)
				return
			}
			d, err := time.ParseDuration(req.Duration)
			if err != nil || d <= 0 {
				http.Error(w, "invalid duration", http.StatusBadRequest)
				return
			}
			until := time.Now().Add(d)
			s.Snooze(job, until)
			w.WriteHeader(http.StatusOK)
			json.NewEncoder(w).Encode(snoozeRecord{Job: job, Until: until})

		case http.MethodGet:
			until, ok := s.SnoozedUntil(job)
			if !ok {
				w.WriteHeader(http.StatusNotFound)
				json.NewEncoder(w).Encode(map[string]string{"status": "not snoozed"})
				return
			}
			json.NewEncoder(w).Encode(snoozeRecord{Job: job, Until: until})

		case http.MethodDelete:
			s.Unsnooze(job)
			w.WriteHeader(http.StatusNoContent)

		default:
			http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
		}
	})
}
