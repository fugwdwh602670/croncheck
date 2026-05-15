// Package expiry provides per-job expiry deadline tracking for croncheck.
//
// A deadline is set for a job using Set with a TTL duration. Once the
// deadline passes, IsExpired returns true, allowing the watcher or other
// components to take action (e.g. fire an alert or move the job to the
// dead-letter queue).
//
// The HTTP handler exposes GET, PUT, and DELETE endpoints:
//
//	GET  /expiry?job=<name>  – retrieve deadline for a specific job
//	GET  /expiry             – list all registered deadlines
//	PUT  /expiry?job=<name>  – set or refresh a deadline (body: {"ttl":"10m"})
//	DELETE /expiry?job=<name> – remove a deadline
package expiry
