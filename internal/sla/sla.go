// Package sla tracks SLA compliance for monitored cron jobs.
// It records whether jobs have met their expected execution windows
// and exposes a per-job compliance percentage.
package sla

import (
	"sync"
	"time"
)

// Entry holds SLA tracking data for a single job.
type Entry struct {
	JobName    string    `json:"job_name"`
	Total      int       `json:"total_checks"`
	Missed     int       `json:"missed_checks"`
	Compliance float64   `json:"compliance_pct"`
	WindowStart time.Time `json:"window_start"`
}

// Store tracks SLA compliance per job.
type Store struct {
	mu      sync.RWMutex
	entries map[string]*Entry
}

// New creates an empty SLA store.
func New() *Store {
	return &Store{
		entries: make(map[string]*Entry),
	}
}

// RecordCheck records a single check result for a job.
// missed=true means the job failed to run within its expected window.
func (s *Store) RecordCheck(jobName string, missed bool) {
	s.mu.Lock()
	defer s.mu.Unlock()

	e, ok := s.entries[jobName]
	if !ok {
		e = &Entry{
			JobName:     jobName,
			WindowStart: time.Now().UTC(),
		}
		s.entries[jobName] = e
	}

	e.Total++
	if missed {
		e.Missed++
	}
	e.Compliance = compliance(e.Total, e.Missed)
}

// Get returns the SLA entry for a job, and whether it was found.
func (s *Store) Get(jobName string) (Entry, bool) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	e, ok := s.entries[jobName]
	if !ok {
		return Entry{}, false
	}
	return *e, true
}

// All returns a snapshot of all SLA entries.
func (s *Store) All() []Entry {
	s.mu.RLock()
	defer s.mu.RUnlock()

	out := make([]Entry, 0, len(s.entries))
	for _, e := range s.entries {
		out = append(out, *e)
	}
	return out
}

// Reset clears SLA data for a specific job.
func (s *Store) Reset(jobName string) {
	s.mu.Lock()
	defer s.mu.Unlock()
	delete(s.entries, jobName)
}

func compliance(total, missed int) float64 {
	if total == 0 {
		return 100.0
	}
	return float64(total-missed) / float64(total) * 100.0
}
