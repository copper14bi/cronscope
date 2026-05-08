// Package silenceapi provides an HTTP handler for managing per-job alert silences.
// A silence suppresses all alerts for a specific job until it expires.
package silenceapi

import (
	"encoding/json"
	"net/http"
	"strings"
	"time"
)

// Silencer is the interface for managing job-level silences.
type Silencer interface {
	Silence(job string, until time.Time)
	Unsilence(job string)
	IsSilenced(job string) (bool, time.Time)
}

type silenceRequest struct {
	Duration string `json:"duration"`
}

type silenceResponse struct {
	Job      string    `json:"job"`
	Silenced bool      `json:"silenced"`
	Until    time.Time `json:"until,omitempty"`
}

// Handler returns an HTTP handler for GET, POST, and DELETE on /silence/{job}.
func Handler(s Silencer) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		job := strings.TrimPrefix(r.URL.Path, "/")
		if job == "" {
			http.Error(w, "job name required", http.StatusBadRequest)
			return
		}

		w.Header().Set("Content-Type", "application/json")

		switch r.Method {
		case http.MethodGet:
			silenced, until := s.IsSilenced(job)
			json.NewEncoder(w).Encode(silenceResponse{Job: job, Silenced: silenced, Until: until})

		case http.MethodPost:
			var req silenceRequest
			if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
				http.Error(w, "invalid JSON", http.StatusBadRequest)
				return
			}
			d, err := time.ParseDuration(req.Duration)
			if err != nil || d <= 0 {
				http.Error(w, "invalid duration", http.StatusBadRequest)
				return
			}
			until := time.Now().Add(d)
			s.Silence(job, until)
			w.WriteHeader(http.StatusCreated)
			json.NewEncoder(w).Encode(silenceResponse{Job: job, Silenced: true, Until: until})

		case http.MethodDelete:
			s.Unsilence(job)
			json.NewEncoder(w).Encode(silenceResponse{Job: job, Silenced: false})

		default:
			http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
		}
	})
}
