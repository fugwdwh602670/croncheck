// Package fingerprint generates stable identifiers for alert events,
// allowing deduplication of repeated notifications for the same condition.
package fingerprint

import (
	"crypto/sha256"
	"fmt"
	"sync"
	"time"
)

// Entry records when a fingerprint was first seen and how many times it has fired.
type Entry struct {
	Fingerprint string
	JobName     string
	Reason      string
	FirstSeen   time.Time
	LastSeen    time.Time
	Count       int
}

// Store tracks alert fingerprints to detect duplicates.
type Store struct {
	mu      sync.Mutex
	entries map[string]*Entry
}

// New returns an initialised fingerprint Store.
func New() *Store {
	return &Store{entries: make(map[string]*Entry)}
}

// Compute returns a deterministic hex fingerprint for a (job, reason) pair.
func Compute(job, reason string) string {
	h := sha256.New()
	h.Write([]byte(job + "\x00" + reason))
	return fmt.Sprintf("%x", h.Sum(nil))[:16]
}

// IsDuplicate returns true if the fingerprint has been seen before.
// It always records the event, updating LastSeen and Count.
func (s *Store) IsDuplicate(job, reason string, now time.Time) bool {
	fp := Compute(job, reason)
	s.mu.Lock()
	defer s.mu.Unlock()

	if e, ok := s.entries[fp]; ok {
		e.LastSeen = now
		e.Count++
		return true
	}

	s.entries[fp] = &Entry{
		Fingerprint: fp,
		JobName:     job,
		Reason:      reason,
		FirstSeen:   now,
		LastSeen:    now,
		Count:       1,
	}
	return false
}

// Get returns the Entry for a fingerprint, or false if not found.
func (s *Store) Get(fp string) (Entry, bool) {
	s.mu.Lock()
	defer s.mu.Unlock()
	e, ok := s.entries[fp]
	if !ok {
		return Entry{}, false
	}
	return *e, true
}

// Remove deletes a fingerprint entry, allowing the next occurrence to be treated as new.
func (s *Store) Remove(fp string) {
	s.mu.Lock()
	defer s.mu.Unlock()
	delete(s.entries, fp)
}

// All returns a snapshot of all current entries.
func (s *Store) All() []Entry {
	s.mu.Lock()
	defer s.mu.Unlock()
	out := make([]Entry, 0, len(s.entries))
	for _, e := range s.entries {
		out = append(out, *e)
	}
	return out
}
