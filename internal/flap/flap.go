// Package flap detects jobs that are repeatedly transitioning between
// healthy and missed states (flapping), and suppresses noisy alerts.
package flap

import (
	"sync"
	"time"
)

// Entry tracks state-change history for a single job.
type Entry struct {
	Changes    int
	LastChange time.Time
	Flapping   bool
}

// Store tracks flap state for monitored jobs.
type Store struct {
	mu        sync.Mutex
	entries   map[string]*Entry
	window    time.Duration
	threshold int
}

// New creates a Store. window is the look-back period; threshold is the
// number of state changes within that window required to declare flapping.
func New(window time.Duration, threshold int) *Store {
	if threshold <= 0 {
		threshold = 4
	}
	if window <= 0 {
		window = 10 * time.Minute
	}
	return &Store{
		entries:   make(map[string]*Entry),
		window:    window,
		threshold: threshold,
	}
}

// Record registers a state change for job and returns whether the job is
// currently considered flapping.
func (s *Store) Record(job string, now time.Time) bool {
	s.mu.Lock()
	defer s.mu.Unlock()

	e, ok := s.entries[job]
	if !ok {
		e = &Entry{}
		s.entries[job] = e
	}

	// Reset counter if the last change is outside the window.
	if !e.LastChange.IsZero() && now.Sub(e.LastChange) > s.window {
		e.Changes = 0
	}

	e.Changes++
	e.LastChange = now
	e.Flapping = e.Changes >= s.threshold
	return e.Flapping
}

// IsFlapping reports whether job is currently flapping.
func (s *Store) IsFlapping(job string) bool {
	s.mu.Lock()
	defer s.mu.Unlock()
	e, ok := s.entries[job]
	if !ok {
		return false
	}
	return e.Flapping
}

// Reset clears the flap state for job.
func (s *Store) Reset(job string) {
	s.mu.Lock()
	defer s.mu.Unlock()
	delete(s.entries, job)
}

// All returns a snapshot of all tracked entries.
func (s *Store) All() map[string]Entry {
	s.mu.Lock()
	defer s.mu.Unlock()
	out := make(map[string]Entry, len(s.entries))
	for k, v := range s.entries {
		out[k] = *v
	}
	return out
}
