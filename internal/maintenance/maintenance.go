// Package maintenance tracks scheduled maintenance windows for jobs.
// During a maintenance window, missed/failed alerts are suppressed.
package maintenance

import (
	"errors"
	"sync"
	"time"
)

// Window represents a scheduled maintenance period for a job.
type Window struct {
	Job       string    `json:"job"`
	StartsAt  time.Time `json:"starts_at"`
	EndsAt    time.Time `json:"ends_at"`
	CreatedAt time.Time `json:"created_at"`
}

// Store holds active maintenance windows.
type Store struct {
	mu      sync.RWMutex
	windows map[string]Window
	now     func() time.Time
}

// New creates a new maintenance Store.
func New() *Store {
	return &Store{
		windows: make(map[string]Window),
		now:     time.Now,
	}
}

// Set registers a maintenance window for the given job.
func (s *Store) Set(job string, duration time.Duration) error {
	if job == "" {
		return errors.New("job name must not be empty")
	}
	if duration <= 0 {
		return errors.New("duration must be positive")
	}
	now := s.now()
	s.mu.Lock()
	defer s.mu.Unlock()
	s.windows[job] = Window{
		Job:       job,
		StartsAt:  now,
		EndsAt:    now.Add(duration),
		CreatedAt: now,
	}
	return nil
}

// IsActive returns true if the job currently has an active maintenance window.
func (s *Store) IsActive(job string) bool {
	s.mu.RLock()
	defer s.mu.RUnlock()
	w, ok := s.windows[job]
	if !ok {
		return false
	}
	now := s.now()
	return now.Before(w.EndsAt)
}

// Remove deletes the maintenance window for a job.
func (s *Store) Remove(job string) {
	s.mu.Lock()
	defer s.mu.Unlock()
	delete(s.windows, job)
}

// All returns a snapshot of all windows, active or expired.
func (s *Store) All() []Window {
	s.mu.RLock()
	defer s.mu.RUnlock()
	out := make([]Window, 0, len(s.windows))
	for _, w := range s.windows {
		out = append(out, w)
	}
	return out
}
