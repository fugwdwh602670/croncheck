// Package retention manages automatic cleanup of stale job state.
// Jobs that have not sent a heartbeat within a configurable TTL are
// removed from the in-memory store to prevent unbounded growth.
package retention

import (
	"context"
	"log"
	"time"
)

// Pruner defines the interface required to remove stale jobs.
type Pruner interface {
	PruneStale(olderThan time.Duration) []string
}

// Cleaner periodically removes jobs that have not been seen recently.
type Cleaner struct {
	pruner   Pruner
	ttl      time.Duration
	interval time.Duration
}

// New creates a Cleaner that removes jobs unseen for ttl, checking every interval.
func New(pruner Pruner, ttl, interval time.Duration) *Cleaner {
	return &Cleaner{
		pruner:   pruner,
		ttl:      ttl,
		interval: interval,
	}
}

// Run starts the cleanup loop and blocks until ctx is cancelled.
func (c *Cleaner) Run(ctx context.Context) {
	ticker := time.NewTicker(c.interval)
	defer ticker.Stop()

	for {
		select {
		case <-ctx.Done():
			return
		case <-ticker.C:
			removed := c.pruner.PruneStale(c.ttl)
			if len(removed) > 0 {
				log.Printf("retention: pruned %d stale job(s): %v", len(removed), removed)
			}
		}
	}
}
