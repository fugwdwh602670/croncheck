// Package runlog tracks the last known run result (success/failure) for each
// monitored job and exposes a simple in-memory store that other packages can
// query.
package runlog

import (
	"sync"
	"time"
)

// Status represents the outcome of a cron job run.
type Status string

const (
	StatusSuccess Status = "success"
	StatusFailure Status = "failure"
)

// Entry holds the most-recent run information for a single job.
type Entry struct {
	JobName   string
	Status    Status
	Message   string
	RecordedAt time.Time
}

// Store records and retrieves run-log entries.
type Store struct {
	mu      sync.RWMutex
	entries map[string]Entry
}

// New returns an initialised Store.
func New() *Store {
	return &Store{entries: make(map[string]Entry)}
}

// Record stores (or overwrites) the run result for the given job.
func (s *Store) Record(jobName string, status Status, message string) {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.entries[jobName] = Entry{
		JobName:    jobName,
		Status:     status,
		Message:    message,
		RecordedAt: time.Now().UTC(),
	}
}

// Get returns the latest Entry for jobName and whether it was found.
func (s *Store) Get(jobName string) (Entry, bool) {
	s.mu.RLock()
	defer s.mu.RUnlock()
	e, ok := s.entries[jobName]
	return e, ok
}

// All returns a snapshot of every recorded entry.
func (s *Store) All() []Entry {
	s.mu.RLock()
	defer s.mu.RUnlock()
	out := make([]Entry, 0, len(s.entries))
	for _, e := range s.entries {
		out = append(out, e)
	}
	return out
}

// Clear removes the entry for jobName, if present.
func (s *Store) Clear(jobName string) {
	s.mu.Lock()
	defer s.mu.Unlock()
	delete(s.entries, jobName)
}
