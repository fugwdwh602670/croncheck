// Package metrics exposes a simple HTTP handler that serves Prometheus-style
// plain-text metrics for croncheck job state.
package metrics

import (
	"fmt"
	"net/http"
	"time"

	"github.com/example/croncheck/internal/store"
)

// Provider supplies job state snapshots to the metrics handler.
type Provider interface {
	All() []store.JobState
}

// Handler returns an http.HandlerFunc that writes Prometheus-compatible
// plain-text metrics for all tracked jobs.
func Handler(p Provider) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "text/plain; version=0.0.4")

		jobs := p.All()
		now := time.Now()

		for _, j := range jobs {
			name := sanitize(j.Name)

			// 1 = healthy (received heartbeat), 0 = no heartbeat yet
			healthy := 0
			if !j.LastSeen.IsZero() {
				healthy = 1
			}
			fmt.Fprintf(w, "croncheck_job_healthy{job=%q} %d\n", name, healthy)

			// seconds since last heartbeat (-1 when never seen)
			sinceLastSeen := -1.0
			if !j.LastSeen.IsZero() {
				sinceLastSeen = now.Sub(j.LastSeen).Seconds()
			}
			fmt.Fprintf(w, "croncheck_job_last_seen_seconds{job=%q} %.3f\n", name, sinceLastSeen)

			// total consecutive missed runs
			fmt.Fprintf(w, "croncheck_job_missed_total{job=%q} %d\n", name, j.MissedCount)
		}
	}
}

// sanitize replaces characters that are invalid in Prometheus label values
// with underscores so metric output remains parseable.
func sanitize(s string) string {
	b := []byte(s)
	for i, c := range b {
		if !isLabelSafe(c) {
			b[i] = '_'
		}
	}
	return string(b)
}

func isLabelSafe(c byte) bool {
	return (c >= 'a' && c <= 'z') ||
		(c >= 'A' && c <= 'Z') ||
		(c >= '0' && c <= '9') ||
		c == '_' || c == '-' || c == '.'
}
