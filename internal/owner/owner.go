// Package owner tracks ownership assignments for monitored cron jobs.
// Each job may be assigned an owner (team or individual) with optional
// contact information. Ownership records have no expiry and persist until
// explicitly removed.
package owner

import (
	"errors"
	"sync"
)

// Entry holds ownership information for a single job.
type Entry struct {
	Job     string `json:"job"`
	Owner   string `json:"owner"`
	Contact string `json:"contact,omitempty"`
}

// Store holds owner assignments keyed by job name.
type Store struct {
	mu      sync.RWMutex
	entries map[string]Entry
}

// New returns an initialised Store.
func New() *Store {
	return &Store{entries: make(map[string]Entry)}
}

// Set assigns an owner to a job, overwriting any previous entry.
func (s *Store) Set(job, owner, contact string) error {
	if job == "" {
		return errors.New("job name must not be empty")
	}
	if owner == "" {
		return errors.New("owner must not be empty")
	}
	s.mu.Lock()
	defer s.mu.Unlock()
	s.entries[job] = Entry{Job: job, Owner: owner, Contact: contact}
	return nil
}

// Get returns the ownership entry for a job, or false if not found.
func (s *Store) Get(job string) (Entry, bool) {
	s.mu.RLock()
	defer s.mu.RUnlock()
	e, ok := s.entries[job]
	return e, ok
}

// Remove deletes the ownership record for a job.
func (s *Store) Remove(job string) {
	s.mu.Lock()
	defer s.mu.Unlock()
	delete(s.entries, job)
}

// All returns a snapshot of every ownership entry.
func (s *Store) All() []Entry {
	s.mu.RLock()
	defer s.mu.RUnlock()
	out := make([]Entry, 0, len(s.entries))
	for _, e := range s.entries {
		out = append(out, e)
	}
	return out
}
