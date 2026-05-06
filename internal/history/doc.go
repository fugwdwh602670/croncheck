// Package history provides a bounded in-memory ring buffer for recording
// recent alert events (missed or failed runs) per cron job.
//
// Events are stored per job name up to a configurable limit; once the buffer
// is full the oldest event is evicted to make room for the newest.  All
// operations are safe for concurrent use.
package history
