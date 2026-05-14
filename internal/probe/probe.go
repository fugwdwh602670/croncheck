// Package probe provides a lightweight liveness probe mechanism that
// tracks whether a job has recently checked in within an expected window.
package probe

import (
	"errors"
	"sync"
	"time"
)

// Status represents the liveness state of a job probe.
type Status string

const (
	StatusAlive   Status = "alive"
	StatusDead    Status = "dead"
	StatusUnknown Status = "unknown"
)

// Entry holds probe state for a single job.
type Entry struct {
	Job       string
	LastProbe time.Time
	TTL       time.Duration
	Status    Status
}

// Store holds probe entries for all tracked jobs.
type Store struct {
	mu      sync.RWMutex
	entries map[string]*Entry
	now     func() time.Time
}

// New returns a new probe Store.
func New() *Store {
	return &Store{
		entries: make(map[string]*Entry),
		now:     time.Now,
	}
}

// Record registers a probe check-in for the given job with the given TTL.
func (s *Store) Record(job string, ttl time.Duration) error {
	if job == "" {
		return errors.New("job name must not be empty")
	}
	if ttl <= 0 {
		return errors.New("ttl must be positive")
	}
	s.mu.Lock()
	defer s.mu.Unlock()
	s.entries[job] = &Entry{
		Job:       job,
		LastProbe: s.now(),
		TTL:       ttl,
		Status:    StatusAlive,
	}
	return nil
}

// Get returns the current probe status for the given job.
func (s *Store) Get(job string) (Entry, bool) {
	s.mu.RLock()
	defer s.mu.RUnlock()
	e, ok := s.entries[job]
	if !ok {
		return Entry{Job: job, Status: StatusUnknown}, false
	}
	copy := *e
	if s.now().After(e.LastProbe.Add(e.TTL)) {
		copy.Status = StatusDead
	}
	return copy, true
}

// All returns a snapshot of all probe entries with current status applied.
func (s *Store) All() []Entry {
	s.mu.RLock()
	defer s.mu.RUnlock()
	out := make([]Entry, 0, len(s.entries))
	for _, e := range s.entries {
		copy := *e
		if s.now().After(e.LastProbe.Add(e.TTL)) {
			copy.Status = StatusDead
		}
		out = append(out, copy)
	}
	return out
}
