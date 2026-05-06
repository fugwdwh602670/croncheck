// Package heartbeat provides an HTTP handler that cron jobs can call to
// signal successful execution. A POST to /heartbeat/{job} records a
// timestamped heartbeat in the store, which the watcher uses to detect
// missed or overdue runs.
//
// Usage:
//
//	mux.Handle("/heartbeat/", heartbeat.Handler(store))
package heartbeat
