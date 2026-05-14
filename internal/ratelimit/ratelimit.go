// Package ratelimit provides per-job alert rate limiting to prevent
// notification storms when a job repeatedly fails or is missed.
package ratelimit

import (
	"sync"
	"time"
)

// Limiter tracks the last alert time per job and suppresses alerts
// that occur within the configured cooldown window.
type Limiter struct {
	mu       sync.Mutex
	cooldown time.Duration
	last     map[string]time.Time
}

// New creates a Limiter with the given cooldown duration.
// Alerts for the same job within the cooldown window are suppressed.
func New(cooldown time.Duration) *Limiter {
	return &Limiter{
		cooldown: cooldown,
		last:     make(map[string]time.Time),
	}
}

// Allow reports whether an alert for the given job should be sent.
// It returns true and records the current time if the cooldown has
// elapsed since the last alert (or no alert has been sent yet).
func (l *Limiter) Allow(job string) bool {
	l.mu.Lock()
	defer l.mu.Unlock()

	now := time.Now()
	if t, ok := l.last[job]; ok && now.Sub(t) < l.cooldown {
		return false
	}
	l.last[job] = now
	return true
}

// Reset clears the rate-limit state for a specific job, allowing the
// next alert to be sent immediately regardless of the cooldown.
func (l *Limiter) Reset(job string) {
	l.mu.Lock()
	defer l.mu.Unlock()
	delete(l.last, job)
}

// LastAlert returns the time of the most recent allowed alert for the
// given job and whether an entry exists.
func (l *Limiter) LastAlert(job string) (time.Time, bool) {
	l.mu.Lock()
	defer l.mu.Unlock()
	t, ok := l.last[job]
	return t, ok
}

// ResetAll clears the rate-limit state for all jobs, allowing the next
// alert for every job to be sent immediately regardless of the cooldown.
func (l *Limiter) ResetAll() {
	l.mu.Lock()
	defer l.mu.Unlock()
	l.last = make(map[string]time.Time)
}
