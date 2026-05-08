package sla

import (
	"encoding/json"
	"net/http"
)

// HTTPHandler returns an http.Handler for the SLA endpoint.
// GET /sla?job=<name> returns SLA stats for a specific job.
// GET /sla returns stats for all jobs.
func HTTPHandler(s *Store) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodGet {
			http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
			return
		}

		w.Header().Set("Content-Type", "application/json")

		jobName := r.URL.Query().Get("job")
		if jobName != "" {
			entry, ok := s.Get(jobName)
			if !ok {
				http.Error(w, "job not found", http.StatusNotFound)
				return
			}
			json.NewEncoder(w).Encode(entry)
			return
		}

		all := s.All()
		json.NewEncoder(w).Encode(all)
	})
}
