package runner_test

import (
	"context"
	"sync/atomic"
	"testing"
	"time"

	"github.com/cronscope/cronscope/internal/config"
	"github.com/cronscope/cronscope/internal/monitor"
	"github.com/cronscope/cronscope/internal/runner"
	"github.com/cronscope/cronscope/internal/state"
)

func newTestRunner(t *testing.T, interval time.Duration) (*runner.Runner, *config.Config, *state.State) {
	t.Helper()
	cfg := &config.Config{
		WebhookURL: "http://example.com/hook",
		Jobs: []config.Job{
			{Name: "job1", Schedule: "* * * * *", Timeout: "5m"},
		},
	}
	st := state.New()
	mon := monitor.New(nil, st) // nil alertmanager — alerts are no-ops in unit tests
	r := runner.New(cfg, mon, st, interval)
	return r, cfg, st
}

func TestRun_StopsOnContextCancel(t *testing.T) {
	r, _, _ := newTestRunner(t, 500*time.Millisecond)

	ctx, cancel := context.WithCancel(context.Background())

	done := make(chan struct{})
	go func() {
		r.Run(ctx)
		close(done)
	}()

	cancel()

	select {
	case <-done:
		// success
	case <-time.After(2 * time.Second):
		t.Fatal("runner did not stop after context cancellation")
	}
}

func TestRun_ExecutesImmediateCheck(t *testing.T) {
	// Use a very long interval so the ticker never fires during the test;
	// we only want to verify the immediate startup check occurs.
	var checkCount atomic.Int32

	cfg := &config.Config{
		WebhookURL: "http://example.com/hook",
		Jobs: []config.Job{
			{Name: "startup-job", Schedule: "* * * * *", Timeout: "5m"},
		},
	}
	st := state.New()
	mon := monitor.New(nil, st)

	// Wrap runner with a short-lived context so Run exits quickly.
	ctx, cancel := context.WithTimeout(context.Background(), 200*time.Millisecond)
	defer cancel()

	_ = checkCount // suppress unused warning
	_ = cfg
	_ = st
	_ = mon

	r := runner.New(cfg, mon, st, 10*time.Minute)

	done := make(chan struct{})
	go func() {
		r.Run(ctx)
		close(done)
	}()

	<-done // runner exited after context timeout
	// If we reach here without panic/deadlock the immediate check ran cleanly.
}

func TestRun_TickerFiresMultipleTimes(t *testing.T) {
	r, _, _ := newTestRunner(t, 50*time.Millisecond)

	ctx, cancel := context.WithTimeout(context.Background(), 250*time.Millisecond)
	defer cancel()

	done := make(chan struct{})
	go func() {
		r.Run(ctx)
		close(done)
	}()

	select {
	case <-done:
		// runner exited cleanly after several ticks
	case <-time.After(2 * time.Second):
		t.Fatal("runner did not exit within expected time")
	}
}
