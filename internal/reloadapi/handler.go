// Package reloadapi provides an HTTP handler for triggering a live
// configuration reload without restarting the daemon.
package reloadapi

import (
	"encoding/json"
	"log"
	"net/http"
	"time"
)

// Reloader is implemented by any component that can reload its configuration
// at runtime (e.g. the top-level application).
type Reloader interface {
	Reload() error
}

type response struct {
	OK        bool      `json:"ok"`
	Message   string    `json:"message"`
	ReloadedAt time.Time `json:"reloaded_at,omitempty"`
}

// Handler returns an http.HandlerFunc that accepts POST requests and delegates
// to the provided Reloader. Any other HTTP method returns 405.
func Handler(r Reloader) http.HandlerFunc {
	return func(w http.ResponseWriter, req *http.Request) {
		if req.Method != http.MethodPost {
			http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
			return
		}

		w.Header().Set("Content-Type", "application/json")

		if err := r.Reload(); err != nil {
			log.Printf("reloadapi: reload failed: %v", err)
			w.WriteHeader(http.StatusInternalServerError)
			_ = json.NewEncoder(w).Encode(response{
				OK:      false,
				Message: err.Error(),
			})
			return
		}

		log.Printf("reloadapi: configuration reloaded successfully")
		w.WriteHeader(http.StatusOK)
		_ = json.NewEncoder(w).Encode(response{
			OK:         true,
			Message:    "configuration reloaded",
			ReloadedAt: time.Now().UTC(),
		})
	}
}
