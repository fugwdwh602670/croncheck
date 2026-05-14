package forecast

import (
	"encoding/json"
	"net/http"
)

// HTTPHandler serves forecast data over HTTP.
//
// GET /forecast?job=<name>  — returns forecast for a single job.
// GET /forecast             — returns all forecasts.
func HTTPHandler(s *Store) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodGet {
			http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
			return
		}

		w.Header().Set("Content-Type", "application/json")

		if job := r.URL.Query().Get("job"); job != "" {
			e, ok := s.Get(job)
			if !ok {
				http.Error(w, "job not found", http.StatusNotFound)
				return
			}
			_ = json.NewEncoder(w).Encode(e)
			return
		}

		all := s.All()
		if all == nil {
			all = []Entry{}
		}
		_ = json.NewEncoder(w).Encode(all)
	}
}
