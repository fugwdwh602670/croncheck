// Package jitter provides per-job grace period (jitter) configuration,
// allowing a configurable window after a job's expected run time before
// it is considered missed.
package jitter

import (
	"errors"
	"sync"
	"time"
)

// Store holds jitter (grace period) settings for each job.
type Store struct {
	mu      sync.RWMutex
	entries map[string]time.Duration
}

// New returns an initialised Store.
func New() *Store {
	return &Store{entries: make(map[string]time.Duration)}
}

// Set records a grace period for the named job.
// Returns an error if job is empty or duration is non-positive.
func (s *Store) Set(job string, d time.Duration) error {
	if job == "" {
		return errors.New("job name must not be empty")
	}
	if d <= 0 {
		return errors.New("jitter duration must be positive")
	}
	s.mu.Lock()
	defer s.mu.Unlock()
	s.entries[job] = d
	return nil
}

// Get returns the grace period for the named job and whether it was found.
func (s *Store) Get(job string) (time.Duration, bool) {
	s.mu.RLock()
	defer s.mu.RUnlock()
	d, ok := s.entries[job]
	return d, ok
}

// Remove deletes the grace period entry for the named job.
func (s *Store) Remove(job string) {
	s.mu.Lock()
	defer s.mu.Unlock()
	delete(s.entries, job)
}

// All returns a snapshot of all current jitter entries.
func (s *Store) All() map[string]time.Duration {
	s.mu.RLock()
	defer s.mu.RUnlock()
	out := make(map[string]time.Duration, len(s.entries))
	for k, v := range s.entries {
		out[k] = v
	}
	return out
}

// Grace returns the effective grace period for a job, falling back to
// defaultGrace when no entry exists.
func (s *Store) Grace(job string, defaultGrace time.Duration) time.Duration {
	if d, ok := s.Get(job); ok {
		return d
	}
	return defaultGrace
}
