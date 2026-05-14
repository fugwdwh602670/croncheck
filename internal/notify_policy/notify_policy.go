// Package notify_policy defines per-job notification policies that control
// which severity levels trigger alerts and how many consecutive misses
// are required before an alert is sent.
package notify_policy

import (
	"errors"
	"fmt"
	"sync"
)

// Severity represents the minimum alert level required to notify.
type Severity string

const (
	SeverityInfo     Severity = "info"
	SeverityWarning  Severity = "warning"
	SeverityCritical Severity = "critical"
)

var validSeverities = map[Severity]struct{}{
	SeverityInfo:     {},
	SeverityWarning:  {},
	SeverityCritical: {},
}

// Policy holds notification settings for a single job.
type Policy struct {
	MinSeverity    Severity `json:"min_severity"`
	MinConsecMisses int     `json:"min_consec_misses"`
}

// Store manages per-job notification policies.
type Store struct {
	mu       sync.RWMutex
	policies map[string]Policy
}

// New returns an initialised Store.
func New() *Store {
	return &Store{policies: make(map[string]Policy)}
}

// Set stores a policy for the given job, validating inputs.
func (s *Store) Set(job string, p Policy) error {
	if job == "" {
		return errors.New("job name must not be empty")
	}
	if _, ok := validSeverities[p.MinSeverity]; !ok {
		return fmt.Errorf("unknown severity %q", p.MinSeverity)
	}
	if p.MinConsecMisses < 1 {
		return errors.New("min_consec_misses must be at least 1")
	}
	s.mu.Lock()
	defer s.mu.Unlock()
	s.policies[job] = p
	return nil
}

// Get returns the policy for a job and whether one was found.
func (s *Store) Get(job string) (Policy, bool) {
	s.mu.RLock()
	defer s.mu.RUnlock()
	p, ok := s.policies[job]
	return p, ok
}

// Remove deletes the policy for a job.
func (s *Store) Remove(job string) {
	s.mu.Lock()
	defer s.mu.Unlock()
	delete(s.policies, job)
}

// All returns a snapshot of all stored policies.
func (s *Store) All() map[string]Policy {
	s.mu.RLock()
	defer s.mu.RUnlock()
	out := make(map[string]Policy, len(s.policies))
	for k, v := range s.policies {
		out[k] = v
	}
	return out
}
