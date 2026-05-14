package notify_policy

import (
	"encoding/json"
	"net/http"
	"strings"
)

// HTTPHandler exposes the notify-policy store over HTTP.
//
//   GET    /notify-policy?job=<name>  — retrieve policy for a job
//   GET    /notify-policy             — list all policies
//   PUT    /notify-policy?job=<name>  — create or replace a policy
//   DELETE /notify-policy?job=<name>  — remove a policy
func HTTPHandler(s *Store) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		job := strings.TrimSpace(r.URL.Query().Get("job"))
		switch r.Method {
		case http.MethodGet:
			if job == "" {
				writeJSON(w, s.All())
				return
			}
			p, ok := s.Get(job)
			if !ok {
				http.Error(w, "not found", http.StatusNotFound)
				return
			}
			writeJSON(w, p)

		case http.MethodPut:
			if job == "" {
				http.Error(w, "missing job query param", http.StatusBadRequest)
				return
			}
			var p Policy
			if err := json.NewDecoder(r.Body).Decode(&p); err != nil {
				http.Error(w, "invalid JSON", http.StatusBadRequest)
				return
			}
			if err := s.Set(job, p); err != nil {
				http.Error(w, err.Error(), http.StatusBadRequest)
				return
			}
			w.WriteHeader(http.StatusNoContent)

		case http.MethodDelete:
			if job == "" {
				http.Error(w, "missing job query param", http.StatusBadRequest)
				return
			}
			s.Remove(job)
			w.WriteHeader(http.StatusNoContent)

		default:
			http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
		}
	}
}

func writeJSON(w http.ResponseWriter, v any) {
	w.Header().Set("Content-Type", "application/json")
	_ = json.NewEncoder(w).Encode(v)
}
