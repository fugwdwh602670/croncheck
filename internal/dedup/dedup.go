// Package dedup provides alert deduplication tracking for cron jobs.
// It suppresses repeated alerts for the same job within a configurable window.
package dedup

import (
	"sync"
	"time"
)

// entry holds deduplication state for a single job.
type entry struct {
	key       string
	firstSeen time.Time
	count     int
	window    time.Duration
}

// Store tracks deduplication state for jobs.
type Store struct {
	mu      sync.Mutex
	entries map[string]*entry
	now     func() time.Time
}

// New returns a new deduplication Store.
func New() *Store {
	return &Store{
		entries: make(map[string]*entry),
		now:     time.Now,
	}
}

// IsDuplicate reports whether an alert for job is a duplicate within window.
// The first call for a given job+window always returns false (it is the original).
// Subsequent calls within the window return true.
func (s *Store) IsDuplicate(job string, window time.Duration) bool {
	s.mu.Lock()
	defer s.mu.Unlock()

	now := s.now()
	e, ok := s.entries[job]
	if !ok || now.After(e.firstSeen.Add(e.window)) {
		s.entries[job] = &entry{
			key:       job,
			firstSeen: now,
			count:     1,
			window:    window,
		}
		return false
	}
	e.count++
	return true
}

// Reset clears deduplication state for a job, allowing the next alert through.
func (s *Store) Reset(job string) {
	s.mu.Lock()
	defer s.mu.Unlock()
	delete(s.entries, job)
}

// Count returns the number of deduplicated (suppressed) alerts for a job.
// Returns 0 for unknown jobs.
func (s *Store) Count(job string) int {
	s.mu.Lock()
	defer s.mu.Unlock()
	e, ok := s.entries[job]
	if !ok {
		return 0
	}
	return e.count
}

// All returns a snapshot of all active dedup entries keyed by job name.
func (s *Store) All() map[string]int {
	s.mu.Lock()
	defer s.mu.Unlock()
	out := make(map[string]int, len(s.entries))
	now := s.now()
	for job, e := range s.entries {
		if !now.After(e.firstSeen.Add(e.window)) {
			out[job] = e.count
		}
	}
	return out
}
