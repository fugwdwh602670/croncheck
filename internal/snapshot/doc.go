// Package snapshot provides point-in-time capture of all monitored job states.
//
// A snapshot records the health, missed count, and last-seen time for every
// known job at the moment of capture. Snapshots are stored in a bounded
// in-memory ring buffer and can be retrieved by ID or listed newest-first.
//
// Typical usage:
//
//	store := snapshot.New(20)
//	snap := store.Capture("manual-2024-06-01", jobSource)
package snapshot
