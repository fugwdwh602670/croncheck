// Package budget tracks alert budget consumption per job.
// An alert budget limits how many alerts may fire for a given job
// within a rolling time window, preventing notification storms.
package budget

import (
	"errors"
	"sync"
	"time"
)

// Entry holds the budget configuration and current consumption for a job.
type Entry struct {
	Job       string
	Limit     int
	Window    time.Duration
	Consumed  int
	WindowEnd time.Time
}

// Store manages alert budgets for jobs.
type Store struct {
	mu      sync.Mutex
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

// Set configures a budget for a job. Limit is the maximum number of alerts
// allowed within the given window duration.
func (s *Store) Set(job string, limit int, window time.Duration) error {
	if job == "" {
		return errors.New("job name must not be empty")
	}
	if limit <= 0 {
		return errors.New("limit must be positive")
	}
	if window <= 0 {
		return errors.New("window must be positive")
	}
	s.mu.Lock()
	defer s.mu.Unlock()
	s.entries[job] = &Entry{
		Job:       job,
		Limit:     limit,
		Window:    window,
		Consumed:  0,
		WindowEnd: s.now().Add(window),
	}
	return nil
}

// Allow returns true and increments the consumed counter if the job is within
// its budget. Returns false when the budget is exhausted for the current window.
// Jobs with no configured budget always return true.
func (s *Store) Allow(job string) bool {
	s.mu.Lock()
	defer s.mu.Unlock()
	e, ok := s.entries[job]
	if !ok {
		return true
	}
	now := s.now()
	if now.After(e.WindowEnd) {
		e.Consumed = 0
		e.WindowEnd = now.Add(e.Window)
	}
	if e.Consumed >= e.Limit {
		return false
	}
	e.Consumed++
	return true
}

// Get returns the current entry for a job, or false if none exists.
func (s *Store) Get(job string) (Entry, bool) {
	s.mu.Lock()
	defer s.mu.Unlock()
	e, ok := s.entries[job]
	if !ok {
		return Entry{}, false
	}
	return *e, true
}

// Remove deletes the budget entry for a job.
func (s *Store) Remove(job string) {
	s.mu.Lock()
	defer s.mu.Unlock()
	delete(s.entries, job)
}

// All returns a snapshot of all configured budget entries.
func (s *Store) All() []Entry {
	s.mu.Lock()
	defer s.mu.Unlock()
	out := make([]Entry, 0, len(s.entries))
	for _, e := range s.entries {
		out = append(out, *e)
	}
	return out
}
