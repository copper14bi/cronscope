// Package labelapi provides HTTP handlers for managing per-job labels.
// Labels are arbitrary key/value string pairs that can be attached to jobs
// for grouping, routing, or display purposes.
package labelapi

import (
	"encoding/json"
	"net/http"
	"strings"
	"sync"
)

// Store holds labels for each job name.
type Store struct {
	mu     sync.RWMutex
	labels map[string]map[string]string
}

// NewStore returns an initialised label Store.
func NewStore() *Store {
	return &Store{labels: make(map[string]map[string]string)}
}

// Set replaces all labels for a job.
func (s *Store) Set(job string, labels map[string]string) {
	s.mu.Lock()
	defer s.mu.Unlock()
	copy := make(map[string]string, len(labels))
	for k, v := range labels {
		copy[k] = v
	}
	s.labels[job] = copy
}

// Get returns the labels for a job, or an empty map if none exist.
func (s *Store) Get(job string) map[string]string {
	s.mu.RLock()
	defer s.mu.RUnlock()
	if m, ok := s.labels[job]; ok {
		copy := make(map[string]string, len(m))
		for k, v := range m {
			copy[k] = v
		}
		return copy
	}
	return map[string]string{}
}

// Delete removes all labels for a job.
func (s *Store) Delete(job string) {
	s.mu.Lock()
	defer s.mu.Unlock()
	delete(s.labels, job)
}

// Handler returns an http.Handler that serves GET, PUT, and DELETE for
// /labels/{job}.
func Handler(s *Store) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		job := strings.TrimPrefix(r.URL.Path, "/")
		if job == "" {
			http.Error(w, "job name required", http.StatusBadRequest)
			return
		}

		w.Header().Set("Content-Type", "application/json")

		switch r.Method {
		case http.MethodGet:
			json.NewEncoder(w).Encode(s.Get(job))

		case http.MethodPut:
			var labels map[string]string
			if err := json.NewDecoder(r.Body).Decode(&labels); err != nil {
				http.Error(w, "invalid JSON", http.StatusBadRequest)
				return
			}
			s.Set(job, labels)
			w.WriteHeader(http.StatusNoContent)

		case http.MethodDelete:
			s.Delete(job)
			w.WriteHeader(http.StatusNoContent)

		default:
			http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
		}
	})
}
