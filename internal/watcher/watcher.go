// Package watcher ties together the scheduler, store, and notifier into a
// periodic background loop that checks for missed cron jobs and fires alerts.
package watcher

import (
	"context"
	"log"
	"time"

	"github.com/example/croncheck/internal/config"
	"github.com/example/croncheck/internal/notifier"
	"github.com/example/croncheck/internal/scheduler"
)

// Watcher runs the missed-job check loop.
type Watcher struct {
	sched    *scheduler.Scheduler
	notif    *notifier.Notifier
	interval time.Duration
	jobs     []config.Job
}

// New creates a Watcher that checks every interval.
func New(s *scheduler.Scheduler, n *notifier.Notifier, jobs []config.Job, interval time.Duration) *Watcher {
	return &Watcher{
		sched:    s,
		notif:    n,
		interval: interval,
		jobs:     jobs,
	}
}

// Run blocks until ctx is cancelled, checking for missed jobs on each tick.
func (w *Watcher) Run(ctx context.Context) {
	ticker := time.NewTicker(w.interval)
	defer ticker.Stop()

	log.Printf("watcher: starting — check interval %s", w.interval)

	for {
		select {
		case <-ctx.Done():
			log.Println("watcher: shutting down")
			return
		case t := <-ticker.C:
			w.check(t)
		}
	}
}

func (w *Watcher) check(now time.Time) {
	missed := w.sched.CheckMissed(now)
	for _, alert := range missed {
		log.Printf("watcher: missed job detected — %s (missed %d time(s))", alert.JobName, alert.MissedCount)
		if err := w.notif.Send(alert); err != nil {
			log.Printf("watcher: failed to send alert for %s: %v", alert.JobName, err)
		}
	}
}
