// Package oncall manages on-call schedule overrides for alert routing.
// It allows associating a contact (e.g. email or webhook URL) with a job
// for a fixed duration so alerts are routed to the on-call person.
package oncall

import (
	"errors"
	"sync"
	"time"
)

// Entry represents an active on-call override for a single job.
type Entry struct {
	Job      string    `json:"job"`
	Contact  string    `json:"contact"`
	ExpiresAt time.Time `json:"expires_at"`
}

// Store holds on-call overrides keyed by job name.
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

// Set registers an on-call override for job lasting duration d.
func (s *Store) Set(job, contact string, d time.Duration) error {
	if job == "" {
		return errors.New("job name must not be empty")
	}
	if contact == "" {
		return errors.New("contact must not be empty")
	}
	if d <= 0 {
		return errors.New("duration must be positive")
	}
	s.mu.Lock()
	defer s.mu.Unlock()
	s.entries[job] = Entry{
		Job:      job,
		Contact:  contact,
		ExpiresAt: s.now().Add(d),
	}
	return nil
}

// Get returns the active on-call entry for job, if any.
func (s *Store) Get(job string) (Entry, bool) {
	s.mu.RLock()
	defer s.mu.RUnlock()
	e, ok := s.entries[job]
	if !ok || s.now().After(e.ExpiresAt) {
		return Entry{}, false
	}
	return e, true
}

// Remove deletes the on-call override for job.
func (s *Store) Remove(job string) {
	s.mu.Lock()
	defer s.mu.Unlock()
	delete(s.entries, job)
}

// All returns a snapshot of currently active on-call entries.
func (s *Store) All() []Entry {
	s.mu.RLock()
	defer s.mu.RUnlock()
	now := s.now()
	out := make([]Entry, 0, len(s.entries))
	for _, e := range s.entries {
		if !now.After(e.ExpiresAt) {
			out = append(out, e)
		}
	}
	return out
}
