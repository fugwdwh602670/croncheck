// Package trend tracks miss-rate trends for cron jobs over a sliding window.
package trend

import (
	"fmt"
	"sync"
	"time"
)

// Entry holds trend data for a single job.
type Entry struct {
	Job       string        `json:"job"`
	Window    time.Duration `json:"window_seconds"`
	Total     int           `json:"total_checks"`
	Misses    int           `json:"missed_checks"`
	MissRate  float64       `json:"miss_rate"`
	UpdatedAt time.Time     `json:"updated_at"`
}

type record struct {
	events []event
}

type event struct {
	at     time.Time
	missed bool
}

// Store records check outcomes and computes miss-rate trends.
type Store struct {
	mu     sync.Mutex
	data   map[string]*record
	window time.Duration
	now    func() time.Time
}

// New creates a Store with the given sliding window duration.
func New(window time.Duration) *Store {
	return &Store{
		data:   make(map[string]*record),
		window: window,
		now:    time.Now,
	}
}

// Record appends a check outcome for the given job.
func (s *Store) Record(job string, missed bool) error {
	if job == "" {
		return fmt.Errorf("trend: job name must not be empty")
	}
	s.mu.Lock()
	defer s.mu.Unlock()
	if _, ok := s.data[job]; !ok {
		s.data[job] = &record{}
	}
	s.data[job].events = append(s.data[job].events, event{at: s.now(), missed: missed})
	return nil
}

// Get returns the current trend entry for a job, pruning stale events first.
func (s *Store) Get(job string) (Entry, bool) {
	s.mu.Lock()
	defer s.mu.Unlock()
	r, ok := s.data[job]
	if !ok {
		return Entry{}, false
	}
	s.prune(r)
	return s.toEntry(job, r), true
}

// All returns trend entries for every tracked job.
func (s *Store) All() []Entry {
	s.mu.Lock()
	defer s.mu.Unlock()
	out := make([]Entry, 0, len(s.data))
	for job, r := range s.data {
		s.prune(r)
		out = append(out, s.toEntry(job, r))
	}
	return out
}

func (s *Store) prune(r *record) {
	cutoff := s.now().Add(-s.window)
	start := 0
	for start < len(r.events) && r.events[start].at.Before(cutoff) {
		start++
	}
	r.events = r.events[start:]
}

func (s *Store) toEntry(job string, r *record) Entry {
	total := len(r.events)
	misses := 0
	for _, e := range r.events {
		if e.missed {
			misses++
		}
	}
	var rate float64
	if total > 0 {
		rate = float64(misses) / float64(total)
	}
	return Entry{
		Job:       job,
		Window:    s.window,
		Total:     total,
		Misses:    misses,
		MissRate:  rate,
		UpdatedAt: s.now(),
	}
}
