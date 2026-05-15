// Package cooldown enforces a minimum gap between successive heartbeats
// for a job. If a heartbeat arrives sooner than the configured cooldown
// duration, it is rejected so that noisy jobs do not flood the store.
package cooldown

import (
	"errors"
	"sync"
	"time"
)

// ErrTooSoon is returned when a heartbeat arrives within the cooldown window.
var ErrTooSoon = errors.New("heartbeat received within cooldown window")

// Entry holds the cooldown configuration and last-accepted time for a job.
type Entry struct {
	Duration  time.Duration
	LastSeen  time.Time
}

// Store tracks per-job cooldown state.
type Store struct {
	mu      sync.RWMutex
	entries map[string]Entry
	now     func() time.Time
}

// New returns an initialised Store.
func New() *Store {
	return &Store{
		entries: make(map[string]Entry),
		now:     time.Now,
	}
}

// Set configures the cooldown duration for a job.
// Returns an error when job is empty or duration is non-positive.
func (s *Store) Set(job string, d time.Duration) error {
	if job == "" {
		return errors.New("job name must not be empty")
	}
	if d <= 0 {
		return errors.New("cooldown duration must be positive")
	}
	s.mu.Lock()
	defer s.mu.Unlock()
	e := s.entries[job]
	e.Duration = d
	s.entries[job] = e
	return nil
}

// Allow checks whether a heartbeat for job should be accepted.
// It updates LastSeen on success. If no cooldown is configured for the job,
// the heartbeat is always allowed.
func (s *Store) Allow(job string) error {
	s.mu.Lock()
	defer s.mu.Unlock()
	e, ok := s.entries[job]
	if !ok {
		return nil
	}
	now := s.now()
	if !e.LastSeen.IsZero() && now.Sub(e.LastSeen) < e.Duration {
		return ErrTooSoon
	}
	e.LastSeen = now
	s.entries[job] = e
	return nil
}

// Get returns the Entry for a job and whether it exists.
func (s *Store) Get(job string) (Entry, bool) {
	s.mu.RLock()
	defer s.mu.RUnlock()
	e, ok := s.entries[job]
	return e, ok
}

// Remove deletes the cooldown configuration for a job.
func (s *Store) Remove(job string) {
	s.mu.Lock()
	defer s.mu.Unlock()
	delete(s.entries, job)
}

// All returns a snapshot of all configured cooldowns.
func (s *Store) All() map[string]Entry {
	s.mu.RLock()
	defer s.mu.RUnlock()
	out := make(map[string]Entry, len(s.entries))
	for k, v := range s.entries {
		out[k] = v
	}
	return out
}
