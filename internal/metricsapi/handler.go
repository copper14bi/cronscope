// Package metricsapi exposes a simple JSON metrics endpoint summarising
// runtime counters for cronscope: total alerts fired, jobs tracked, and
// per-job missed/long-running counts derived from the history store.
package metricsapi

import (
	"encoding/json"
	"net/http"
	"time"

	"github.com/example/cronscope/internal/history"
	"github.com/example/cronscope/internal/state"
)

// JobMetrics holds counters for a single tracked job.
type JobMetrics struct {
	Missed      int `json:"missed"`
	LongRunning int `json:"long_running"`
}

// Response is the top-level payload returned by the metrics endpoint.
type Response struct {
	GeneratedAt  time.Time             `json:"generated_at"`
	TotalJobs    int                   `json:"total_jobs"`
	TotalAlerts  int                   `json:"total_alerts"`
	Jobs         map[string]JobMetrics `json:"jobs"`
}

// Handler returns an http.Handler that serves the metrics JSON response.
func Handler(st *state.Store, hs *history.Store) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodGet {
			http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
			return
		}

		jobNames := st.Names()
		jobMetrics := make(map[string]JobMetrics, len(jobNames))
		totalAlerts := 0

		for _, name := range jobNames {
			events := hs.ForJob(name)
			var missed, long int
			for _, e := range events {
				switch e.Kind {
				case history.KindMissed:
					missed++
				case history.KindLongRunning:
					long++
				}
			}
			totalAlerts += missed + long
			jobMetrics[name] = JobMetrics{Missed: missed, LongRunning: long}
		}

		resp := Response{
			GeneratedAt: time.Now().UTC(),
			TotalJobs:   len(jobNames),
			TotalAlerts: totalAlerts,
			Jobs:        jobMetrics,
		}

		w.Header().Set("Content-Type", "application/json")
		_ = json.NewEncoder(w).Encode(resp)
	})
}
