// Package runner provides the main polling loop that periodically checks
// all registered cron jobs against their expected schedules.
package runner

import (
	"context"
	"log"
	"time"

	"github.com/cronscope/cronscope/internal/config"
	"github.com/cronscope/cronscope/internal/monitor"
	"github.com/cronscope/cronscope/internal/state"
)

// Runner drives the periodic check loop for all configured jobs.
type Runner struct {
	cfg      *config.Config
	mon      *monitor.Monitor
	state    *state.State
	interval time.Duration
}

// New creates a Runner with the given config, monitor, state store, and poll interval.
func New(cfg *config.Config, mon *monitor.Monitor, st *state.State, interval time.Duration) *Runner {
	return &Runner{
		cfg:      cfg,
		mon:      mon,
		state:    st,
		interval: interval,
	}
}

// Run starts the polling loop and blocks until ctx is cancelled.
func (r *Runner) Run(ctx context.Context) {
	ticker := time.NewTicker(r.interval)
	defer ticker.Stop()

	log.Printf("runner: starting poll loop (interval=%s, jobs=%d)", r.interval, len(r.cfg.Jobs))

	// Run an immediate check on startup before waiting for the first tick.
	r.checkAll()

	for {
		select {
		case <-ticker.C:
			r.checkAll()
		case <-ctx.Done():
			log.Println("runner: shutting down")
			return
		}
	}
}

func (r *Runner) checkAll() {
	now := time.Now()
	for _, job := range r.cfg.Jobs {
		if err := r.mon.Check(job, now); err != nil {
			log.Printf("runner: check failed for job %q: %v", job.Name, err)
		}
	}
}
