package schedule_test

import (
	"testing"
	"time"

	"github.com/cronscope/cronscope/internal/schedule"
)

func mustParse(s string) time.Time {
	t, err := time.Parse(time.RFC3339, s)
	if err != nil {
		panic(err)
	}
	return t
}

func TestNextTime_Valid(t *testing.T) {
	// Every hour at minute 0
	from := mustParse("2024-01-15T10:00:00Z")
	next, err := schedule.NextTime("0 * * * *", from)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	expected := mustParse("2024-01-15T11:00:00Z")
	if !next.Equal(expected) {
		t.Errorf("expected %v, got %v", expected, next)
	}
}

func TestNextTime_InvalidExpr(t *testing.T) {
	_, err := schedule.NextTime("not-a-cron", time.Now())
	if err == nil {
		t.Fatal("expected error for invalid cron expression")
	}
}

func TestIsMissed_True(t *testing.T) {
	lastSeen := mustParse("2024-01-15T10:00:00Z")
	// Next run would be 11:00; grace is 5 min; deadline is 11:05
	now := mustParse("2024-01-15T11:10:00Z")
	missed, err := schedule.IsMissed("0 * * * *", lastSeen, now, 5*time.Minute)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !missed {
		t.Error("expected job to be missed")
	}
}

func TestIsMissed_False(t *testing.T) {
	lastSeen := mustParse("2024-01-15T10:00:00Z")
	// Next run would be 11:00; grace is 5 min; now is 11:03 — within grace
	now := mustParse("2024-01-15T11:03:00Z")
	missed, err := schedule.IsMissed("0 * * * *", lastSeen, now, 5*time.Minute)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if missed {
		t.Error("expected job NOT to be missed")
	}
}

func TestIsLongRunning_True(t *testing.T) {
	started := mustParse("2024-01-15T10:00:00Z")
	now := mustParse("2024-01-15T10:35:00Z")
	if !schedule.IsLongRunning(started, now, 30*time.Minute) {
		t.Error("expected job to be long-running")
	}
}

func TestIsLongRunning_False(t *testing.T) {
	started := mustParse("2024-01-15T10:00:00Z")
	now := mustParse("2024-01-15T10:20:00Z")
	if schedule.IsLongRunning(started, now, 30*time.Minute) {
		t.Error("expected job NOT to be long-running")
	}
}

func TestIsLongRunning_ZeroStart(t *testing.T) {
	if schedule.IsLongRunning(time.Time{}, time.Now(), time.Minute) {
		t.Error("zero startedAt should never be long-running")
	}
}
