// Package dependency tracks job dependencies so that a job is only
// considered healthy when all of its declared upstream jobs are also healthy.
package dependency

import (
	"fmt"
	"sync"
)

// Store holds the dependency graph for all monitored jobs.
type Store struct {
	mu   sync.RWMutex
	edges map[string][]string // job -> list of upstream jobs it depends on
}

// New returns an initialised Store.
func New() *Store {
	return &Store{edges: make(map[string][]string)}
}

// Set replaces the dependency list for job with the given upstreams.
// Passing an empty slice removes all dependencies.
func (s *Store) Set(job string, upstreams []string) error {
	for _, u := range upstreams {
		if u == job {
			return fmt.Errorf("dependency: job %q cannot depend on itself", job)
		}
	}
	s.mu.Lock()
	defer s.mu.Unlock()
	cp := make([]string, len(upstreams))
	copy(cp, upstreams)
	s.edges[job] = cp
	return nil
}

// Get returns the upstream dependencies declared for job.
// Returns nil if no dependencies are registered.
func (s *Store) Get(job string) []string {
	s.mu.RLock()
	defer s.mu.RUnlock()
	ups := s.edges[job]
	if len(ups) == 0 {
		return nil
	}
	cp := make([]string, len(ups))
	copy(cp, ups)
	return cp
}

// Remove deletes all dependency information for job.
func (s *Store) Remove(job string) {
	s.mu.Lock()
	defer s.mu.Unlock()
	delete(s.edges, job)
}

// BlockedBy returns the subset of job's upstream dependencies whose health
// status is reported as unhealthy by the provided isHealthy function.
// If all upstreams are healthy (or there are none) the returned slice is empty.
func (s *Store) BlockedBy(job string, isHealthy func(string) bool) []string {
	ups := s.Get(job)
	var blocked []string
	for _, u := range ups {
		if !isHealthy(u) {
			blocked = append(blocked, u)
		}
	}
	return blocked
}

// All returns a snapshot of the full dependency map.
func (s *Store) All() map[string][]string {
	s.mu.RLock()
	defer s.mu.RUnlock()
	out := make(map[string][]string, len(s.edges))
	for k, v := range s.edges {
		cp := make([]string, len(v))
		copy(cp, v)
		out[k] = cp
	}
	return out
}
