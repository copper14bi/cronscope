package schedule

import (
	"fmt"
	"time"

	"github.com/robfig/cron/v3"
)

// NextTime returns the next scheduled time after 'from' for the given cron expression.
func NextTime(expr string, from time.Time) (time.Time, error) {
	p := cron.NewParser(cron.Minute | cron.Hour | cron.Dom | cron.Month | cron.Dow)
	sched, err := p.Parse(expr)
	if err != nil {
		return time.Time{}, fmt.Errorf("invalid cron expression %q: %w", expr, err)
	}
	return sched.Next(from), nil
}

// IsMissed reports whether a job with the given schedule has missed its window.
// A job is considered missed if the next expected run after 'lastSeen' is more
// than 'gracePeriod' in the past relative to 'now'.
func IsMissed(expr string, lastSeen time.Time, now time.Time, gracePeriod time.Duration) (bool, error) {
	next, err := NextTime(expr, lastSeen)
	if err != nil {
		return false, err
	}
	deadline := next.Add(gracePeriod)
	return now.After(deadline), nil
}

// IsLongRunning reports whether a job that started at 'startedAt' has exceeded
// its maximum allowed duration relative to 'now'.
func IsLongRunning(startedAt time.Time, now time.Time, maxDuration time.Duration) bool {
	if startedAt.IsZero() {
		return false
	}
	return now.Sub(startedAt) > maxDuration
}
