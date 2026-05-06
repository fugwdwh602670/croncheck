// Package annotations provides per-job free-form key/value metadata storage.
package annotations

import (
	"maps"
	"sync"
)

// Store holds arbitrary string annotations keyed by job name.
type Store struct {
	mu   sync.RWMutex
	data map[string]map[string]string
}

// New returns an initialised Store.
func New() *Store {
	return &Store{data: make(map[string]map[string]string)}
}

// Set replaces all annotations for job with the provided map.
// A nil or empty map clears the job's annotations.
func (s *Store) Set(job string, annotations map[string]string) {
	s.mu.Lock()
	defer s.mu.Unlock()
	if len(annotations) == 0 {
		delete(s.data, job)
		return
	}
	s.data[job] = maps.Clone(annotations)
}

// Get returns a copy of the annotations for job.
// The second return value is false if the job has no annotations.
func (s *Store) Get(job string) (map[string]string, bool) {
	s.mu.RLock()
	defer s.mu.RUnlock()
	a, ok := s.data[job]
	if !ok {
		return nil, false
	}
	return maps.Clone(a), true
}

// Remove deletes all annotations for job.
func (s *Store) Remove(job string) {
	s.mu.Lock()
	defer s.mu.Unlock()
	delete(s.data, job)
}

// All returns a snapshot of every job's annotations.
func (s *Store) All() map[string]map[string]string {
	s.mu.RLock()
	defer s.mu.RUnlock()
	out := make(map[string]map[string]string, len(s.data))
	for k, v := range s.data {
		out[k] = maps.Clone(v)
	}
	return out
}
