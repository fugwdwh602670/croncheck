// Package tags provides job tagging support for croncheck.
//
// Tags are arbitrary key-value string pairs attached to a job name.
// They can be used to group jobs by environment, team, or any other
// operator-defined dimension, and to filter API responses.
//
// The HTTP handler exposes:
//
//	GET  /tags              — list all job tags
//	GET  /tags?job=name     — get tags for a specific job
//	GET  /tags?filter=k:v   — list jobs matching tag query
//	PUT  /tags?job=name     — set tags for a job (JSON body)
//	DELETE /tags?job=name   — remove all tags for a job
package tags
