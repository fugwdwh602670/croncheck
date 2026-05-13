package quota

import (
	"encoding/json"
	"net/http"
	"time"
)

type quotaResponse struct {
	Job       string    `json:"job"`
	Count     int       `json:"count"`
	WindowEnd time.Time `json:"window_end"`
}

// HTTPHandler returns an http.Handler that exposes quota state.
//
// GET /quota          — list all jobs with active quota entries
// GET /quota?job=NAME — get quota entry for a specific job
// DELETE /quota?job=NAME — reset quota for a specific job
func HTTPHandler(s *Store) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		switch r.Method {
		case http.MethodGet:
			job := r.URL.Query().Get("job")
			if job != "" {
				e := s.Get(job)
				if e == nil {
					http.Error(w, "job not found", http.StatusNotFound)
					return
				}
				w.Header().Set("Content-Type", "application/json")
				json.NewEncoder(w).Encode(quotaResponse{
					Job:       job,
					Count:     e.Count,
					WindowEnd: e.WindowEnd,
				})
				return
			}
			all := s.All()
			results := make([]quotaResponse, 0, len(all))
			for name, e := range all {
				results = append(results, quotaResponse{
					Job:       name,
					Count:     e.Count,
					WindowEnd: e.WindowEnd,
				})
			}
			w.Header().Set("Content-Type", "application/json")
			json.NewEncoder(w).Encode(results)

		case http.MethodDelete:
			job := r.URL.Query().Get("job")
			if job == "" {
				http.Error(w, "missing job parameter", http.StatusBadRequest)
				return
			}
			s.Reset(job)
			w.WriteHeader(http.StatusNoContent)

		default:
			http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
		}
	})
}
