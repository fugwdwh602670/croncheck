package ratelimit

import (
	"encoding/json"
	"net/http"
	"time"
)

type statusEntry struct {
	Job       string     `json:"job"`
	LastAlert *time.Time `json:"last_alert,omitempty"`
	Cooldown  string     `json:"cooldown"`
}

// HTTPHandler returns an http.Handler that exposes the current rate-limit
// state for all jobs as a JSON array. Only GET is supported.
func HTTPHandler(l *Limiter, jobs []string) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodGet {
			http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
			return
		}

		entries := make([]statusEntry, 0, len(jobs))
		for _, job := range jobs {
			e := statusEntry{
				Job:      job,
				Cooldown: l.cooldown.String(),
			}
			if t, ok := l.LastAlert(job); ok {
				e.LastAlert = &t
			}
			entries = append(entries, e)
		}

		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(entries) //nolint:errcheck
	})
}
