// Package catalog provides a job metadata registry for croncheck.
//
// Each job can be annotated with a human-readable description, an owner
// identifier, and the cron schedule expression it is expected to follow.
// The catalog is independent of runtime state (heartbeats, missed counts)
// and serves as a static reference for dashboards and alerting context.
//
// Usage:
//
//	s := catalog.New()
//	_ = s.Set(catalog.Entry{Job: "backup", Owner: "ops", Schedule: "0 3 * * *"})
//	e, ok := s.Get("backup")
package catalog
