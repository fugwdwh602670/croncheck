package oncall

import (
	"encoding/json"
	"net/http"
	"strings"
	"time"
)

type setRequest struct {
	Contact  string `json:"contact"`
	Duration string `json:"duration"`
}

// HTTPHandler returns an http.Handler for the /oncall endpoint.
//
//	GET  /oncall          — list all active on-call entries
//	GET  /oncall?job=name — get entry for a specific job
//	PUT  /oncall?job=name — set on-call override (body: {contact, duration})
//	DELETE /oncall?job=name — remove override
func HTTPHandler(s *Store) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		job := strings.TrimSpace(r.URL.Query().Get("job"))
		w.Header().Set("Content-Type", "application/json")

		switch r.Method {
		case http.MethodGet:
			if job == "" {
				_ = json.NewEncoder(w).Encode(s.All())
				return
			}
			e, ok := s.Get(job)
			if !ok {
				http.Error(w, `{"error":"not found"}`, http.StatusNotFound)
				return
			}
			_ = json.NewEncoder(w).Encode(e)

		case http.MethodPut:
			if job == "" {
				http.Error(w, `{"error":"job query param required"}`, http.StatusBadRequest)
				return
			}
			var req setRequest
			if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
				http.Error(w, `{"error":"invalid JSON"}`, http.StatusBadRequest)
				return
			}
			d, err := time.ParseDuration(req.Duration)
			if err != nil || d <= 0 {
				http.Error(w, `{"error":"invalid duration"}`, http.StatusBadRequest)
				return
			}
			if err := s.Set(job, req.Contact, d); err != nil {
				http.Error(w, `{"error":"`+err.Error()+`"}`, http.StatusBadRequest)
				return
			}
			w.WriteHeader(http.StatusNoContent)

		case http.MethodDelete:
			if job == "" {
				http.Error(w, `{"error":"job query param required"}`, http.StatusBadRequest)
				return
			}
			s.Remove(job)
			w.WriteHeader(http.StatusNoContent)

		default:
			http.Error(w, `{"error":"method not allowed"}`, http.StatusMethodNotAllowed)
		}
	})
}
