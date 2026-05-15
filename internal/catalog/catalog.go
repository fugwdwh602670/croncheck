// Package catalog maintains a registry of known jobs with their metadata
// such as description, owner, and schedule expression.
package catalog

import (
	"errors"
	"sync"
)

// Entry holds static metadata for a registered job.
type Entry struct {
	Job         string `json:"job"`
	Description string `json:"description"`
	Owner       string `json:"owner"`
	Schedule    string `json:"schedule"`
}

// Store holds the catalog of job entries.
type Store struct {
	mu      sync.RWMutex
	entries map[string]Entry
}

// New returns an empty catalog Store.
func New() *Store {
	return &Store{entries: make(map[string]Entry)}
}

// Set registers or updates the metadata entry for a job.
func (s *Store) Set(e Entry) error {
	if e.Job == "" {
		return errors.New("job name must not be empty")
	}
	s.mu.Lock()
	defer s.mu.Unlock()
	s.entries[e.Job] = e
	return nil
}

// Get returns the catalog entry for the given job.
func (s *Store) Get(job string) (Entry, bool) {
	s.mu.RLock()
	defer s.mu.RUnlock()
	e, ok := s.entries[job]
	return e, ok
}

// Remove deletes the catalog entry for the given job.
func (s *Store) Remove(job string) {
	s.mu.Lock()
	defer s.mu.Unlock()
	delete(s.entries, job)
}

// All returns a snapshot of all catalog entries.
func (s *Store) All() []Entry {
	s.mu.RLock()
	defer s.mu.RUnlock()
	out := make([]Entry, 0, len(s.entries))
	for _, e := range s.entries {
		out = append(out, e)
	}
	return out
}
