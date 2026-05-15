package catalog

import (
	"encoding/json"
	"net/http"
	"strings"
)

// HTTPHandler returns an http.HandlerFunc for the catalog API.
// GET  /catalog          — list all entries
// GET  /catalog?job=X    — get single entry
// PUT  /catalog          — register/update entry (JSON body)
// DELETE /catalog?job=X  — remove entry
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
	job := strings.TrimSpace(r.URL.Query().Get("job"))
	w.Header().Set("Content-Type", "application/json")
	if job == "" {
		_ = json.NewEncoder(w).Encode(s.All())
		return
	}
	e, ok := s.Get(job)
	if !ok {
		http.Error(w, "not found", http.StatusNotFound)
		return
	}
	_ = json.NewEncoder(w).Encode(e)
}

func handlePut(w http.ResponseWriter, r *http.Request, s *Store) {
	var e Entry
	if err := json.NewDecoder(r.Body).Decode(&e); err != nil {
		http.Error(w, "invalid JSON", http.StatusBadRequest)
		return
	}
	if err := s.Set(e); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}
	w.WriteHeader(http.StatusNoContent)
}

func handleDelete(w http.ResponseWriter, r *http.Request, s *Store) {
	job := strings.TrimSpace(r.URL.Query().Get("job"))
	if job == "" {
		http.Error(w, "job parameter required", http.StatusBadRequest)
		return
	}
	s.Remove(job)
	w.WriteHeader(http.StatusNoContent)
}
