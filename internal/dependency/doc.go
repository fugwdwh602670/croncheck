// Package dependency provides a thread-safe store for job dependency
// relationships within croncheck.
//
// A dependency declares that a job relies on one or more upstream jobs being
// healthy before its own missed/failed alerts are considered actionable.
// The BlockedBy helper integrates with any health-check function so that the
// watcher or notifier can suppress alerts when a root-cause upstream is already
// known to be unhealthy.
package dependency
