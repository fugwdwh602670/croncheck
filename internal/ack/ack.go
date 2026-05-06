// Package ack provides acknowledgement tracking for missed cron job alerts.
// An acknowledged job will suppress further alerts until the ack expires.
package ack

import (
	"sync"
	"time"
)

// Ack represents an active acknowledgement for a job.
type Ack struct {
	JobName   string    `json:"job_name"`
	ExpiresAt time.Time `json:"expires_at"`
	Reason    string    `json:"reason,omitempty"`
}

// Store holds active acknowledgements.
type Store struct {
	mu   sync.RWMutex
	acks map[string]Ack
	now  func() time.Time
}

// New creates a new ack Store.
func New() *Store {
	return &Store{
		acks: make(map[string]Ack),
		now:  time.Now,
	}
}

// Add records an acknowledgement for the given job lasting duration d.
func (s *Store) Add(jobName, reason string, d time.Duration) {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.acks[jobName] = Ack{
		JobName:   jobName,
		ExpiresAt: s.now().Add(d),
		Reason:    reason,
	}
}

// IsAcked reports whether the job currently has an active acknowledgement.
func (s *Store) IsAcked(jobName string) bool {
	s.mu.RLock()
	defer s.mu.RUnlock()
	a, ok := s.acks[jobName]
	if !ok {
		return false
	}
	return s.now().Before(a.ExpiresAt)
}

// Remove deletes an acknowledgement for the given job.
func (s *Store) Remove(jobName string) {
	s.mu.Lock()
	defer s.mu.Unlock()
	delete(s.acks, jobName)
}

// All returns a snapshot of currently active acknowledgements.
func (s *Store) All() []Ack {
	s.mu.RLock()
	defer s.mu.RUnlock()
	now := s.now()
	out := make([]Ack, 0, len(s.acks))
	for _, a := range s.acks {
		if now.Before(a.ExpiresAt) {
			out = append(out, a)
		}
	}
	return out
}
