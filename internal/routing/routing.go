// Package routing provides job-to-channel routing rules,
// allowing different alert channels to be assigned per job.
package routing

import (
	"errors"
	"sync"
)

// Rule maps a job to a named alert channel.
type Rule struct {
	Job     string `json:"job"`
	Channel string `json:"channel"`
}

// Store holds routing rules keyed by job name.
type Store struct {
	mu    sync.RWMutex
	rules map[string]string // job -> channel
}

// New returns an initialised Store.
func New() *Store {
	return &Store{rules: make(map[string]string)}
}

// Set assigns a channel to a job, overwriting any previous rule.
func (s *Store) Set(job, channel string) error {
	if job == "" {
		return errors.New("job name must not be empty")
	}
	if channel == "" {
		return errors.New("channel must not be empty")
	}
	s.mu.Lock()
	defer s.mu.Unlock()
	s.rules[job] = channel
	return nil
}

// Get returns the channel assigned to job, or ("", false) if none.
func (s *Store) Get(job string) (string, bool) {
	s.mu.RLock()
	defer s.mu.RUnlock()
	ch, ok := s.rules[job]
	return ch, ok
}

// Remove deletes the routing rule for job.
func (s *Store) Remove(job string) {
	s.mu.Lock()
	defer s.mu.Unlock()
	delete(s.rules, job)
}

// All returns a snapshot of all current routing rules.
func (s *Store) All() []Rule {
	s.mu.RLock()
	defer s.mu.RUnlock()
	out := make([]Rule, 0, len(s.rules))
	for job, ch := range s.rules {
		out = append(out, Rule{Job: job, Channel: ch})
	}
	return out
}
