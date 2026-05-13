package deadletter

import (
	"encoding/json"
	"net/http"
)

// HTTPHandler exposes the dead-letter store over HTTP.
//
//   GET  /deadletter          – list all entries
//   DELETE /deadletter?job=X  – remove all entries for job X
func HTTPHandler(s *Store) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		switch r.Method {
		case http.MethodGet:
			handleList(w, s)
		case http.MethodDelete:
			handleRemove(w, r, s)
		default:
			http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
		}
	}
}

func handleList(w http.ResponseWriter, s *Store) {
	entries := s.All()
	if entries == nil {
		entries = []Entry{}
	}
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(entries) //nolint:errcheck
}

func handleRemove(w http.ResponseWriter, r *http.Request, s *Store) {
	job := r.URL.Query().Get("job")
	if job == "" {
		http.Error(w, "missing job query parameter", http.StatusBadRequest)
		return
	}
	s.Remove(job)
	w.WriteHeader(http.StatusNoContent)
}
