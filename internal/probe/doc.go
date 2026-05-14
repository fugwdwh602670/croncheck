// Package probe implements a liveness probe store for croncheck jobs.
//
// Each job may register a check-in with an associated TTL. If the job
// does not check in again before the TTL expires, its status transitions
// from "alive" to "dead". Jobs that have never checked in are reported
// as "unknown".
//
// The HTTP handler exposes:
//
//	GET  /probe          — list all probe entries
//	GET  /probe?job=name — get probe status for a specific job
//	POST /probe?job=name&ttl=5m — record a probe check-in
package probe
