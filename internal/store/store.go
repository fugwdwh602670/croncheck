package store

import (
	"sync"
	"time"
)

// JobStatus holds the last heartbeat and alert state for a job.
type JobStatus struct {
	Name        string
	LastSeen    time.Time
	MissedCount int
	Alerted     bool
}

// Store is an in-memory thread-safe store for job statuses.
type Store struct {
	mu   sync.RWMutex
	jobs map[string]*JobStatus
}

// New creates an empty Store.
func New() *Store {
	return &Store{
		jobs: make(map[string]*JobStatus),
	}
}

// RecordHeartbeat updates (or creates) the last-seen time for a job.
func (s *Store) RecordHeartbeat(name string) {
	s.mu.Lock()
	defer s.mu.Unlock()

	if js, ok := s.jobs[name]; ok {
		js.LastSeen = time.Now()
		js.MissedCount = 0
		js.Alerted = false
		return
	}
	s.jobs[name] = &JobStatus{
		Name:     name,
		LastSeen: time.Now(),
	}
}

// Get returns a copy of the JobStatus for the given job name.
// The second return value is false if the job is unknown.
func (s *Store) Get(name string) (JobStatus, bool) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	js, ok := s.jobs[name]
	if !ok {
		return JobStatus{}, false
	}
	return *js, true
}

// IncrementMissed increments the missed counter and marks the job as alerted.
func (s *Store) IncrementMissed(name string) {
	s.mu.Lock()
	defer s.mu.Unlock()

	if js, ok := s.jobs[name]; ok {
		js.MissedCount++
		js.Alerted = true
	}
}

// All returns a snapshot of all job statuses.
func (s *Store) All() []JobStatus {
	s.mu.RLock()
	defer s.mu.RUnlock()

	out := make([]JobStatus, 0, len(s.jobs))
	for _, js := range s.jobs {
		out = append(out, *js)
	}
	return out
}
