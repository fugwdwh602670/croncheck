// Package pausejob allows individual cron jobs to be paused so that
// missed-run alerts are suppressed while the pause is active.
package pausejob

import (
	"errors"
	"sync"
	"time"
)

// Entry holds the pause state for a single job.
type Entry struct {
	Job       string    `json:"job"`
	Reason    string    `json:"reason"`
	PausedAt  time.Time `json:"paused_at"`
	ExpiresAt time.Time `json:"expires_at"`
}

// Store manages pause entries.
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

// Pause adds or replaces a pause entry for the given job.
func (s *Store) Pause(job, reason string, duration time.Duration) error {
	if job == "" {
		return errors.New("job name must not be empty")
	}
	if duration <= 0 {
		return errors.New("duration must be positive")
	}
	now := s.now()
	s.mu.Lock()
	defer s.mu.Unlock()
	s.entries[job] = Entry{
		Job:       job,
		Reason:    reason,
		PausedAt:  now,
		ExpiresAt: now.Add(duration),
	}
	return nil
}

// Resume removes a pause entry for the given job.
func (s *Store) Resume(job string) {
	s.mu.Lock()
	defer s.mu.Unlock()
	delete(s.entries, job)
}

// IsPaused reports whether the job is currently paused.
func (s *Store) IsPaused(job string) bool {
	s.mu.RLock()
	defer s.mu.RUnlock()
	e, ok := s.entries[job]
	if !ok {
		return false
	}
	return s.now().Before(e.ExpiresAt)
}

// All returns a snapshot of active (non-expired) pause entries.
func (s *Store) All() []Entry {
	now := s.now()
	s.mu.RLock()
	defer s.mu.RUnlock()
	out := make([]Entry, 0, len(s.entries))
	for _, e := range s.entries {
		if now.Before(e.ExpiresAt) {
			out = append(out, e)
		}
	}
	return out
}
