// Package retention provides automatic cleanup of stale job state.
//
// A Cleaner runs on a configurable interval and removes jobs from the
// store that have not sent a heartbeat within a given TTL. This prevents
// unbounded memory growth when ephemeral or renamed jobs stop reporting.
//
// Usage:
//
//	cleaner := retention.New(store, 24*time.Hour, 1*time.Hour)
//	go cleaner.Run(ctx)
package retention
