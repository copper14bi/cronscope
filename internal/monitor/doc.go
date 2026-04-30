// Package monitor implements job lifecycle tracking for cronscope.
//
// A Monitor is created from a parsed [config.Config] and a [webhook.Client].
// Callers record job lifecycle events via RecordStart and RecordFinish, then
// invoke Check periodically (e.g. every minute) to evaluate whether any jobs
// have been missed or have exceeded their configured maximum duration.
//
// Alerts are delivered as webhook payloads with the following event types:
//
//   - "missed"       — a job did not run within its expected schedule window.
//   - "long_running" — a job has been running longer than its MaxDuration.
//
// Monitor is safe for concurrent use.
package monitor
