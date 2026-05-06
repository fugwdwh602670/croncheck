// Package dashboard provides a lightweight status summary endpoint
// that aggregates job health across the store for quick operational overviews.
package dashboard

import (
	"encoding/json"
	"net/http"
	"time"

	"github.com/example/croncheck/internal/store"
)

// JobSummary holds the condensed health status for a single job.
type JobSummary struct {
	Name        string    `json:"name"`
	Healthy     bool      `json:"healthy"`
	MissedCount int       `json:"missed_count"`
	LastSeen    time.Time `json:"last_seen,omitempty"`
}

// Summary is the top-level response returned by the dashboard endpoint.
type Summary struct {
	Total    int          `json:"total"`
	Healthy  int          `json:"healthy"`
	Unhealthy int         `json:"unhealthy"`
	NeverSeen int         `json:"never_seen"`
	Jobs     []JobSummary `json:"jobs"`
	GeneratedAt time.Time `json:"generated_at"`
}

// Storer is the subset of store.Store used by the dashboard handler.
type Storer interface {
	All() []store.JobState
}

// HTTPHandler returns an http.HandlerFunc that renders a JSON summary
// of all known jobs and their current health state.
func HTTPHandler(s Storer) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodGet {
			http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
			return
		}

		states := s.All()
		summary := Summary{
			Total:       len(states),
			GeneratedAt: time.Now().UTC(),
			Jobs:        make([]JobSummary, 0, len(states)),
		}

		for _, st := range states {
			js := JobSummary{
				Name:        st.Name,
				Healthy:     st.Healthy,
				MissedCount: st.MissedCount,
				LastSeen:    st.LastSeen,
			}
			summary.Jobs = append(summary.Jobs, js)

			switch {
			case st.LastSeen.IsZero():
				summary.NeverSeen++
			case st.Healthy:
				summary.Healthy++
			default:
				summary.Unhealthy++
			}
		}

		w.Header().Set("Content-Type", "application/json")
		_ = json.NewEncoder(w).Encode(summary)
	}
}
