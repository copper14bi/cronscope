// Package metricsapi provides an HTTP handler that exposes runtime metrics
// for the cronscope daemon in JSON format.
//
// The /metrics endpoint (GET only) returns:
//
//   - generated_at  – RFC 3339 timestamp of when the response was built.
//   - total_jobs    – number of jobs currently tracked in the state store.
//   - total_alerts  – cumulative missed + long-running events across all jobs.
//   - jobs          – per-job breakdown with individual missed and
//                     long_running counters sourced from the history store.
//
// The handler is intentionally read-only and stateless; it computes all
// values on each request directly from the provided Store references.
package metricsapi
