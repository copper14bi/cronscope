// Package history provides an in-memory ring buffer for recording
// job lifecycle events such as missed runs, long-running detections,
// and recoveries.
//
// Events are stored in insertion order and the buffer automatically
// evicts the oldest entry when its capacity is reached. All operations
// are safe for concurrent use.
//
// Typical usage:
//
//	store := history.New(200)
//	store.Record(history.Event{
//		JobName: "nightly-backup",
//		Type:    history.EventMissed,
//		Message: "job did not run within grace period",
//	})
//
//	events := store.ForJob("nightly-backup")
package history
