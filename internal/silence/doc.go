// Package silence implements a thread-safe registry for suppressing cron-job
// alerts during planned maintenance windows.
//
// A Silence is keyed by job name and carries an expiry time. The Registry
// exposes Add, Remove, IsSilenced, and All methods for programmatic use, and
// HTTPHandler provides a REST API for managing silences at runtime:
//
//	GET    /silences          – list all currently active silences
//	POST   /silences          – create a new silence (JSON body)
//	DELETE /silences?job=name – remove an existing silence immediately
package silence
