// Package backoff provides exponential back-off tracking for repeated
// missed or failed cron job executions. When a job continues to miss
// its schedule, alerts can be spaced further apart to avoid flooding
// the notification channel.
package backoff

import (
	"errors"
	"math"
	"sync"
	"time"
)

// Entry holds the current back-off state for a single job.
type Entry struct {
	// Attempt is the number of consecutive misses recorded so far.
	Attempt int
	// NextAllowed is the earliest time the next alert may be sent.
	NextAllowed time.Time
}

// Store tracks exponential back-off state for cron jobs.
type Store struct {
	mu      sync.Mutex
	entries map[string]*Entry

	// baseDelay is the delay applied after the first miss.
	baseDelay time.Duration
	// maxDelay caps the computed delay so it does not grow unbounded.
	maxDelay time.Duration
	// now is used for time injection in tests.
	now func() time.Time
}

// New creates a Store with the given base and maximum delays.
// Typical values: baseDelay = 1 minute, maxDelay = 2 hours.
func New(baseDelay, maxDelay time.Duration) (*Store, error) {
	if baseDelay <= 0 {
		return nil, errors.New("backoff: baseDelay must be positive")
	}
	if maxDelay < baseDelay {
		return nil, errors.New("backoff: maxDelay must be >= baseDelay")
	}
	return &Store{
		entries:   make(map[string]*Entry),
		baseDelay: baseDelay,
		maxDelay:  maxDelay,
		now:       time.Now,
	}, nil
}

// Record increments the miss counter for job and updates NextAllowed
// using exponential back-off: delay = baseDelay * 2^(attempt-1), capped
// at maxDelay. It returns the resulting Entry.
func (s *Store) Record(job string) (Entry, error) {
	if job == "" {
		return Entry{}, errors.New("backoff: job name must not be empty")
	}
	s.mu.Lock()
	defer s.mu.Unlock()

	e, ok := s.entries[job]
	if !ok {
		e = &Entry{}
		s.entries[job] = e
	}
	e.Attempt++

	// delay = baseDelay * 2^(attempt-1)
	exponent := float64(e.Attempt - 1)
	delay := time.Duration(float64(s.baseDelay) * math.Pow(2, exponent))
	if delay > s.maxDelay {
		delay = s.maxDelay
	}
	e.NextAllowed = s.now().Add(delay)
	return *e, nil
}

// IsReady reports whether the back-off window for job has elapsed and
// a new alert is permitted. Unknown jobs are considered ready.
func (s *Store) IsReady(job string) bool {
	s.mu.Lock()
	defer s.mu.Unlock()

	e, ok := s.entries[job]
	if !ok {
		return true
	}
	return s.now().After(e.NextAllowed) || s.now().Equal(e.NextAllowed)
}

// Reset clears the back-off state for job, allowing the next alert
// immediately. This should be called when a job recovers.
func (s *Store) Reset(job string) {
	s.mu.Lock()
	defer s.mu.Unlock()
	delete(s.entries, job)
}

// Get returns the current Entry for job and a boolean indicating
// whether an entry exists.
func (s *Store) Get(job string) (Entry, bool) {
	s.mu.Lock()
	defer s.mu.Unlock()
	e, ok := s.entries[job]
	if !ok {
		return Entry{}, false
	}
	return *e, true
}

// All returns a snapshot of all current back-off entries keyed by job name.
func (s *Store) All() map[string]Entry {
	s.mu.Lock()
	defer s.mu.Unlock()
	out := make(map[string]Entry, len(s.entries))
	for k, v := range s.entries {
		out[k] = *v
	}
	return out
}
