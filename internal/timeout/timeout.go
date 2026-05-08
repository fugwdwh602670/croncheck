// Package timeout tracks per-job execution timeout thresholds and reports
// whether a job has exceeded its allowed runtime since last heartbeat.
package timeout

import (
	"fmt"
	"sync"
	"time"
)

// Entry holds the timeout configuration for a single job.
type Entry struct {
	JobName   string        `json:"job_name"`
	Threshold time.Duration `json:"threshold_seconds"`
	SetAt     time.Time     `json:"set_at"`
}

// Store holds timeout thresholds keyed by job name.
type Store struct {
	mu      sync.RWMutex
	entries map[string]Entry
}

// New returns an initialised Store.
func New() *Store {
	return &Store{entries: make(map[string]Entry)}
}

// Set registers or updates the timeout threshold for a job.
// threshold must be positive.
func (s *Store) Set(jobName string, threshold time.Duration) error {
	if jobName == "" {
		return fmt.Errorf("job name must not be empty")
	}
	if threshold <= 0 {
		return fmt.Errorf("threshold must be positive, got %s", threshold)
	}
	s.mu.Lock()
	defer s.mu.Unlock()
	s.entries[jobName] = Entry{
		JobName:   jobName,
		Threshold: threshold,
		SetAt:     time.Now(),
	}
	return nil
}

// Get returns the Entry for the given job and true, or zero Entry and false if
// no threshold has been registered.
func (s *Store) Get(jobName string) (Entry, bool) {
	s.mu.RLock()
	defer s.mu.RUnlock()
	e, ok := s.entries[jobName]
	return e, ok
}

// Remove deletes the timeout entry for the given job.
func (s *Store) Remove(jobName string) {
	s.mu.Lock()
	defer s.mu.Unlock()
	delete(s.entries, jobName)
}

// IsExceeded reports whether the job's last heartbeat (lastSeen) is older than
// its registered threshold. Returns false if no threshold is registered.
func (s *Store) IsExceeded(jobName string, lastSeen time.Time, now time.Time) bool {
	e, ok := s.Get(jobName)
	if !ok {
		return false
	}
	return now.Sub(lastSeen) > e.Threshold
}

// All returns a snapshot of all registered timeout entries.
func (s *Store) All() []Entry {
	s.mu.RLock()
	defer s.mu.RUnlock()
	out := make([]Entry, 0, len(s.entries))
	for _, e := range s.entries {
		out = append(out, e)
	}
	return out
}
