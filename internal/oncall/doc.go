// Package oncall manages on-call schedules for cron jobs.
// Each job can have an active on-call entry that specifies who is
// responsible for responding to alerts during a given time window.
//
// Entries expire automatically when their end time is reached.
// The HTTPHandler exposes GET, PUT, and DELETE endpoints for managing
// on-call assignments per job.
package oncall
