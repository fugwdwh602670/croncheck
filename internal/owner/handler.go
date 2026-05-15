package owner

import (
	"encoding/json"
	"net/http"
	"strings"
)

// HTTPHandler returns an http.HandlerFunc that exposes owner CRUD over HTTP.
//
//	GET    /owner?job=<name>   – fetch owner for a specific job
//	GET    /owner              – list all owners
//	PUT    /owner?job=<name>   – set owner (body: {"owner":"...","contact":"..."})
//	DELETE /owner?job=<name>   – remove owner
func HTTPHandler(s *Store) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		switch r.Method {
		case http.MethodGet:
			handleGet(s, w, r)
		case http.MethodPut:
			handlePut(s, w, r)
		case http.MethodDelete:
			handleDelete(s, w, r)
		default:
			http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
		}
	}
}

func handleGet(s *Store, w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	job := strings.TrimSpace(r.URL.Query().Get("job"))
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

func handlePut(s *Store, w http.ResponseWriter, r *http.Request) {
	job := strings.TrimSpace(r.URL.Query().Get("job"))
	if job == "" {
		http.Error(w, "missing job parameter", http.StatusBadRequest)
		return
	}
	var body struct {
		Owner   string `json:"owner"`
		Contact string `json:"contact"`
	}
	if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
		http.Error(w, "invalid JSON", http.StatusBadRequest)
		return
	}
	if err := s.Set(job, body.Owner, body.Contact); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}
	w.WriteHeader(http.StatusNoContent)
}

func handleDelete(s *Store, w http.ResponseWriter, r *http.Request) {
	job := strings.TrimSpace(r.URL.Query().Get("job"))
	if job == "" {
		http.Error(w, "missing job parameter", http.StatusBadRequest)
		return
	}
	s.Remove(job)
	w.WriteHeader(http.StatusNoContent)
}
