package escalation

import (
	"encoding/json"
	"net/http"
)

// HTTPHandler returns an http.Handler that exposes escalation state.
// GET /escalation?job=<name> returns the escalation entry for a single job.
// GET /escalation returns all escalation entries.
// DELETE /escalation?job=<name> resets the escalation state for a job.
func HTTPHandler(s *Store) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		switch r.Method {
		case http.MethodGet:
			handleGet(s, w, r)
		case http.MethodDelete:
			handleDelete(s, w, r)
		default:
			http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
		}
	})
}

func handleGet(s *Store, w http.ResponseWriter, r *http.Request) {
	job := r.URL.Query().Get("job")
	w.Header().Set("Content-Type", "application/json")

	if job == "" {
		all := s.All()
		if err := json.NewEncoder(w).Encode(all); err != nil {
			http.Error(w, "encode error", http.StatusInternalServerError)
		}
		return
	}

	entry, ok := s.Get(job)
	if !ok {
		http.Error(w, "job not found", http.StatusNotFound)
		return
	}
	if err := json.NewEncoder(w).Encode(entry); err != nil {
		http.Error(w, "encode error", http.StatusInternalServerError)
	}
}

func handleDelete(s *Store, w http.ResponseWriter, r *http.Request) {
	job := r.URL.Query().Get("job")
	if job == "" {
		http.Error(w, "missing job query parameter", http.StatusBadRequest)
		return
	}
	s.Reset(job)
	w.WriteHeader(http.StatusNoContent)
}
