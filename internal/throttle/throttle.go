// Package throttle limits how many alerts can be sent for a job within a
// rolling time window, preventing notification storms.
package throttle

import (
	"errors"
	"sync"
	"time"
)

// entry tracks the alert count and window start for a single job.
type entry struct {
	count     int
	windowEnd time.Time
}

// Store holds per-job throttle configuration and state.
type Store struct {
	mu      sync.Mutex
	limits  map[string]Config
	state   map[string]*entry
	nowFunc func() time.Time
}

// Config defines the throttle policy for a job.
type Config struct {
	MaxAlerts int
	Window    time.Duration
}

// New returns an initialised Store.
func New() *Store {
	return &Store{
		limits:  make(map[string]Config),
		state:   make(map[string]*entry),
		nowFunc: time.Now,
	}
}

// Set configures the throttle policy for a job.
func (s *Store) Set(job string, cfg Config) error {
	if job == "" {
		return errors.New("job name must not be empty")
	}
	if cfg.MaxAlerts <= 0 {
		return errors.New("max_alerts must be positive")
	}
	if cfg.Window <= 0 {
		return errors.New("window must be positive")
	}
	s.mu.Lock()
	defer s.mu.Unlock()
	s.limits[job] = cfg
	return nil
}

// Get returns the throttle config for a job and whether it exists.
func (s *Store) Get(job string) (Config, bool) {
	s.mu.Lock()
	defer s.mu.Unlock()
	cfg, ok := s.limits[job]
	return cfg, ok
}

// Remove deletes the throttle config and state for a job.
func (s *Store) Remove(job string) {
	s.mu.Lock()
	defer s.mu.Unlock()
	delete(s.limits, job)
	delete(s.state, job)
}

// Allow reports whether an alert for job is permitted under its throttle
// policy. If no policy is configured, Allow always returns true.
func (s *Store) Allow(job string) bool {
	s.mu.Lock()
	defer s.mu.Unlock()
	cfg, ok := s.limits[job]
	if !ok {
		return true
	}
	now := s.nowFunc()
	e, exists := s.state[job]
	if !exists || now.After(e.windowEnd) {
		s.state[job] = &entry{count: 1, windowEnd: now.Add(cfg.Window)}
		return true
	}
	if e.count >= cfg.MaxAlerts {
		return false
	}
	e.count++
	return true
}

// All returns a snapshot of all configured throttle policies.
func (s *Store) All() map[string]Config {
	s.mu.Lock()
	defer s.mu.Unlock()
	out := make(map[string]Config, len(s.limits))
	for k, v := range s.limits {
		out[k] = v
	}
	return out
}
