// Package schedule provides helpers for evaluating cron job timing.
//
// It exposes three core functions used by the cronscope monitor:
//
//   - NextTime: parses a standard 5-field cron expression and returns the next
//     scheduled time after a given reference point.
//
//   - IsMissed: determines whether a job has missed its expected execution
//     window, taking an optional grace period into account.
//
//   - IsLongRunning: determines whether a currently-running job has exceeded
//     its configured maximum duration.
//
// All cron expressions follow the standard POSIX format:
//
//	minute hour day-of-month month day-of-week
//
// Example:
//
//	"0 * * * *"   // every hour on the hour
//	"*/5 * * * *" // every five minutes
//	"0 9 * * 1"   // every Monday at 09:00
package schedule
