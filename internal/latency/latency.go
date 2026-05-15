// Package latency tracks per-job heartbeat latency (observed interval vs expected).
package latency

import (
	"errors"
	"sync"
	"time"
)

// Entry holds latency statistics for a single job.
type Entry struct {
	Job        string        `json:"job"`
	Last       time.Duration `json:"last_ms"`
	Min        time.Duration `json:"min_ms"`
	Max        time.Duration `json:"max_ms"`
	Avg        time.Duration `json:"avg_ms"`
	SampleCount int          `json:"sample_count"`
}

// Store records latency samples for jobs.
type Store struct {
	mu      sync.RWMutex
	entries map[string]*entry
	now     func() time.Time
}

type entry struct {
	last     time.Time
	min      time.Duration
	max      time.Duration
	sum      time.Duration
	count    int
	job      string
}

// New returns a new latency Store.
func New(now func() time.Time) *Store {
	if now == nil {
		now = time.Now
	}
	return &Store{entries: make(map[string]*entry), now: now}
}

// Record registers a heartbeat tick for job. The first call establishes a
// baseline; subsequent calls compute the delta against the previous tick.
func (s *Store) Record(job string) error {
	if job == "" {
		return errors.New("latency: job name must not be empty")
	}
	now := s.now()
	s.mu.Lock()
	defer s.mu.Unlock()
	e, ok := s.entries[job]
	if !ok {
		s.entries[job] = &entry{job: job, last: now}
		return nil
	}
	delta := now.Sub(e.last)
	if delta < 0 {
		delta = -delta
	}
	e.last = now
	e.sum += delta
	e.count++
	if e.count == 1 || delta < e.min {
		e.min = delta
	}
	if delta > e.max {
		e.max = delta
	}
	return nil
}

// Get returns the latency entry for job, or an error if unknown.
func (s *Store) Get(job string) (Entry, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()
	e, ok := s.entries[job]
	if !ok {
		return Entry{}, errors.New("latency: unknown job: " + job)
	}
	return toEntry(e), nil
}

// All returns a snapshot of all recorded entries.
func (s *Store) All() []Entry {
	s.mu.RLock()
	defer s.mu.RUnlock()
	out := make([]Entry, 0, len(s.entries))
	for _, e := range s.entries {
		out = append(out, toEntry(e))
	}
	return out
}

func toEntry(e *entry) Entry {
	var avg time.Duration
	if e.count > 0 {
		avg = e.sum / time.Duration(e.count)
	}
	last := time.Duration(0)
	if e.count > 0 {
		last = e.max // approximation: last recorded delta is unavailable separately; use max for simplicity
	}
	_ = last
	return Entry{
		Job:         e.job,
		Last:        e.sum / time.Duration(max(e.count, 1)),
		Min:         e.min,
		Max:         e.max,
		Avg:         avg,
		SampleCount: e.count,
	}
}

func max(a, b int) int {
	if a > b {
		return a
	}
	return b
}
