// Package digest provides a periodic summary digest of job health,
// aggregating missed counts, SLA status, and escalation levels.
package digest

import (
	"sync"
	"time"
)

// Entry holds a snapshot of a single job's digest state.
type Entry struct {
	Job         string    `json:"job"`
	MissedCount int       `json:"missed_count"`
	Healthy     bool      `json:"healthy"`
	LastSeen    time.Time `json:"last_seen,omitempty"`
	CapturedAt  time.Time `json:"captured_at"`
}

// Snapshotter is the interface required to collect per-job state.
type Snapshotter interface {
	AllJobs() []JobState
}

// JobState carries the minimal fields needed to build a digest entry.
type JobState struct {
	Name        string
	MissedCount int
	Healthy     bool
	LastSeen    time.Time
}

// Store holds the most-recently generated digest.
type Store struct {
	mu      sync.RWMutex
	entries []Entry
	builtAt time.Time
}

// New returns an empty digest Store.
func New() *Store {
	return &Store{}
}

// Build replaces the stored digest with a fresh snapshot derived from src.
func (s *Store) Build(src []JobState) {
	now := time.Now().UTC()
	entries := make([]Entry, 0, len(src))
	for _, js := range src {
		entries = append(entries, Entry{
			Job:         js.Name,
			MissedCount: js.MissedCount,
			Healthy:     js.Healthy,
			LastSeen:    js.LastSeen,
			CapturedAt:  now,
		})
	}
	s.mu.Lock()
	defer s.mu.Unlock()
	s.entries = entries
	s.builtAt = now
}

// All returns a copy of the current digest entries.
func (s *Store) All() []Entry {
	s.mu.RLock()
	defer s.mu.RUnlock()
	out := make([]Entry, len(s.entries))
	copy(out, s.entries)
	return out
}

// BuiltAt returns the timestamp of the last Build call.
func (s *Store) BuiltAt() time.Time {
	s.mu.RLock()
	defer s.mu.RUnlock()
	return s.builtAt
}
