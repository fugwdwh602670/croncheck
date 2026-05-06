// Package inhibit implements dependency-based alert inhibition for croncheck.
//
// An inhibition rule suppresses alerts for a target job when a source job
// is currently unhealthy (missed or failed). This prevents alert storms
// caused by a single upstream failure cascading into many dependent jobs.
//
// Example: if "db-backup" is missed, alerts for "db-report" and "db-cleanup"
// can be inhibited until "db-backup" recovers.
package inhibit
