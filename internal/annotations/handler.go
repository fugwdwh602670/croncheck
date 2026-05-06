package annotations

import (
	"encoding/json"
	"net/http"
	"strings"
)

// HTTPHandler returns an http.Handler that exposes annotation CRUD over HTTP.
//
//	GET    /annotations?job=<name>  – retrieve annotations for a job
//	PUT    /annotations?job=<name>  – replace annotations for a job (JSON body)
//	DELETE /annotations?job=<name>  – remove all annotations for a job
func HTTPHandler(s *Store) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		job := strings.TrimSpace(r.URL.Query().Get("job"))

		switch r.Method {
		case http.MethodGet:
			if job == "" {
				w.Header().Set("Content-Type", "application/json")
				_ = json.NewEncoder(w).Encode(s.All())
				return
			}
			a, ok := s.Get(job)
			if !ok {
				http.Error(w, "job not found", http.StatusNotFound)
				return
			}
			w.Header().Set("Content-Type", "application/json")
			_ = json.NewEncoder(w).Encode(a)

		case http.MethodPut:
			if job == "" {
				http.Error(w, "job query parameter required", http.StatusBadRequest)
				return
			}
			var payload map[string]string
			if err := json.NewDecoder(r.Body).Decode(&payload); err != nil {
				http.Error(w, "invalid JSON body", http.StatusBadRequest)
				return
			}
			s.Set(job, payload)
			w.WriteHeader(http.StatusNoContent)

		case http.MethodDelete:
			if job == "" {
				http.Error(w, "job query parameter required", http.StatusBadRequest)
				return
			}
			s.Remove(job)
			w.WriteHeader(http.StatusNoContent)

		default:
			http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
		}
	})
}
