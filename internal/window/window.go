// Package window tracks the expected execution window for cron jobs,
// allowing croncheck to detect jobs that ran outside their scheduled window.
package window

import (
	"errors"
	"sync"
	"time"
)

// Entry defines the allowed execution window for a job.
type Entry struct {
	Job      string
	Start    time.Duration // offset from the scheduled time
	End      time.Duration // offset from the scheduled time
	UpdatedAt time.Time
}

// Store holds window configurations keyed by job name.
type Store struct {
	mu      sync.RWMutex
	entries map[string]Entry
}

// New returns an initialised Store.
func New() *Store {
	return &Store{entries: make(map[string]Entry)}
}

// Set registers or updates the execution window for a job.
func (s *Store) Set(job string, start, end time.Duration) error {
	if job == "" {
		return errors.New("job name must not be empty")
	}
	if end <= start {
		return errors.New("end must be greater than start")
	}
	s.mu.Lock()
	defer s.mu.Unlock()
	s.entries[job] = Entry{
		Job:       job,
		Start:     start,
		End:       end,
		UpdatedAt: time.Now(),
	}
	return nil
}

// Get returns the window entry for a job.
func (s *Store) Get(job string) (Entry, bool) {
	s.mu.RLock()
	defer s.mu.RUnlock()
	e, ok := s.entries[job]
	return e, ok
}

// Remove deletes the window configuration for a job.
func (s *Store) Remove(job string) {
	s.mu.Lock()
	defer s.mu.Unlock()
	delete(s.entries, job)
}

// All returns a snapshot of all window entries.
func (s *Store) All() []Entry {
	s.mu.RLock()
	defer s.mu.RUnlock()
	out := make([]Entry, 0, len(s.entries))
	for _, e := range s.entries {
		out = append(out, e)
	}
	return out
}

// InWindow reports whether t falls within the window defined for job,
// relative to the scheduled time sched. Returns true when no window is
// configured (permissive default).
func (s *Store) InWindow(job string, sched, t time.Time) bool {
	e, ok := s.Get(job)
	if !ok {
		return true
	}
	earliest := sched.Add(e.Start)
	latest := sched.Add(e.End)
	return !t.Before(earliest) && !t.After(latest)
}
