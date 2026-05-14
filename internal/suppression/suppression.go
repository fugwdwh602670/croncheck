// Package suppression provides per-job alert suppression rules based on
// consecutive miss counts. When a job's miss count is below the configured
// threshold, alerts are suppressed.
package suppression

import (
	"errors"
	"fmt"
	"sync"
)

// Rule defines the suppression configuration for a single job.
type Rule struct {
	Job              string `json:"job"`
	MinConsecMisses  int    `json:"min_consec_misses"`
}

// Store holds suppression rules keyed by job name.
type Store struct {
	mu    sync.RWMutex
	rules map[string]Rule
}

// New returns an empty suppression Store.
func New() *Store {
	return &Store{rules: make(map[string]Rule)}
}

// Set registers or replaces the suppression rule for a job.
func (s *Store) Set(job string, minConsecMisses int) error {
	if job == "" {
		return errors.New("job name must not be empty")
	}
	if minConsecMisses <= 0 {
		return fmt.Errorf("min_consec_misses must be positive, got %d", minConsecMisses)
	}
	s.mu.Lock()
	defer s.mu.Unlock()
	s.rules[job] = Rule{Job: job, MinConsecMisses: minConsecMisses}
	return nil
}

// Get returns the suppression rule for a job, or false if none is set.
func (s *Store) Get(job string) (Rule, bool) {
	s.mu.RLock()
	defer s.mu.RUnlock()
	r, ok := s.rules[job]
	return r, ok
}

// Remove deletes the suppression rule for a job.
func (s *Store) Remove(job string) {
	s.mu.Lock()
	defer s.mu.Unlock()
	delete(s.rules, job)
}

// IsSuppressed reports whether an alert for job should be suppressed given the
// current consecutive miss count. Returns false when no rule is set.
func (s *Store) IsSuppressed(job string, consecMisses int) bool {
	r, ok := s.Get(job)
	if !ok {
		return false
	}
	return consecMisses < r.MinConsecMisses
}

// All returns a snapshot of all configured rules.
func (s *Store) All() []Rule {
	s.mu.RLock()
	defer s.mu.RUnlock()
	out := make([]Rule, 0, len(s.rules))
	for _, r := range s.rules {
		out = append(out, r)
	}
	return out
}
