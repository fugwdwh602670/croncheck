// Package sla tracks SLA (Service Level Agreement) compliance for monitored
// cron jobs. It records each scheduled check as either a hit (job ran on time)
// or a miss, and exposes per-job statistics including total checks, hits,
// misses, and the computed compliance percentage.
//
// Use HTTPHandler to expose SLA data over HTTP for dashboards or alerting
// integrations.
package sla
