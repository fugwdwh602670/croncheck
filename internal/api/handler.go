// Package api provides an HTTP handler for the croncheck status endpoint.
package api

import (
	"encoding/json"
	"net/http"
	"time"

	"github.com/croncheck/internal/store"
)

// JobStatus represents the JSON-serialisable status of a single job.
type JobStatus struct {
	Name         string    `json:"name"`
	LastSeen     time.Time `json:"last_seen,omitempty"`
	MissedCount  int       `json:"missed_count"`
	State        string    `json:"state"`
}

// Handler returns an http.Handler that exposes job statuses as JSON.
func Handler(s *store.Store) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodGet {
			http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
			return
		}

		all := s.All()
		statuses := make([]JobStatus, 0, len(all))
		for name, entry := range all {
			statuses = append(statuses, JobStatus{
				Name:        name,
				LastSeen:    entry.LastSeen,
				MissedCount: entry.MissedCount,
				State:       entry.State,
			})
		}

		w.Header().Set("Content-Type", "application/json")
		if err := json.NewEncoder(w).Encode(statuses); err != nil {
			http.Error(w, "failed to encode response", http.StatusInternalServerError)
		}
	})
}
