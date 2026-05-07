// Package pauseapi provides HTTP handlers for pausing and resuming
// individual cron job monitoring without removing them from the config.
package pauseapi

import (
	"encoding/json"
	"net/http"
	"strings"
	"time"
)

// Pauser defines the interface for pausing and resuming job monitoring.
type Pauser interface {
	Pause(jobName string, until time.Time)
	Resume(jobName string)
	IsPaused(jobName string) bool
	PausedUntil(jobName string) (time.Time, bool)
}

type pauseRequest struct {
	Duration string `json:"duration"`
}

type pauseResponse struct {
	Job      string `json:"job"`
	Paused   bool   `json:"paused"`
	Until    string `json:"until,omitempty"`
}

// Handler returns an HTTP handler for the pause API.
// POST /pause/{job}?duration=1h  — pause a job for the given duration
// DELETE /pause/{job}            — resume a paused job
// GET /pause/{job}               — check pause status
func Handler(p Pauser) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		jobName := strings.TrimPrefix(r.URL.Path, "/")
		if jobName == "" {
			http.Error(w, "job name required", http.StatusBadRequest)
			return
		}

		w.Header().Set("Content-Type", "application/json")

		switch r.Method {
		case http.MethodPost:
			durStr := r.URL.Query().Get("duration")
			if durStr == "" {
				http.Error(w, `{"error":"duration query param required"}`, http.StatusBadRequest)
				return
			}
			dur, err := time.ParseDuration(durStr)
			if err != nil || dur <= 0 {
				http.Error(w, `{"error":"invalid duration"}`, http.StatusBadRequest)
				return
			}
			until := time.Now().Add(dur)
			p.Pause(jobName, until)
			json.NewEncoder(w).Encode(pauseResponse{
				Job:    jobName,
				Paused: true,
				Until:  until.UTC().Format(time.RFC3339),
			})

		case http.MethodDelete:
			p.Resume(jobName)
			json.NewEncoder(w).Encode(pauseResponse{Job: jobName, Paused: false})

		case http.MethodGet:
			until, ok := p.PausedUntil(jobName)
			resp := pauseResponse{Job: jobName, Paused: ok}
			if ok {
				resp.Until = until.UTC().Format(time.RFC3339)
			}
			json.NewEncoder(w).Encode(resp)

		default:
			http.Error(w, `{"error":"method not allowed"}`, http.StatusMethodNotAllowed)
		}
	})
}
