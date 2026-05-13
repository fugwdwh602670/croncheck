package replay

import (
	"encoding/json"
	"net/http"
)

// HTTPHandler returns an http.Handler for the replay API.
//
// POST /replay?job=<name>   — create a replay request
// PUT  /replay?job=<name>   — acknowledge a replay request
// DELETE /replay?job=<name> — remove a replay entry
// GET  /replay              — list all entries; optionally ?job=<name> for one
func HTTPHandler(s *Store) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		job := r.URL.Query().Get("job")
		switch r.Method {
		case http.MethodGet:
			handleGet(w, r, s, job)
		case http.MethodPost:
			handlePost(w, s, job)
		case http.MethodPut:
			handlePut(w, s, job)
		case http.MethodDelete:
			handleDelete(w, s, job)
		default:
			http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
		}
	})
}

func handleGet(w http.ResponseWriter, _ *http.Request, s *Store, job string) {
	w.Header().Set("Content-Type", "application/json")
	if job != "" {
		e, ok := s.Get(job)
		if !ok {
			http.Error(w, "not found", http.StatusNotFound)
			return
		}
		_ = json.NewEncoder(w).Encode(e)
		return
	}
	_ = json.NewEncoder(w).Encode(s.All())
}

func handlePost(w http.ResponseWriter, s *Store, job string) {
	if job == "" {
		http.Error(w, "missing job parameter", http.StatusBadRequest)
		return
	}
	if err := s.Request(job); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}
	w.WriteHeader(http.StatusCreated)
}

func handlePut(w http.ResponseWriter, s *Store, job string) {
	if job == "" {
		http.Error(w, "missing job parameter", http.StatusBadRequest)
		return
	}
	if err := s.Ack(job); err != nil {
		http.Error(w, err.Error(), http.StatusNotFound)
		return
	}
	w.WriteHeader(http.StatusNoContent)
}

func handleDelete(w http.ResponseWriter, s *Store, job string) {
	if job == "" {
		http.Error(w, "missing job parameter", http.StatusBadRequest)
		return
	}
	s.Remove(job)
	w.WriteHeader(http.StatusNoContent)
}
