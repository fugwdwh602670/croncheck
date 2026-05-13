// Package replay provides a store for tracking manual job replay requests.
// A replay entry records when a job was requested to re-run and whether it
// has been acknowledged by the scheduler.
package replay

import (
	"errors"
	"sync"
	"time"
)

// Entry represents a single replay request for a job.
type Entry struct {
	Job         string    `json:"job"`
	RequestedAt time.Time `json:"requested_at"`
	Acked        bool      `json:"acked"`
	AckedAt      time.Time `json:"acked_at,omitempty"`
}

// Store holds pending and completed replay requests.
type Store struct {
	mu      sync.RWMutex
	entries map[string]Entry
	now     func() time.Time
}

// New returns an initialised replay Store.
func New() *Store {
	return &Store{
		entries: make(map[string]Entry),
		now:     time.Now,
	}
}

// Request records a new replay request for the given job.
// Returns an error if the job name is empty.
func (s *Store) Request(job string) error {
	if job == "" {
		return errors.New("replay: job name must not be empty")
	}
	s.mu.Lock()
	defer s.mu.Unlock()
	s.entries[job] = Entry{
		Job:         job,
		RequestedAt: s.now(),
		Acked:        false,
	}
	return nil
}

// Ack marks a replay request as acknowledged.
// Returns an error if the job has no pending request.
func (s *Store) Ack(job string) error {
	s.mu.Lock()
	defer s.mu.Unlock()
	e, ok := s.entries[job]
	if !ok {
		return errors.New("replay: no request found for job " + job)
	}
	e.Acked = true
	e.AckedAt = s.now()
	s.entries[job] = e
	return nil
}

// Get returns the replay entry for a job, or false if none exists.
func (s *Store) Get(job string) (Entry, bool) {
	s.mu.RLock()
	defer s.mu.RUnlock()
	e, ok := s.entries[job]
	return e, ok
}

// Remove deletes the replay entry for a job.
func (s *Store) Remove(job string) {
	s.mu.Lock()
	defer s.mu.Unlock()
	delete(s.entries, job)
}

// All returns a snapshot of all replay entries.
func (s *Store) All() []Entry {
	s.mu.RLock()
	defer s.mu.RUnlock()
	out := make([]Entry, 0, len(s.entries))
	for _, e := range s.entries {
		out = append(out, e)
	}
	return out
}
