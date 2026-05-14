// Package commentapi provides an HTTP handler for managing per-job comments.
// Comments are free-form text annotations that operators can attach to a job
// to provide context (e.g. "disabled during maintenance window").
package commentapi

import (
	"encoding/json"
	"net/http"
	"strings"
	"sync"
)

// Store holds comments keyed by job name.
type Store struct {
	mu       sync.RWMutex
	comments map[string]string
}

// NewStore returns an initialised Store.
func NewStore() *Store {
	return &Store{comments: make(map[string]string)}
}

// Set stores a comment for the given job, replacing any existing one.
func (s *Store) Set(job, comment string) {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.comments[job] = comment
}

// Get returns the comment for a job and whether one exists.
func (s *Store) Get(job string) (string, bool) {
	s.mu.RLock()
	defer s.mu.RUnlock()
	v, ok := s.comments[job]
	return v, ok
}

// Delete removes the comment for a job.
func (s *Store) Delete(job string) {
	s.mu.Lock()
	defer s.mu.Unlock()
	delete(s.comments, job)
}

// Handler returns an http.Handler that manages comments for a single job.
// Routes:
//
//	GET    /comments/{job}  – retrieve comment
//	PUT    /comments/{job}  – set comment   (body: {"comment":"..."} )
//	DELETE /comments/{job}  – remove comment
func Handler(store *Store) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		job := strings.TrimPrefix(r.URL.Path, "/")
		if job == "" {
			http.Error(w, "job name required", http.StatusBadRequest)
			return
		}

		w.Header().Set("Content-Type", "application/json")

		switch r.Method {
		case http.MethodGet:
			comment, _ := store.Get(job)
			json.NewEncoder(w).Encode(map[string]string{"job": job, "comment": comment})

		case http.MethodPut:
			var body struct {
				Comment string `json:"comment"`
			}
			if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
				http.Error(w, "invalid JSON", http.StatusBadRequest)
				return
			}
			store.Set(job, body.Comment)
			w.WriteHeader(http.StatusNoContent)

		case http.MethodDelete:
			store.Delete(job)
			w.WriteHeader(http.StatusNoContent)

		default:
			http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
		}
	})
}
