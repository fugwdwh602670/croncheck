// Package grouping provides job grouping so that alerts and status can be
// aggregated by an arbitrary label such as team, service, or environment.
package grouping

import (
	"fmt"
	"sync"
)

// Store maps job names to a named group.
type Store struct {
	mu     sync.RWMutex
	groups map[string]string // job -> group
}

// New returns an empty grouping Store.
func New() *Store {
	return &Store{groups: make(map[string]string)}
}

// Set assigns job to group. Both must be non-empty.
func (s *Store) Set(job, group string) error {
	if job == "" {
		return fmt.Errorf("job name must not be empty")
	}
	if group == "" {
		return fmt.Errorf("group name must not be empty")
	}
	s.mu.Lock()
	defer s.mu.Unlock()
	s.groups[job] = group
	return nil
}

// Get returns the group for job and whether it was found.
func (s *Store) Get(job string) (string, bool) {
	s.mu.RLock()
	defer s.mu.RUnlock()
	g, ok := s.groups[job]
	return g, ok
}

// Remove deletes the group assignment for job.
func (s *Store) Remove(job string) {
	s.mu.Lock()
	defer s.mu.Unlock()
	delete(s.groups, job)
}

// JobsInGroup returns a snapshot of all job names that belong to group.
func (s *Store) JobsInGroup(group string) []string {
	s.mu.RLock()
	defer s.mu.RUnlock()
	var out []string
	for job, g := range s.groups {
		if g == group {
			out = append(out, job)
		}
	}
	return out
}

// All returns a snapshot copy of the entire job→group mapping.
func (s *Store) All() map[string]string {
	s.mu.RLock()
	defer s.mu.RUnlock()
	out := make(map[string]string, len(s.groups))
	for k, v := range s.groups {
		out[k] = v
	}
	return out
}
