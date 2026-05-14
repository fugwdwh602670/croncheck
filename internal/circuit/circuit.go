// Package circuit implements a circuit-breaker per job.
// When a job exceeds the configured consecutive-failure threshold the circuit
// opens and further alerts are suppressed until it is manually reset or the
// job recovers.
package circuit

import (
	"fmt"
	"sync"
	"time"
)

// State represents the circuit state for a single job.
type State struct {
	Job              string    `json:"job"`
	ConsecutiveMisses int      `json:"consecutive_misses"`
	Open             bool      `json:"open"`
	OpenedAt         time.Time `json:"opened_at,omitempty"`
	Threshold        int       `json:"threshold"`
}

// Store tracks circuit-breaker state for every monitored job.
type Store struct {
	mu        sync.RWMutex
	entries   map[string]*State
	threshold int
}

// New creates a Store that opens the circuit after threshold consecutive misses.
func New(threshold int) *Store {
	if threshold <= 0 {
		threshold = 3
	}
	return &Store{
		entries:   make(map[string]*State),
		threshold: threshold,
	}
}

// RecordMiss increments the miss counter for job and opens the circuit when
// the threshold is reached. It returns true the first time the circuit opens.
func (s *Store) RecordMiss(job string, now time.Time) (justOpened bool) {
	s.mu.Lock()
	defer s.mu.Unlock()

	e := s.entry(job)
	e.ConsecutiveMisses++
	if !e.Open && e.ConsecutiveMisses >= e.Threshold {
		e.Open = true
		e.OpenedAt = now
		return true
	}
	return false
}

// Reset clears the miss counter and closes the circuit for job.
func (s *Store) Reset(job string) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	if _, ok := s.entries[job]; !ok {
		return fmt.Errorf("circuit: unknown job %q", job)
	}
	e := s.entries[job]
	e.ConsecutiveMisses = 0
	e.Open = false
	e.OpenedAt = time.Time{}
	return nil
}

// IsOpen reports whether the circuit for job is currently open.
func (s *Store) IsOpen(job string) bool {
	s.mu.RLock()
	defer s.mu.RUnlock()
	e, ok := s.entries[job]
	return ok && e.Open
}

// Get returns a copy of the State for job.
func (s *Store) Get(job string) (State, bool) {
	s.mu.RLock()
	defer s.mu.RUnlock()
	e, ok := s.entries[job]
	if !ok {
		return State{}, false
	}
	return *e, true
}

// All returns a snapshot of every tracked circuit.
func (s *Store) All() []State {
	s.mu.RLock()
	defer s.mu.RUnlock()
	out := make([]State, 0, len(s.entries))
	for _, e := range s.entries {
		out = append(out, *e)
	}
	return out
}

// entry returns (or creates) the mutable State for job. Caller must hold mu.
func (s *Store) entry(job string) *State {
	if e, ok := s.entries[job]; ok {
		return e
	}
	e := &State{Job: job, Threshold: s.threshold}
	s.entries[job] = e
	return e
}
