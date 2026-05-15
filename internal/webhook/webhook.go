// Package webhook manages per-job webhook endpoint configurations.
// Each job may have a webhook URL registered; when an alert fires the
// notifier can POST a JSON payload to that URL in addition to (or
// instead of) the global alert endpoint.
package webhook

import (
	"errors"
	"sync"
)

// Entry holds the webhook configuration for a single job.
type Entry struct {
	Job string `json:"job"`
	URL string `json:"url"`
	Secret string `json:"secret,omitempty"`
}

// Store holds webhook registrations keyed by job name.
type Store struct {
	mu      sync.RWMutex
	entries map[string]Entry
}

// New returns an initialised Store.
func New() *Store {
	return &Store{entries: make(map[string]Entry)}
}

// Set registers or replaces the webhook for job.
func (s *Store) Set(job, url, secret string) error {
	if job == "" {
		return errors.New("job name must not be empty")
	}
	if url == "" {
		return errors.New("url must not be empty")
	}
	s.mu.Lock()
	defer s.mu.Unlock()
	s.entries[job] = Entry{Job: job, URL: url, Secret: secret}
	return nil
}

// Get returns the webhook entry for job and whether it exists.
func (s *Store) Get(job string) (Entry, bool) {
	s.mu.RLock()
	defer s.mu.RUnlock()
	e, ok := s.entries[job]
	return e, ok
}

// Remove deletes the webhook registration for job.
func (s *Store) Remove(job string) {
	s.mu.Lock()
	defer s.mu.Unlock()
	delete(s.entries, job)
}

// All returns a snapshot of all registered webhooks.
func (s *Store) All() []Entry {
	s.mu.RLock()
	defer s.mu.RUnlock()
	out := make([]Entry, 0, len(s.entries))
	for _, e := range s.entries {
		out = append(out, e)
	}
	return out
}
