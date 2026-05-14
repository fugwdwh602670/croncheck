// Package forecast predicts the next expected run time for a job
// based on its observed heartbeat history and configured schedule.
package forecast

import (
	"fmt"
	"sync"
	"time"
)

// Entry holds the forecast data for a single job.
type Entry struct {
	Job         string        `json:"job"`
	NextExpected time.Time    `json:"next_expected"`
	AvgInterval  time.Duration `json:"avg_interval_seconds"`
	SampleCount  int           `json:"sample_count"`
	UpdatedAt    time.Time     `json:"updated_at"`
}

// Store tracks interval samples and produces forecasts.
type Store struct {
	mu      sync.RWMutex
	entries map[string]*state
	maxSamples int
}

type state struct {
	lastSeen  time.Time
	intervals []time.Duration
	entry     Entry
}

// New creates a Store. maxSamples controls how many intervals are averaged.
func New(maxSamples int) *Store {
	if maxSamples <= 0 {
		maxSamples = 10
	}
	return &Store{entries: make(map[string]*state), maxSamples: maxSamples}
}

// Record registers a heartbeat tick for job and updates the forecast.
func (s *Store) Record(job string, now time.Time) error {
	if job == "" {
		return fmt.Errorf("forecast: job name must not be empty")
	}
	s.mu.Lock()
	defer s.mu.Unlock()

	st, ok := s.entries[job]
	if !ok {
		s.entries[job] = &state{lastSeen: now}
		return nil
	}

	interval := now.Sub(st.lastSeen)
	if interval > 0 {
		st.intervals = append(st.intervals, interval)
		if len(st.intervals) > s.maxSamples {
			st.intervals = st.intervals[len(st.intervals)-s.maxSamples:]
		}
	}
	st.lastSeen = now

	var avg time.Duration
	for _, iv := range st.intervals {
		avg += iv
	}
	if len(st.intervals) > 0 {
		avg /= time.Duration(len(st.intervals))
	}

	st.entry = Entry{
		Job:          job,
		NextExpected: now.Add(avg),
		AvgInterval:  avg,
		SampleCount:  len(st.intervals),
		UpdatedAt:    now,
	}
	return nil
}

// Get returns the current forecast for job, or false if unknown.
func (s *Store) Get(job string) (Entry, bool) {
	s.mu.RLock()
	defer s.mu.RUnlock()
	st, ok := s.entries[job]
	if !ok || st.entry.Job == "" {
		return Entry{}, false
	}
	return st.entry, true
}

// All returns a snapshot of all forecasts that have at least one interval.
func (s *Store) All() []Entry {
	s.mu.RLock()
	defer s.mu.RUnlock()
	out := make([]Entry, 0, len(s.entries))
	for _, st := range s.entries {
		if st.entry.Job != "" {
			out = append(out, st.entry)
		}
	}
	return out
}
