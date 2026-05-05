// Package watcher implements the background monitoring loop for croncheck.
//
// A Watcher is constructed with a Scheduler, a Notifier, the list of
// configured jobs, and a check interval.  Calling Run blocks until the
// supplied context is cancelled, periodically invoking Scheduler.CheckMissed
// and forwarding any resulting alerts to the Notifier.
//
// Typical usage:
//
//	w := watcher.New(sched, notif, cfg.Jobs, 30*time.Second)
//	w.Run(ctx)
package watcher
