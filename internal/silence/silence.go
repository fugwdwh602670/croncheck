// Package silence provides a registry for temporarily suppressing alerts
// for specific cron jobs during maintenance windows or known downtime.
package silence

import (
	"sync"
	"time"
)

// Silence represents a suppression window for a single job.
type Silence struct {
	JobName   string
	Reason    string
	ExpiresAt time.Time
}

// Registry holds active silences keyed by job name.
type Registry struct {
	mu      sync.RWMutex
	records map[string]Silence
}

// New returns an initialised Registry.
func New() *Registry {
	return &Registry{records: make(map[string]Silence)}
}

// Add registers a silence for the given job until expiresAt.
func (r *Registry) Add(jobName, reason string, expiresAt time.Time) {
	r.mu.Lock()
	defer r.mu.Unlock()
	r.records[jobName] = Silence{
		JobName:   jobName,
		Reason:    reason,
		ExpiresAt: expiresAt,
	}
}

// Remove deletes a silence for the given job immediately.
func (r *Registry) Remove(jobName string) {
	r.mu.Lock()
	defer r.mu.Unlock()
	delete(r.records, jobName)
}

// IsSilenced reports whether the job currently has an active (non-expired) silence.
func (r *Registry) IsSilenced(jobName string, now time.Time) bool {
	r.mu.RLock()
	defer r.mu.RUnlock()
	s, ok := r.records[jobName]
	if !ok {
		return false
	}
	if now.After(s.ExpiresAt) {
		return false
	}
	return true
}

// All returns a snapshot of all currently active silences.
func (r *Registry) All(now time.Time) []Silence {
	r.mu.RLock()
	defer r.mu.RUnlock()
	out := make([]Silence, 0, len(r.records))
	for _, s := range r.records {
		if !now.After(s.ExpiresAt) {
			out = append(out, s)
		}
	}
	return out
}
