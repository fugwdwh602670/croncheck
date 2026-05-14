// Package watcher implements the background monitoring loop for croncheck.
//
// A Watcher is constructed with a Scheduler, a Notifier, the list of
// configured jobs, and a check interval.  Calling Run blocks until the
// supplied context is cancelled, periodically invoking Scheduler.CheckMissed
// and forwarding any resulting alerts to the Notifier.
//
// # Lifecycle
//
// Create a Watcher with New, then call Run with a cancellable context.  When
// the context is cancelled Run drains any in-flight checks and returns.  It
// is safe to call Run exactly once per Watcher instance.
//
// # Error handling
//
// Transient errors from the Notifier are logged and do not stop the loop.
// A persistent failure to reach the Notifier will be retried on the next
// tick rather than causing Run to exit, so the process remains alive to
// attempt recovery.
//
// Typical usage:
//
//	w := watcher.New(sched, notif, cfg.Jobs, 30*time.Second)
//	w.Run(ctx)
package watcher
