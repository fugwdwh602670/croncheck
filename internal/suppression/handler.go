package suppression

import (
	"encoding/json"
	"net/http"
	"strconv"
)

// HTTPHandler returns an http.HandlerFunc that exposes suppression rules over
// a REST-style API.
//
//   GET  /suppression          — list all rules
//   GET  /suppression?job=X    — get rule for job
//   PUT  /suppression?job=X&min_consec_misses=N — set rule
//   DELETE /suppression?job=X  — remove rule
func HTTPHandler(s *Store) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		switch r.Method {
		case http.MethodGet:
			job := r.URL.Query().Get("job")
			if job == "" {
				_ = json.NewEncoder(w).Encode(s.All())
				return
			}
			rule, ok := s.Get(job)
			if !ok {
				http.Error(w, `{"error":"not found"}`, http.StatusNotFound)
				return
			}
			_ = json.NewEncoder(w).Encode(rule)
		case http.MethodPut:
			job := r.URL.Query().Get("job")
			raw := r.URL.Query().Get("min_consec_misses")
			n, err := strconv.Atoi(raw)
			if err != nil {
				http.Error(w, `{"error":"invalid min_consec_misses"}`, http.StatusBadRequest)
				return
			}
			if err := s.Set(job, n); err != nil {
				http.Error(w, `{"error":"`+err.Error()+`"}`, http.StatusBadRequest)
				return
			}
			w.WriteHeader(http.StatusNoContent)
		case http.MethodDelete:
			job := r.URL.Query().Get("job")
			if job == "" {
				http.Error(w, `{"error":"job required"}`, http.StatusBadRequest)
				return
			}
			s.Remove(job)
			w.WriteHeader(http.StatusNoContent)
		default:
			http.Error(w, `{"error":"method not allowed"}`, http.StatusMethodNotAllowed)
		}
	}
}
