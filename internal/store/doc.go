// Package store provides an in-memory thread-safe store for tracking
// cron job heartbeats and missed-run state.
//
// Each job entry records the last time a heartbeat was received,
// how many consecutive missed runs have occurred, and whether the
// job is currently considered healthy.
//
// The store is safe for concurrent use by multiple goroutines.
package store
