// Package checkpoint tracks the last successful completion time
// for each monitored cron job, independent of the heartbeat store.
package checkpoint

import (
	"sync"
	"time"
)

// Entry holds checkpoint data for a single job.
type Entry struct {
	JobName   string    `json:"job_name"`
	LastOK    time.Time `json:"last_ok"`
	UpdatedAt time.Time `json:"updated_at"`
}

// Store persists the last-known-good timestamp for each job.
type Store struct {
	mu      sync.RWMutex
	entries map[string]Entry
	now     func() time.Time
}

// New returns an initialised Store.
func New() *Store {
	return &Store{
		entries: make(map[string]Entry),
		now:     time.Now,
	}
}

// Record marks job as successfully completed at the current time.
func (s *Store) Record(jobName string) {
	s.mu.Lock()
	defer s.mu.Unlock()
	now := s.now()
	s.entries[jobName] = Entry{
		JobName:   jobName,
		LastOK:    now,
		UpdatedAt: now,
	}
}

// Get returns the checkpoint entry for jobName and whether it exists.
func (s *Store) Get(jobName string) (Entry, bool) {
	s.mu.RLock()
	defer s.mu.RUnlock()
	e, ok := s.entries[jobName]
	return e, ok
}

// Remove deletes the checkpoint entry for jobName.
func (s *Store) Remove(jobName string) {
	s.mu.Lock()
	defer s.mu.Unlock()
	delete(s.entries, jobName)
}

// All returns a snapshot of all checkpoint entries.
func (s *Store) All() []Entry {
	s.mu.RLock()
	defer s.mu.RUnlock()
	out := make([]Entry, 0, len(s.entries))
	for _, e := range s.entries {
		out = append(out, e)
	}
	return out
}
