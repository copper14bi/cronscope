package history_test

import (
	"testing"
	"time"

	"github.com/yourorg/cronscope/internal/history"
)

func makeEvent(job string, t history.EventType) history.Event {
	return history.Event{
		JobName:    job,
		Type:       t,
		OccurredAt: time.Now().UTC(),
		Message:    "test",
	}
}

func TestRecord_StoresEvent(t *testing.T) {
	s := history.New(10)
	s.Record(makeEvent("backup", history.EventMissed))
	if s.Len() != 1 {
		t.Fatalf("expected 1 event, got %d", s.Len())
	}
}

func TestRecord_EvictsOldestWhenFull(t *testing.T) {
	s := history.New(3)
	s.Record(makeEvent("a", history.EventMissed))
	s.Record(makeEvent("b", history.EventMissed))
	s.Record(makeEvent("c", history.EventMissed))
	s.Record(makeEvent("d", history.EventMissed))

	events := s.All()
	if len(events) != 3 {
		t.Fatalf("expected 3 events after eviction, got %d", len(events))
	}
	if events[0].JobName != "b" {
		t.Errorf("expected oldest surviving event to be 'b', got %q", events[0].JobName)
	}
}

func TestForJob_FiltersCorrectly(t *testing.T) {
	s := history.New(20)
	s.Record(makeEvent("alpha", history.EventMissed))
	s.Record(makeEvent("beta", history.EventLongRunning))
	s.Record(makeEvent("alpha", history.EventRecovered))

	results := s.ForJob("alpha")
	if len(results) != 2 {
		t.Fatalf("expected 2 events for 'alpha', got %d", len(results))
	}
	for _, e := range results {
		if e.JobName != "alpha" {
			t.Errorf("unexpected job name %q in filtered results", e.JobName)
		}
	}
}

func TestRecord_SetsTimestampIfZero(t *testing.T) {
	s := history.New(10)
	s.Record(history.Event{JobName: "job", Type: history.EventMissed})
	events := s.All()
	if events[0].OccurredAt.IsZero() {
		t.Error("expected OccurredAt to be set automatically")
	}
}

func TestAll_ReturnsCopy(t *testing.T) {
	s := history.New(10)
	s.Record(makeEvent("x", history.EventMissed))
	a := s.All()
	a[0].JobName = "mutated"
	b := s.All()
	if b[0].JobName == "mutated" {
		t.Error("All() should return a copy, not a reference to internal slice")
	}
}
