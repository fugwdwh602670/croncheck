// Package grouping allows operators to assign cron jobs to named groups
// (e.g. "ops", "finance", "infra") so that dashboards, alerts, and API
// responses can be filtered or aggregated at the group level.
//
// The grouping store is safe for concurrent use. Group assignments are
// held in memory and are not persisted across restarts.
package grouping
