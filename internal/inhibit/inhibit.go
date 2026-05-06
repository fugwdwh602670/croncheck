// Package inhibit provides dependency-based alert suppression.
// When a parent job is missed or failing, alerts for dependent jobs
// can be inhibited to reduce noise.
package inhibit

import (
	"sync"
	"time"
)

// Rule defines an inhibition: if SourceJob is unhealthy, suppress alerts for TargetJob.
type Rule struct {
	SourceJob string
	TargetJob string
}

// Store holds inhibition rules and tracks which source jobs are currently unhealthy.
type Store struct {
	mu       sync.RWMutex
	rules    []Rule
	unhealthy map[string]time.Time // source job -> time it became unhealthy
}

// New creates a new inhibition Store with the given rules.
func New(rules []Rule) *Store {
	return &Store{
		rules:     rules,
		unhealthy: make(map[string]time.Time),
	}
}

// SetUnhealthy marks a source job as unhealthy as of now.
func (s *Store) SetUnhealthy(job string) {
	s.mu.Lock()
	defer s.mu.Unlock()
	if _, exists := s.unhealthy[job]; !exists {
		s.unhealthy[job] = time.Now()
	}
}

// SetHealthy clears the unhealthy state for a job.
func (s *Store) SetHealthy(job string) {
	s.mu.Lock()
	defer s.mu.Unlock()
	delete(s.unhealthy, job)
}

// IsInhibited returns true if any rule suppresses alerts for the given target job.
func (s *Store) IsInhibited(targetJob string) bool {
	s.mu.RLock()
	defer s.mu.RUnlock()
	for _, rule := range s.rules {
		if rule.TargetJob == targetJob {
			if _, unhealthy := s.unhealthy[rule.SourceJob]; unhealthy {
				return true
			}
		}
	}
	return false
}

// UnhealthySources returns a snapshot of currently unhealthy source jobs.
func (s *Store) UnhealthySources() map[string]time.Time {
	s.mu.RLock()
	defer s.mu.RUnlock()
	out := make(map[string]time.Time, len(s.unhealthy))
	for k, v := range s.unhealthy {
		out[k] = v
	}
	return out
}

// Rules returns a copy of all configured inhibition rules.
func (s *Store) Rules() []Rule {
	s.mu.RLock()
	defer s.mu.RUnlock()
	out := make([]Rule, len(s.rules))
	copy(out, s.rules)
	return out
}
