package trend

import (
	"encoding/json"
	"net/http"
)

// HTTPHandler returns an HTTP handler exposing trend data.
//
// GET /trend?job=<name>  — returns trend for a single job.
// GET /trend             — returns trends for all jobs.
func HTTPHandler(s *Store) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodGet {
			http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
			return
		}
		w.Header().Set("Content-Type", "application/json")

		job := r.URL.Query().Get("job")
		if job != "" {
			e, ok := s.Get(job)
			if !ok {
				http.Error(w, "job not found", http.StatusNotFound)
				return
			}
			_ = json.NewEncoder(w).Encode(e)
			return
		}

		_ = json.NewEncoder(w).Encode(s.All())
	}
}
