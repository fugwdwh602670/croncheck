// Package tags provides job tagging and filtering support.
// Tags are arbitrary key-value labels attached to jobs, allowing
// operators to group, filter, and query jobs by metadata.
package tags

import (
	"fmt"
	"sync"
)

// Store holds tag mappings for jobs.
type Store struct {
	mu   sync.RWMutex
	tags map[string]map[string]string // job -> key -> value
}

// New returns an initialised Store.
func New() *Store {
	return &Store{tags: make(map[string]map[string]string)}
}

// Set replaces all tags for a job with the provided map.
func (s *Store) Set(job string, t map[string]string) {
	s.mu.Lock()
	defer s.mu.Unlock()
	copy := make(map[string]string, len(t))
	for k, v := range t {
		copy[k] = v
	}
	s.tags[job] = copy
}

// Get returns the tags for a job. The second return value is false
// if the job has no tags registered.
func (s *Store) Get(job string) (map[string]string, bool) {
	s.mu.RLock()
	defer s.mu.RUnlock()
	t, ok := s.tags[job]
	if !ok {
		return nil, false
	}
	copy := make(map[string]string, len(t))
	for k, v := range t {
		copy[k] = v
	}
	return copy, true
}

// Remove deletes all tags for a job.
func (s *Store) Remove(job string) {
	s.mu.Lock()
	defer s.mu.Unlock()
	delete(s.tags, job)
}

// Filter returns job names whose tags match all provided key-value pairs.
func (s *Store) Filter(query map[string]string) []string {
	s.mu.RLock()
	defer s.mu.RUnlock()
	var result []string
outer:
	for job, t := range s.tags {
		for k, v := range query {
			if t[k] != v {
				continue outer
			}
		}
		result = append(result, job)
	}
	return result
}

// All returns a snapshot of all job tags.
func (s *Store) All() map[string]map[string]string {
	s.mu.RLock()
	defer s.mu.RUnlock()
	out := make(map[string]map[string]string, len(s.tags))
	for job, t := range s.tags {
		copy := make(map[string]string, len(t))
		for k, v := range t {
			copy[k] = v
		}
		out[job] = copy
	}
	return out
}

// Validate checks that tag keys and values are non-empty.
func Validate(t map[string]string) error {
	for k, v := range t {
		if k == "" {
			return fmt.Errorf("tag key must not be empty")
		}
		if v == "" {
			return fmt.Errorf("tag value for key %q must not be empty", k)
		}
	}
	return nil
}
