package state_test

import (
	"sort"
	"testing"
	"time"

	"github.com/yourorg/cronscope/internal/state"
)

var epoch = time.Date(2024, 1, 1, 12, 0, 0, 0, time.UTC)

func TestMarkSeen_CreatesEntry(t *testing.T) {
	s := state.New()
	s.MarkSeen("backup", epoch)

	js, ok := s.Get("backup")
	if !ok {
		t.Fatal("expected entry for 'backup'")
	}
	if !js.LastSeen.Equal(epoch) {
		t.Errorf("LastSeen = %v, want %v", js.LastSeen, epoch)
	}
	if js.Running {
		t.Error("expected Running to be false after MarkSeen")
	}
}

func TestMarkRunning_SetsRunningState(t *testing.T) {
	s := state.New()
	s.MarkRunning("report", epoch)

	js, ok := s.Get("report")
	if !ok {
		t.Fatal("expected entry for 'report'")
	}
	if !js.Running {
		t.Error("expected Running to be true")
	}
	if !js.StartedAt.Equal(epoch) {
		t.Errorf("StartedAt = %v, want %v", js.StartedAt, epoch)
	}
}

func TestMarkSeen_ClearsRunningState(t *testing.T) {
	s := state.New()
	s.MarkRunning("report", epoch)
	s.MarkSeen("report", epoch.Add(time.Minute))

	js, _ := s.Get("report")
	if js.Running {
		t.Error("expected Running to be cleared after MarkSeen")
	}
	if !js.StartedAt.IsZero() {
		t.Error("expected StartedAt to be zero after MarkSeen")
	}
}

func TestGet_MissingJob(t *testing.T) {
	s := state.New()
	_, ok := s.Get("nonexistent")
	if ok {
		t.Error("expected ok=false for unknown job")
	}
}

func TestNames_ReturnsAllJobs(t *testing.T) {
	s := state.New()
	s.MarkSeen("alpha", epoch)
	s.MarkSeen("beta", epoch)
	s.MarkRunning("gamma", epoch)

	names := s.Names()
	sort.Strings(names)

	want := []string{"alpha", "beta", "gamma"}
	if len(names) != len(want) {
		t.Fatalf("Names() = %v, want %v", names, want)
	}
	for i, n := range names {
		if n != want[i] {
			t.Errorf("names[%d] = %q, want %q", i, n, want[i])
		}
	}
}
