// Package deadletter stores alerts that failed to deliver so they can be
// inspected and retried later.
package deadletter

import (
	"sync"
	"time"
)

// Entry represents a single undelivered alert.
type Entry struct {
	Job       string    `json:"job"`
	Reason    string    `json:"reason"`
	Payload   string    `json:"payload"`
	FailedAt  time.Time `json:"failed_at"`
	Attempts  int       `json:"attempts"`
}

// Store holds dead-letter entries in memory.
type Store struct {
	mu      sync.Mutex
	entries []Entry
	limit   int
}

// New creates a Store that retains at most limit entries (oldest evicted first).
// If limit <= 0 it defaults to 100.
func New(limit int) *Store {
	if limit <= 0 {
		limit = 100
	}
	return &Store{limit: limit}
}

// Add appends a new dead-letter entry, evicting the oldest if the store is full.
func (s *Store) Add(job, reason, payload string, attempts int) {
	s.mu.Lock()
	defer s.mu.Unlock()

	entry := Entry{
		Job:      job,
		Reason:   reason,
		Payload:  payload,
		FailedAt: time.Now().UTC(),
		Attempts: attempts,
	}

	if len(s.entries) >= s.limit {
		s.entries = s.entries[1:]
	}
	s.entries = append(s.entries, entry)
}

// All returns a snapshot of all current entries.
func (s *Store) All() []Entry {
	s.mu.Lock()
	defer s.mu.Unlock()

	out := make([]Entry, len(s.entries))
	copy(out, s.entries)
	return out
}

// Remove deletes all entries for the given job name.
func (s *Store) Remove(job string) {
	s.mu.Lock()
	defer s.mu.Unlock()

	filtered := s.entries[:0]
	for _, e := range s.entries {
		if e.Job != job {
			filtered = append(filtered, e)
		}
	}
	s.entries = filtered
}

// Count returns the total number of stored entries.
func (s *Store) Count() int {
	s.mu.Lock()
	defer s.mu.Unlock()
	return len(s.entries)
}
