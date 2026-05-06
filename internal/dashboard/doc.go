// Package dashboard exposes a single HTTP endpoint (/dashboard) that returns
// a JSON summary of all monitored cron jobs, including aggregate counts of
// healthy, unhealthy, and never-seen jobs.
//
// It is intentionally read-only and imposes no side-effects on the store.
package dashboard
