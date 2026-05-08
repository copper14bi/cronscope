// Package retryapi provides an HTTP handler for manually triggering alert
// retries for a specific cron job.
//
// # Endpoint
//
//	POST /retry?job=<name>
//
// The handler delegates to the alertmanager.Manager.Retry method, which
// bypasses the cooldown window and immediately re-dispatches the most recent
// alert for the named job.
//
// Requests are rejected with 409 Conflict when the job is currently
// suppressed, preventing unintended noise during maintenance windows.
package retryapi
