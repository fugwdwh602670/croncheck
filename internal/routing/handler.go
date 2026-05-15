package routing

import (
	"encoding/json"
	"net/http"
)

// HTTPHandler returns an http.HandlerFunc for managing routing rules.
//
// GET  /routing?job=<name>  — retrieve rule for a job (or all if omitted)
// PUT  /routing             — body: {"job":"...","channel":"..."}
// DELETE /routing?job=<name> — remove rule
func HTTPHandler(s *Store) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		switch r.Method {
		case http.MethodGet:
			handleGet(w, r, s)
		case http.MethodPut:
			handlePut(w, r, s)
		case http.MethodDelete:
			handleDelete(w, r, s)
		default:
			http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
		}
	}
}

func handleGet(w http.ResponseWriter, r *http.Request, s *Store) {
	w.Header().Set("Content-Type", "application/json")
	job := r.URL.Query().Get("job")
	if job == "" {
		json.NewEncoder(w).Encode(s.All())
		return
	}
	ch, ok := s.Get(job)
	if !ok {
		http.Error(w, "not found", http.StatusNotFound)
		return
	}
	json.NewEncoder(w).Encode(Rule{Job: job, Channel: ch})
}

func handlePut(w http.ResponseWriter, r *http.Request, s *Store) {
	var rule Rule
	if err := json.NewDecoder(r.Body).Decode(&rule); err != nil {
		http.Error(w, "invalid JSON", http.StatusBadRequest)
		return
	}
	if err := s.Set(rule.Job, rule.Channel); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}
	w.WriteHeader(http.StatusNoContent)
}

func handleDelete(w http.ResponseWriter, r *http.Request, s *Store) {
	job := r.URL.Query().Get("job")
	if job == "" {
		http.Error(w, "job parameter required", http.StatusBadRequest)
		return
	}
	s.Remove(job)
	w.WriteHeader(http.StatusNoContent)
}
