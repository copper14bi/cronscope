// Package tagapi provides an HTTP handler for managing job tags,
// allowing operators to annotate cron jobs with arbitrary key-value labels
// for grouping, filtering, and downstream routing.
package tagapi

import (
	"encoding/json"
	"net/http"
	"strings"
	"sync"
)

// Store holds tags for named jobs.
type Store struct {
	mu   sync.RWMutex
	tags map[string]map[string]string // job name -> key -> value
}

// NewStore returns an initialised tag Store.
func NewStore() *Store {
	return &Store{tags: make(map[string]map[string]string)}
}

// Set replaces all tags for the given job.
func (s *Store) Set(job string, tags map[string]string) {
	s.mu.Lock()
	defer s.mu.Unlock()
	copy := make(map[string]string, len(tags))
	for k, v := range tags {
		copy[k] = v
	}
	s.tags[job] = copy
}

// Get returns the tags for the given job and whether the job exists.
func (s *Store) Get(job string) (map[string]string, bool) {
	s.mu.RLock()
	defer s.mu.RUnlock()
	t, ok := s.tags[job]
	if !ok {
		return nil, false
	}
	copy := make(map[string]string, len(t))
	for k, v := range t {
		copy[k] = v
	}
	return copy, true
}

// Delete removes all tags for the given job.
func (s *Store) Delete(job string) {
	s.mu.Lock()
	defer s.mu.Unlock()
	delete(s.tags, job)
}

// Handler returns an http.Handler for GET / PUT / DELETE /tags/{job}.
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
			tags, ok := store.Get(job)
			if !ok {
				tags = map[string]string{}
			}
			json.NewEncoder(w).Encode(tags)

		case http.MethodPut:
			var incoming map[string]string
			if err := json.NewDecoder(r.Body).Decode(&incoming); err != nil {
				http.Error(w, "invalid JSON", http.StatusBadRequest)
				return
			}
			store.Set(job, incoming)
			w.WriteHeader(http.StatusNoContent)

		case http.MethodDelete:
			store.Delete(job)
			w.WriteHeader(http.StatusNoContent)

		default:
			http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
		}
	})
}
