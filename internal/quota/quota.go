// Package quota tracks per-job alert quota usage, enforcing a maximum
// number of alerts that may be sent within a rolling time window.
package quota

import (
	"fmt"
	"sync"
	"time"
)

// Entry holds quota state for a single job.
type Entry struct {
	Count     int
	WindowEnd time.Time
}

// Store manages quota entries for all jobs.
type Store struct {
	mu      sync.Mutex
	entries map[string]*Entry
	max     int
	window  time.Duration
	now     func() time.Time
}

// New creates a Store that allows at most max alerts per window duration.
func New(max int, window time.Duration) (*Store, error) {
	if max <= 0 {
		return nil, fmt.Errorf("quota: max must be positive, got %d", max)
	}
	if window <= 0 {
		return nil, fmt.Errorf("quota: window must be positive, got %s", window)
	}
	return &Store{
		entries: make(map[string]*Entry),
		max:     max,
		window:  window,
		now:     time.Now,
	}, nil
}

// Allow returns true and increments the counter if the job is within quota.
// It resets the window if the previous window has expired.
func (s *Store) Allow(job string) bool {
	s.mu.Lock()
	defer s.mu.Unlock()

	now := s.now()
	e, ok := s.entries[job]
	if !ok || now.After(e.WindowEnd) {
		s.entries[job] = &Entry{Count: 1, WindowEnd: now.Add(s.window)}
		return true
	}
	if e.Count >= s.max {
		return false
	}
	e.Count++
	return true
}

// Get returns the current quota entry for a job, or nil if unseen.
func (s *Store) Get(job string) *Entry {
	s.mu.Lock()
	defer s.mu.Unlock()
	e, ok := s.entries[job]
	if !ok {
		return nil
	}
	copy := *e
	return &copy
}

// Reset clears quota state for a job, allowing a fresh window to begin.
func (s *Store) Reset(job string) {
	s.mu.Lock()
	defer s.mu.Unlock()
	delete(s.entries, job)
}

// All returns a snapshot of all current quota entries.
func (s *Store) All() map[string]Entry {
	s.mu.Lock()
	defer s.mu.Unlock()
	out := make(map[string]Entry, len(s.entries))
	for k, v := range s.entries {
		out[k] = *v
	}
	return out
}
