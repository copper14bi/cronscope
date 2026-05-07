// Package suppressapi provides an HTTP handler for managing alert suppression rules.
// Suppression allows operators to silence alerts for specific jobs during maintenance windows.
package suppressapi

import (
	"encoding/json"
	"net/http"
	"time"

	"github.com/cronscope/cronscope/internal/alertmanager"
)

// Handler returns an http.Handler for GET/POST/DELETE on /suppress.
// GET  returns all active suppression rules.
// POST creates a new suppression rule (body: {"job":"name","duration":"2h"}).
// DELETE removes a suppression rule (body: {"job":"name"}).
func Handler(am *alertmanager.Manager) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")

		switch r.Method {
		case http.MethodGet:
			rules := am.Suppressions()
			if err := json.NewEncoder(w).Encode(rules); err != nil {
				http.Error(w, `{"error":"encode failed"}`, http.StatusInternalServerError)
			}

		case http.MethodPost:
			var req struct {
				Job      string `json:"job"`
				Duration string `json:"duration"`
			}
			if err := json.NewDecoder(r.Body).Decode(&req); err != nil || req.Job == "" || req.Duration == "" {
				http.Error(w, `{"error":"invalid request body"}`, http.StatusBadRequest)
				return
			}
			d, err := time.ParseDuration(req.Duration)
			if err != nil {
				http.Error(w, `{"error":"invalid duration"}`, http.StatusBadRequest)
				return
			}
			am.Suppress(req.Job, d)
			w.WriteHeader(http.StatusCreated)
			_ = json.NewEncoder(w).Encode(map[string]string{"status": "suppressed", "job": req.Job})

		case http.MethodDelete:
			var req struct {
				Job string `json:"job"`
			}
			if err := json.NewDecoder(r.Body).Decode(&req); err != nil || req.Job == "" {
				http.Error(w, `{"error":"invalid request body"}`, http.StatusBadRequest)
				return
			}
			am.Unsuppress(req.Job)
			_ = json.NewEncoder(w).Encode(map[string]string{"status": "unsuppressed", "job": req.Job})

		default:
			http.Error(w, `{"error":"method not allowed"}`, http.StatusMethodNotAllowed)
		}
	})
}
