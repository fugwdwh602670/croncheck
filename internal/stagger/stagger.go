// Package stagger provides per-job stagger delay configuration,
// allowing cron checks to be offset by a fixed duration to avoid
// thundering-herd effects when many jobs share the same schedule.
package stagger

import (
	"errors"
	"sync"
	"time"
)

// Entry holds the stagger configuration for a single job.
type Entry struct {
	Job   string
	Delay time.Duration
}

// Store holds stagger delays keyed by job name.
type Store struct {
	mu      sync.RWMutex
	entries map[string]time.Duration
}

// New returns an initialised Store.
func New() *Store {
	return &Store{entries: make(map[string]time.Duration)}
}

// Set registers a stagger delay for the given job.
// The delay must be positive and the job name must be non-empty.
func (s *Store) Set(job string, delay time.Duration) error {
	if job == "" {
		return errors.New("job name must not be empty")
	}
	if delay <= 0 {
		return errors.New("stagger delay must be positive")
	}
	s.mu.Lock()
	defer s.mu.Unlock()
	s.entries[job] = delay
	return nil
}

// Get returns the stagger delay for a job and whether it was found.
func (s *Store) Get(job string) (time.Duration, bool) {
	s.mu.RLock()
	defer s.mu.RUnlock()
	d, ok := s.entries[job]
	return d, ok
}

// Remove deletes the stagger entry for the given job.
func (s *Store) Remove(job string) {
	s.mu.Lock()
	defer s.mu.Unlock()
	delete(s.entries, job)
}

// All returns a snapshot of all current stagger entries.
func (s *Store) All() []Entry {
	s.mu.RLock()
	defer s.mu.RUnlock()
	out := make([]Entry, 0, len(s.entries))
	for job, delay := range s.entries {
		out = append(out, Entry{Job: job, Delay: delay})
	}
	return out
}
