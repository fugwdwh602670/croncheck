// Package store provides an in-memory, thread-safe store for tracking
// the execution state of monitored cron jobs.
//
// Each job is identified by its name (as defined in the croncheck config).
// The store records the last heartbeat time, the number of consecutive missed
// runs, and whether an alert has already been sent for the current missed window.
//
// Typical usage:
//
//	s := store.New()
//
//	// When a heartbeat arrives for a job:
//	s.RecordHeartbeat(jobName)
//
//	// When the scheduler detects a missed run:
//	s.IncrementMissed(jobName)
//
//	// To inspect current state:
//	status, ok := s.Get(jobName)
package store
