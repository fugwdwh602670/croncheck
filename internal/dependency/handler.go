package dependency

import (
	"encoding/json"
	"net/http"
	"strings"
)

type setRequest struct {
	Upstreams []string `json:"upstreams"`
}

// HTTPHandler returns an http.Handler that exposes the dependency store via
// a small REST-like API.
//
//	GET  /dependencies          — list all dependencies
//	GET  /dependencies?job=X    — list upstreams for job X
//	PUT  /dependencies?job=X    — set upstreams for job X  (body: {"upstreams":[...]})
//	DELETE /dependencies?job=X  — remove dependencies for job X
func HTTPHandler(s *Store) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		job := strings.TrimSpace(r.URL.Query().Get("job"))

		switch r.Method {
		case http.MethodGet:
			w.Header().Set("Content-Type", "application/json")
			if job == "" {
				_ = json.NewEncoder(w).Encode(s.All())
				return
			}
			ups := s.Get(job)
			if ups == nil {
				ups = []string{}
			}
			_ = json.NewEncoder(w).Encode(map[string][]string{"upstreams": ups})

		case http.MethodPut:
			if job == "" {
				http.Error(w, "missing job query parameter", http.StatusBadRequest)
				return
			}
			var req setRequest
			if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
				http.Error(w, "invalid JSON body", http.StatusBadRequest)
				return
			}
			if err := s.Set(job, req.Upstreams); err != nil {
				http.Error(w, err.Error(), http.StatusBadRequest)
				return
			}
			w.WriteHeader(http.StatusNoContent)

		case http.MethodDelete:
			if job == "" {
				http.Error(w, "missing job query parameter", http.StatusBadRequest)
				return
			}
			s.Remove(job)
			w.WriteHeader(http.StatusNoContent)

		default:
			http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
		}
	})
}
