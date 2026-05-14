// Package snapshot provides point-in-time capture of all job states,
// useful for debugging, auditing, and backup/restore workflows.
package snapshot

import (
	"sync"
	"time"
)

// JobState holds a captured state for a single job at snapshot time.
type JobState struct {
	Job         string    `json:"job"`
	Healthy     bool      `json:"healthy"`
	MissedCount int       `json:"missed_count"`
	LastSeen    time.Time `json:"last_seen"`
	CapturedAt  time.Time `json:"captured_at"`
}

// Snapshot represents a full capture of all job states at a moment in time.
type Snapshot struct {
	ID         string      `json:"id"`
	CapturedAt time.Time   `json:"captured_at"`
	Jobs       []JobState  `json:"jobs"`
}

// Source is the interface that provides job states for snapshotting.
type Source interface {
	AllStates() []JobState
}

// Store holds recent snapshots in memory.
type Store struct {
	mu        sync.RWMutex
	snapshots []Snapshot
	limit     int
	now       func() time.Time
}

// New creates a new snapshot Store with the given retention limit.
func New(limit int) *Store {
	if limit <= 0 {
		limit = 10
	}
	return &Store{
		limit: limit,
		now:   time.Now,
	}
}

// Capture takes a snapshot from the provided source and stores it.
func (s *Store) Capture(id string, src Source) Snapshot {
	now := s.now()
	jobs := src.AllStates()
	for i := range jobs {
		jobs[i].CapturedAt = now
	}
	snap := Snapshot{
		ID:         id,
		CapturedAt: now,
		Jobs:       jobs,
	}
	s.mu.Lock()
	defer s.mu.Unlock()
	s.snapshots = append(s.snapshots, snap)
	if len(s.snapshots) > s.limit {
		s.snapshots = s.snapshots[len(s.snapshots)-s.limit:]
	}
	return snap
}

// Get returns the snapshot with the given ID, or false if not found.
func (s *Store) Get(id string) (Snapshot, bool) {
	s.mu.RLock()
	defer s.mu.RUnlock()
	for _, snap := range s.snapshots {
		if snap.ID == id {
			return snap, true
		}
	}
	return Snapshot{}, false
}

// All returns all stored snapshots, newest first.
func (s *Store) All() []Snapshot {
	s.mu.RLock()
	defer s.mu.RUnlock()
	out := make([]Snapshot, len(s.snapshots))
	for i, snap := range s.snapshots {
		out[len(s.snapshots)-1-i] = snap
	}
	return out
}
