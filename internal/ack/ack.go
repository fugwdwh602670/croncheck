// Package ack provides acknowledgement tracking for missed cron job alerts.
// An acknowledged job suppresses repeat alerts until the next missed run.
package ack

import (
	"sync"
	"time"
)

// Acknowledgement records when a job alert was acknowledged and by whom.
type Acknowledgement struct {
	Job       string    `json:"job"`
	AckedAt   time.Time `json:"acked_at"`
	AckedBy   string    `json:"acked_by"`
	ExpiresAt time.Time `json:"expires_at"`
}

// Store holds acknowledgements for jobs.
type Store struct {
	mu   sync.RWMutex
	acks map[string]Acknowledgement
}

// New returns an initialised acknowledgement Store.
func New() *Store {
	return &Store{acks: make(map[string]Acknowledgement)}
}

// Acknowledge records an acknowledgement for the given job, valid for duration d.
// ackedBy is a free-form string identifying who acknowledged the alert.
func (s *Store) Acknowledge(job, ackedBy string, d time.Duration) {
	s.mu.Lock()
	defer s.mu.Unlock()
	now := time.Now()
	s.acks[job] = Acknowledgement{
		Job:       job,
		AckedAt:   now,
		AckedBy:   ackedBy,
		ExpiresAt: now.Add(d),
	}
}

// IsAcknowledged reports whether the job currently has an active acknowledgement.
func (s *Store) IsAcknowledged(job string) bool {
	s.mu.RLock()
	defer s.mu.RUnlock()
	a, ok := s.acks[job]
	if !ok {
		return false
	}
	return time.Now().Before(a.ExpiresAt)
}

// Remove deletes the acknowledgement for the given job.
func (s *Store) Remove(job string) {
	s.mu.Lock()
	defer s.mu.Unlock()
	delete(s.acks, job)
}

// All returns a snapshot of all currently active acknowledgements.
func (s *Store) All() []Acknowledgement {
	s.mu.RLock()
	defer s.mu.RUnlock()
	now := time.Now()
	out := make([]Acknowledgement, 0, len(s.acks))
	for _, a := range s.acks {
		if now.Before(a.ExpiresAt) {
			out = append(out, a)
		}
	}
	return out
}
