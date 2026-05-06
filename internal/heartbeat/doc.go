// Package heartbeat provides an HTTP handler that cron jobs can call to
// signal successful execution. A POST to /heartbeat/{job} records a
// timestamped heartbeat in the store, which the watcher uses to detect
// missed or overdue runs.
//
// Each heartbeat request should include the job name as a path segment.
// The handler responds with 200 OK on success, 405 Method Not Allowed if
// the request method is not POST, and 400 Bad Request if the job name is
// missing or empty.
//
// Usage:
//
//	mux.Handle("/heartbeat/", heartbeat.Handler(store))
//
// Example cron job (curl):
//
//	*/5 * * * * curl -s -X POST https://example.com/heartbeat/my-job
package heartbeat
