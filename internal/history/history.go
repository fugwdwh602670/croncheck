// Package history provides a bounded in-memory ring buffer that records
// recent alert events for each job so operators can review past incidents
// via the API without external storage.
package history

import (
	"sync"
	"time"
)

// Event represents a single alert event recorded for a job.
type Event struct {
	JobName   string    `json:"job_name"`
	Kind      string    `json:"kind"` // "missed" or "failed"
	OccuredAt time.Time `json:"occurred_at"`
}

// Recorder keeps a fixed-size ring buffer of Events per job.
type Recorder struct {
	mu     sync.RWMutex
	events map[string][]Event
	limit  int
}

// New returns a Recorder that retains at most limit events per job.
// If limit is <= 0 it defaults to 10.
func New(limit int) *Recorder {
	if limit <= 0 {
		limit = 10
	}
	return &Recorder{
		events: make(map[string][]Event),
		limit:  limit,
	}
}

// Record appends an event for the given job, evicting the oldest entry
// when the ring buffer is full.
func (r *Recorder) Record(e Event) {
	r.mu.Lock()
	defer r.mu.Unlock()

	buf := r.events[e.JobName]
	if len(buf) >= r.limit {
		buf = buf[1:]
	}
	r.events[e.JobName] = append(buf, e)
}

// Get returns a snapshot of recorded events for the named job.
// It returns nil if no events have been recorded for that job.
func (r *Recorder) Get(jobName string) []Event {
	r.mu.RLock()
	defer r.mu.RUnlock()

	buf := r.events[jobName]
	if len(buf) == 0 {
		return nil
	}
	out := make([]Event, len(buf))
	copy(out, buf)
	return out
}

// All returns a snapshot of all recorded events keyed by job name.
func (r *Recorder) All() map[string][]Event {
	r.mu.RLock()
	defer r.mu.RUnlock()

	out := make(map[string][]Event, len(r.events))
	for k, v := range r.events {
		cp := make([]Event, len(v))
		copy(cp, v)
		out[k] = cp
	}
	return out
}
