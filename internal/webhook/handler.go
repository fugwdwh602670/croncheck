package webhook

import (
	"encoding/json"
	"net/http"
	"strings"
)

// HTTPHandler returns an http.Handler that exposes webhook CRUD over HTTP.
//
//	GET    /webhooks          – list all
//	GET    /webhooks?job=name – get one
//	PUT    /webhooks          – register / update  (JSON body)
//	DELETE /webhooks?job=name – remove
func HTTPHandler(s *Store) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
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
	})
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
	var req struct {
		Job    string `json:"job"`
		URL    string `json:"url"`
		Secret string `json:"secret"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "invalid JSON", http.StatusBadRequest)
		return
	}
	if err := s.Set(req.Job, req.URL, req.Secret); err != nil {
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
