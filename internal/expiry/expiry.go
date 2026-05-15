// Package expiry tracks per-job expiry deadlines and reports whether a job
// has exceeded its expected completion window.
package expiry

import (
	"errors"
	"sync"
	"time"
)

// Entry holds the expiry configuration for a single job.
type Entry struct {
	Job      string
	Deadline time.Time
	SetAt    time.Time
}

// Store holds expiry deadlines for monitored jobs.
type Store struct {
	mu      sync.RWMutex
	entries map[string]Entry
	now     func() time.Time
}

// New returns a new Store.
func New() *Store {
	return &Store{
		entries: make(map[string]Entry),
		now:     time.Now,
	}
}

// Set registers an expiry deadline for the given job.
func (s *Store) Set(job string, ttl time.Duration) error {
	if job == "" {
		return errors.New("job name must not be empty")
	}
	if ttl <= 0 {
		return errors.New("ttl must be positive")
	}
	now := s.now()
	s.mu.Lock()
	defer s.mu.Unlock()
	s.entries[job] = Entry{
		Job:      job,
		Deadline: now.Add(ttl),
		SetAt:    now,
	}
	return nil
}

// Get returns the expiry entry for the given job.
func (s *Store) Get(job string) (Entry, bool) {
	s.mu.RLock()
	defer s.mu.RUnlock()
	e, ok := s.entries[job]
	return e, ok
}

// IsExpired reports whether the job's deadline has passed.
func (s *Store) IsExpired(job string) bool {
	s.mu.RLock()
	defer s.mu.RUnlock()
	e, ok := s.entries[job]
	if !ok {
		return false
	}
	return s.now().After(e.Deadline)
}

// Remove deletes the expiry entry for the given job.
func (s *Store) Remove(job string) {
	s.mu.Lock()
	defer s.mu.Unlock()
	delete(s.entries, job)
}

// All returns a snapshot of all current expiry entries.
func (s *Store) All() []Entry {
	s.mu.RLock()
	defer s.mu.RUnlock()
	out := make([]Entry, 0, len(s.entries))
	for _, e := range s.entries {
		out = append(out, e)
	}
	return out
}
