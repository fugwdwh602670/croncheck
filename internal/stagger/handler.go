package stagger

import (
	"encoding/json"
	"net/http"
	"time"
)

type setRequest struct {
	Job   string `json:"job"`
	Delay string `json:"delay"`
}

// HTTPHandler returns an http.Handler for the stagger API.
//
// GET  /stagger          — list all stagger entries
// PUT  /stagger          — set a stagger delay (body: {"job":"...","delay":"5s"})
// DELETE /stagger?job=x  — remove a stagger entry
func HTTPHandler(s *Store) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		switch r.Method {
		case http.MethodGet:
			all := s.All()
			w.Header().Set("Content-Type", "application/json")
			_ = json.NewEncoder(w).Encode(all)

		case http.MethodPut:
			var req setRequest
			if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
				http.Error(w, "invalid JSON", http.StatusBadRequest)
				return
			}
			d, err := time.ParseDuration(req.Delay)
			if err != nil {
				http.Error(w, "invalid duration: "+err.Error(), http.StatusBadRequest)
				return
			}
			if err := s.Set(req.Job, d); err != nil {
				http.Error(w, err.Error(), http.StatusBadRequest)
				return
			}
			w.WriteHeader(http.StatusNoContent)

		case http.MethodDelete:
			job := r.URL.Query().Get("job")
			if job == "" {
				http.Error(w, "missing job parameter", http.StatusBadRequest)
				return
			}
			s.Remove(job)
			w.WriteHeader(http.StatusNoContent)

		default:
			http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
		}
	})
}
