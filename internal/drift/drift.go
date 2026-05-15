// Package drift tracks schedule drift — the delta between when a job was
// expected to run and when it actually reported a heartbeat.
package drift

import (
	"errors"
	"sync"
	"time"
)

// Entry holds drift statistics for a single job.
type Entry struct {
	Job         string        `json:"job"`
	LastDrift   time.Duration `json:"last_drift_ns"`
	MaxDrift    time.Duration `json:"max_drift_ns"`
	AvgDrift    time.Duration `json:"avg_drift_ns"`
	SampleCount int           `json:"sample_count"`
	RecordedAt  time.Time     `json:"recorded_at"`
}

// Store records drift samples per job.
type Store struct {
	mu      sync.RWMutex
	entries map[string]*Entry
	now     func() time.Time
}

// New returns an initialised Store.
func New() *Store {
	return &Store{
		entries: make(map[string]*Entry),
		now:     time.Now,
	}
}

// Record registers the drift for job between expectedAt and actualAt.
func (s *Store) Record(job string, expectedAt, actualAt time.Time) error {
	if job == "" {
		return errors.New("drift: job name must not be empty")
	}
	delta := actualAt.Sub(expectedAt)
	if delta < 0 {
		delta = -delta
	}

	s.mu.Lock()
	defer s.mu.Unlock()

	e, ok := s.entries[job]
	if !ok {
		e = &Entry{Job: job}
		s.entries[job] = e
	}

	e.LastDrift = delta
	e.SampleCount++
	if delta > e.MaxDrift {
		e.MaxDrift = delta
	}
	// Running average.
	e.AvgDrift = time.Duration((int64(e.AvgDrift)*int64(e.SampleCount-1) + int64(delta)) / int64(e.SampleCount))
	e.RecordedAt = s.now()
	return nil
}

// Get returns the drift entry for job, or false if unknown.
func (s *Store) Get(job string) (Entry, bool) {
	s.mu.RLock()
	defer s.mu.RUnlock()
	e, ok := s.entries[job]
	if !ok {
		return Entry{}, false
	}
	return *e, true
}

// All returns a snapshot of all entries.
func (s *Store) All() []Entry {
	s.mu.RLock()
	defer s.mu.RUnlock()
	out := make([]Entry, 0, len(s.entries))
	for _, e := range s.entries {
		out = append(out, *e)
	}
	return out
}

// Reset removes the drift record for job.
func (s *Store) Reset(job string) {
	s.mu.Lock()
	defer s.mu.Unlock()
	delete(s.entries, job)
}
