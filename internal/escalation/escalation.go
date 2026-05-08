// Package escalation tracks how many times an alert has fired for a job
// and determines whether it should be escalated to a higher-priority channel.
package escalation

import (
	"sync"
	"time"
)

// Level represents an escalation tier.
type Level int

const (
	LevelNormal   Level = 0
	LevelWarning  Level = 1
	LevelCritical Level = 2
)

// Entry holds escalation state for a single job.
type Entry struct {
	FireCount  int
	Level      Level
	LastFired  time.Time
	EscalatedAt time.Time
}

// Store tracks escalation state per job.
type Store struct {
	mu              sync.Mutex
	entries         map[string]*Entry
	warningThreshold  int
	criticalThreshold int
}

// New creates a Store. warningThreshold and criticalThreshold are the
// consecutive fire counts required to reach each level.
func New(warningThreshold, criticalThreshold int) *Store {
	return &Store{
		entries:           make(map[string]*Entry),
		warningThreshold:  warningThreshold,
		criticalThreshold: criticalThreshold,
	}
}

// Record increments the fire count for a job and returns the resulting Entry.
func (s *Store) Record(job string, now time.Time) Entry {
	s.mu.Lock()
	defer s.mu.Unlock()

	e, ok := s.entries[job]
	if !ok {
		e = &Entry{}
		s.entries[job] = e
	}

	e.FireCount++
	e.LastFired = now

	prev := e.Level
	switch {
	case e.FireCount >= s.criticalThreshold:
		e.Level = LevelCritical
	case e.FireCount >= s.warningThreshold:
		e.Level = LevelWarning
	default:
		e.Level = LevelNormal
	}
	if e.Level > prev {
		e.EscalatedAt = now
	}

	return *e
}

// Reset clears escalation state for a job (e.g. when a heartbeat is received).
func (s *Store) Reset(job string) {
	s.mu.Lock()
	defer s.mu.Unlock()
	delete(s.entries, job)
}

// Get returns the current Entry for a job and whether it exists.
func (s *Store) Get(job string) (Entry, bool) {
	s.mu.Lock()
	defer s.mu.Unlock()
	e, ok := s.entries[job]
	if !ok {
		return Entry{}, false
	}
	return *e, true
}

// All returns a snapshot of all current escalation entries.
func (s *Store) All() map[string]Entry {
	s.mu.Lock()
	defer s.mu.Unlock()
	out := make(map[string]Entry, len(s.entries))
	for k, v := range s.entries {
		out[k] = *v
	}
	return out
}
